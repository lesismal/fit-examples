// Command api is the video site's own API, behind fit (video-site/fit.conf):
// it logs a viewer in, lists the site's catalog, and hands the viewer
// playback URLs that expire, which it asks fit's media_tokens for. fit
// keeps none of the site's accounts, and the site keeps none of fit's keys:
// the API decides who watches what, and fit checks every playlist and
// segment against the URL it signed.
//
//	go run ./video-site/api
//
// Viewers log in as alice or bob, password example-password. The flags say
// where fit's media_tokens is and the password of its route's basic_auth,
// which the environment's FIT_TOKENS_PASSWORD may give instead.
//
// 视频网站自己的 API，位于 fit 之后（video-site/fit.conf）：它让观众登录、
// 列出网站的片目，并向观众发放会过期的播放地址，这些地址由它向 fit 的
// media_tokens 申请。fit 不保存网站的任何账号，网站也不持有 fit 的任何
// 密钥：API 决定谁可以看什么，fit 按它签发的地址检查每个播放列表和分片。
// 观众以 alice 或 bob 登录，密码 example-password。
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// A viewer of the site: who logs in, and the opaque ID playback URLs carry
// as their subject, which fit's access log shows and its max_sessions
// counts; never a name or an e-mail address.
type viewer struct {
	name string
	id   string
	hash []byte
}

// The site's viewers. A real site keeps them in its database; both
// passwords are example-password.
var viewers = map[string]*viewer{
	"alice": {name: "alice", id: "v-1001", hash: []byte("$2a$10$gkDJyGagxp1q7R24F0rSjulsgZOFL55Zm6JK2RcxTCLDk23szhZA2")},
	"bob":   {name: "bob", id: "v-1002", hash: []byte("$2a$10$gkDJyGagxp1q7R24F0rSjulsgZOFL55Zm6JK2RcxTCLDk23szhZA2")},
}

// A title of the catalog: a film of the media library films, played from
// fit's vod handler, or a channel of fit's streaming role, played from its
// hls handler.
type title struct {
	Kind  string `json:"kind"` // film or channel
	ID    string `json:"id"`
	Title string `json:"title"`
}

// path is the title's playlist on fit, which the URL is signed for.
func (t title) path() string {
	if t.Kind == "channel" {
		return "/hls/" + t.ID + "/index.m3u8"
	}
	return "/vod/films/" + t.ID + "/index.m3u8"
}

type server struct {
	catalog   []title
	tokensURL string // fit's media_tokens /urls
	tokensPW  string // its route's basic_auth password, for the user site-api
	ttl       time.Duration
	client    *http.Client

	mu       sync.Mutex
	sessions map[string]session // by the session cookie's value
}

type session struct {
	v       *viewer
	expires time.Time
}

const (
	sessionCookie = "site_session"
	sessionTTL    = 12 * time.Hour
	maxSessions   = 10000
	maxBody       = 4096
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8090", "the address to listen on, which fit's site-api upstream names")
	fit := flag.String("fit", "http://127.0.0.1:8080", "fit, whose /media-tokens/urls signs playback URLs")
	catalogFile := flag.String("catalog", "video-site/catalog.json", "the site's catalog")
	ttl := flag.Duration("ttl", 2*time.Minute, "how long a playback URL is good for; a player that started plays on, on the grant its playlist hands it")
	flag.Parse()
	pw := os.Getenv("FIT_TOKENS_PASSWORD")
	if pw == "" {
		pw = "example-password"
	}
	s, err := newServer(*catalogFile, *fit+"/media-tokens/urls", pw, *ttl)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Addr: *listen, Handler: s, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 2 * time.Minute, MaxHeaderBytes: 16 << 10}
	log.Printf("the video site's API on %s, signing through %s", *listen, s.tokensURL)
	log.Fatal(srv.ListenAndServe())
}

func newServer(catalogFile, tokensURL, tokensPW string, ttl time.Duration) (*server, error) {
	data, err := os.ReadFile(catalogFile)
	if err != nil {
		return nil, err
	}
	s := &server{tokensURL: tokensURL, tokensPW: tokensPW, ttl: ttl, sessions: map[string]session{},
		client: &http.Client{Timeout: 10 * time.Second}}
	if err := json.Unmarshal(data, &s.catalog); err != nil {
		return nil, fmt.Errorf("%s: %w", catalogFile, err)
	}
	return s, nil
}

