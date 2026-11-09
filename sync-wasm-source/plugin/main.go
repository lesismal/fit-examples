//go:build wasip1

// Command plugin is the sync-wasm-source example's source: it reads an
// HTTP API of pages, each {"items": [...], "next": "<the next page>"}, the
// last without a next, through fit's HTTP client. Its checkpoint is the
// page it is on and how many of that page's items it has handed on, so a
// restart goes on where it was, and items added to the last page, or a
// next page added after it, are read when it polls again. Build it from
// the repository's root, by Go or by TinyGo:
//
//	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o sync-wasm-source/pages.wasm ./sync-wasm-source/plugin
//	tinygo build -target=wasip1 -buildmode=c-shared -o sync-wasm-source/pages.wasm ./sync-wasm-source/plugin
package main

import (
	"encoding/json"
	"fmt"

	"github.com/lesismal/fit/plugin/wasm/guest"
)

// settings is the config the pipeline gives the plugin: the API's base
// URL and its first page.
type settings struct {
	URL   string `json:"url"`
	First string `json:"first"`
}

// checkpoint is where the plugin is.
type checkpoint struct {
	Page string `json:"page"`
	Done int    `json:"done"`
}

type page struct {
	Items []struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		State string `json:"state"`
	} `json:"items"`
	Next string `json:"next"`
}

func init() { guest.OnPoll(poll) }

// poll reads the page the checkpoint names and hands on its items past
// those done. It moves to the next page when there is one, and has caught
// up when there is not, so that fit waits poll_interval before asking
// again.
func poll(cp []byte) ([]byte, bool, error) {
	var conf settings
	if err := json.Unmarshal([]byte(guest.Config()), &conf); err != nil {
		return nil, false, fmt.Errorf("the config: %w", err)
	}
	at := checkpoint{Page: conf.First}
	if len(cp) > 0 {
		if err := json.Unmarshal(cp, &at); err != nil {
			return nil, false, fmt.Errorf("the checkpoint: %w", err)
		}
	}
	resp, err := guest.Do(guest.Request{URL: conf.URL + at.Page})
	if err != nil {
		return nil, false, err
	}
	if resp.Status != 200 {
		return nil, false, fmt.Errorf("%s answered %d", at.Page, resp.Status)
	}
	var p page
	if err := json.Unmarshal(resp.Body, &p); err != nil {
		return nil, false, fmt.Errorf("%s: %w", at.Page, err)
	}
	for _, it := range p.Items[min(at.Done, len(p.Items)):] {
		guest.Emit(0, &guest.Record{Namespace: "api.tickets", Op: "insert",
			Key: guest.Row{{Name: "id", Value: guest.Int(it.ID)}}.Value(),
			After: guest.Row{
				{Name: "id", Value: guest.Int(it.ID)},
				{Name: "title", Value: guest.String(it.Title)},
				{Name: "state", Value: guest.String(it.State)},
			}.Value()})
	}
	next := checkpoint{Page: at.Page, Done: len(p.Items)}
	if p.Next != "" {
		next = checkpoint{Page: p.Next}
	}
	b, _ := json.Marshal(next)
	if next == at {
		return nil, true, nil // nothing new: the checkpoint stays
	}
	return b, p.Next == "", nil
}

func main() {}