// ServeHTTP serves the API under /api/, as fit's route api passes it on.
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	switch {
	case r.URL.Path == "/api/login" && r.Method == http.MethodPost:
		s.login(w, r)
	case r.URL.Path == "/api/logout" && r.Method == http.MethodPost:
		s.logout(w, r)
	case r.URL.Path == "/api/me" && r.Method == http.MethodGet:
		if v := s.viewer(r); v != nil {
			writeJSON(w, http.StatusOK, map[string]string{"name": v.name})
			return
		}
		http.Error(w, "not logged in", http.StatusUnauthorized)
	case r.URL.Path == "/api/catalog" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, s.catalog)
	case r.URL.Path == "/api/play" && r.Method == http.MethodPost:
		s.play(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ User, Password string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody)).Decode(&in); err != nil {
		http.Error(w, `the body is {"user", "password"}`, http.StatusBadRequest)
		return
	}
	v := viewers[in.User]
	hash := []byte("$2a$10$gkDJyGagxp1q7R24F0rSjulsgZOFL55Zm6JK2RcxTCLDk23szhZA2")
	if v != nil {
		hash = v.hash
	}
	// A user who does not exist costs a check all the same.
	if bcrypt.CompareHashAndPassword(hash, []byte(in.Password)) != nil || v == nil {
		http.Error(w, "the user or the password is wrong", http.StatusUnauthorized)
		return
	}
	id := randomID()
	s.mu.Lock()
	now := time.Now()
	for k, ss := range s.sessions {
		if now.After(ss.expires) {
			delete(s.sessions, k)
		}
	}
	if len(s.sessions) >= maxSessions {
		s.mu.Unlock()
		http.Error(w, "too many sessions; try again later", http.StatusServiceUnavailable)
		return
	}
	s.sessions[id] = session{v: v, expires: now.Add(sessionTTL)}
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: id, Path: "/api/", MaxAge: int(sessionTTL.Seconds()),
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: r.Header.Get("X-Forwarded-Proto") == "https"})
	writeJSON(w, http.StatusOK, map[string]string{"name": v.name})
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		s.mu.Lock()
		delete(s.sessions, c.Value)
		s.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Path: "/api/", MaxAge: -1, HttpOnly: true})
	writeJSON(w, http.StatusOK, map[string]string{})
}

// viewer returns the logged-in viewer of r, or nil.
func (s *server) viewer(r *http.Request) *viewer {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ss, ok := s.sessions[c.Value]
	if !ok || time.Now().After(ss.expires) {
		return nil
	}
	return ss.v
}

// play hands a logged-in viewer a playback URL of a title: signed by fit
// for the viewer's opaque ID, bound to the address it comes from, good for
// the API's ttl.
func (s *server) play(w http.ResponseWriter, r *http.Request) {
	v := s.viewer(r)
	if v == nil {
		http.Error(w, "log in first", http.StatusUnauthorized)
		return
	}
	var in struct{ Kind, ID string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody)).Decode(&in); err != nil {
		http.Error(w, `the body is {"kind", "id"}`, http.StatusBadRequest)
		return
	}
	var t *title
	for i := range s.catalog {
		if s.catalog[i].Kind == in.Kind && s.catalog[i].ID == in.ID {
			t = &s.catalog[i]
		}
	}
	if t == nil {
		http.Error(w, "no such title", http.StatusNotFound)
		return
	}
	u, exp, err := s.sign(v, clientAddr(r), t.path())
	if err != nil {
		log.Printf("signing %s for %s: %v", t.path(), v.id, err)
		http.Error(w, "the playback URL could not be had; try again", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"url": u, "expires": exp, "title": t.Title})
}

// sign asks fit's media_tokens for a URL of path for v at addr.
func (s *server) sign(v *viewer, addr, path string) (string, int64, error) {
	body, _ := json.Marshal(map[string]any{"subject": v.id, "address": addr, "ttl": s.ttl.String(), "paths": []string{path}})
	req, err := http.NewRequest(http.MethodPost, s.tokensURL, bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("site-api", s.tokensPW)
	resp, err := s.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("media_tokens: %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	var out struct {
		Expires int64    `json:"expires"`
		URLs    []string `json:"urls"`
	}
	if err := json.Unmarshal(data, &out); err != nil || len(out.URLs) != 1 {
		return "", 0, errors.New("media_tokens: an answer without one URL")
	}
	return out.URLs[0], out.Expires, nil
}

// clientAddr is the viewer's address: the last X-Forwarded-For fit's proxy
// added, as the API listens on the loopback behind it, or the peer's.
func clientAddr(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		xff := r.Header.Values("X-Forwarded-For")
		if len(xff) > 0 {
			list := strings.Split(xff[len(xff)-1], ",")
			if a := net.ParseIP(strings.TrimSpace(list[len(list)-1])); a != nil {
				return a.String()
			}
		}
	}
	return host
}

func randomID() string {
	var b [24]byte
	_, _ = rand.Read(b[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
