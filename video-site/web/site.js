// The video site's page: a viewer logs in to the site's API, picks a title
// and plays it with hls.js through the playback URL the API hands out. The
// URL expires in minutes; the player plays on, on the grant fit's playlists
// hand it, and once that is refused too (403), it asks for a new URL.
//
// 视频网站的页面：观众登录网站的 API，选一个节目，用 hls.js 播放 API 发放的
// 播放地址。地址几分钟后过期；播放器凭 fit 的播放列表交给它的授权继续播放，
// 当授权也被拒绝（403）时，再申请新的地址。
'use strict';

(function () {
  var $ = function (id) { return document.getElementById(id); };
  var state = { user: '', playing: null, url: '', expires: 0, refused: false, error: '', hls: null };

  async function call(method, path, body) {
    var opts = { method: method, credentials: 'same-origin', headers: {} };
    if (body !== undefined) {
      opts.headers['Content-Type'] = 'application/json';
      opts.body = JSON.stringify(body);
    }
    var res = await fetch(path, opts);
    var text = await res.text();
    var data = null;
    try { data = JSON.parse(text); } catch (e) { data = null; }
    return { status: res.status, data: data, text: text };
  }

  function show() {
    $('login').hidden = !!state.user;
    $('catalog').hidden = !state.user;
    $('logout').hidden = !state.user;
    $('who').textContent = state.user ? 'Logged in as ' + state.user : '';
  }

  function status(msg, error) {
    $('status').textContent = msg;
    $('status').className = 'hint' + (error ? ' error' : '');
  }

  async function loadCatalog() {
    var r = await call('GET', '/api/catalog');
    var ul = $('titles');
    ul.textContent = '';
    (r.data || []).forEach(function (t) {
      var li = document.createElement('li');
      var a = document.createElement('a');
      a.href = '#' + t.kind + '/' + t.id;
      a.textContent = t.title + (t.kind === 'channel' ? ' (live)' : '');
      a.setAttribute('data-kind', t.kind);
      a.setAttribute('data-id', t.id);
      a.addEventListener('click', function (ev) { ev.preventDefault(); play(t); });
      li.appendChild(a);
      ul.appendChild(li);
    });
  }

  function stop() {
    if (state.hls) state.hls.destroy();
    state.hls = null;
    var v = $('video');
    v.removeAttribute('src');
    v.load();
  }

  async function play(t) {
    stop();
    state.playing = t;
    state.refused = false;
    state.error = '';
    $('player').hidden = false;
    $('again').hidden = true;
    $('now').textContent = t.title;
    status('Getting a playback URL…');
    var r = await call('POST', '/api/play', { kind: t.kind, id: t.id });
    if (r.status !== 200) {
      state.error = r.text.trim();
      status('Cannot play: ' + state.error, true);
      if (r.status === 401) { state.user = ''; show(); }
      return;
    }
    state.url = r.data.url;
    state.expires = r.data.expires;
    status('The URL is good until ' + new Date(state.expires * 1000).toLocaleTimeString() + '; a player that started plays on after it.');
    var v = $('video');
    if (window.Hls && Hls.isSupported()) {
      var h = new Hls();
      state.hls = h;
      h.on(Hls.Events.ERROR, function (_, d) {
        var code = d.response && d.response.code;
        if (code === 403 || code === 401) {
          // The URL and the grant it handed on have expired, or its
          // session was ended: ask for a new one.
          state.refused = true;
          stop();
          status('Playback was refused: the URL has expired. Play again for a new one.', true);
          $('again').hidden = false;
        } else if (d.fatal) {
          state.error = d.details;
          stop();
          status('Playback failed: ' + d.details, true);
          $('again').hidden = false;
        }
      });
      h.attachMedia(v);
      h.loadSource(state.url);
      h.on(Hls.Events.MANIFEST_PARSED, function () { v.play().catch(function () {}); });
    } else if (v.canPlayType('application/vnd.apple.mpegurl')) {
      v.src = state.url; // Safari's own HLS
      v.play().catch(function () {});
    } else {
      status('This browser cannot play HLS.', true);
    }
  }

  $('login').addEventListener('submit', async function (ev) {
    ev.preventDefault();
    var r = await call('POST', '/api/login', { user: $('user').value, password: $('password').value });
    $('password').value = '';
    if (r.status !== 200) { alert(r.text); return; }
    state.user = r.data.name;
    show();
    loadCatalog();
  });
  $('logout').addEventListener('click', async function () {
    stop();
    await call('POST', '/api/logout', {});
    state.user = '';
    $('player').hidden = true;
    show();
  });
  $('again').addEventListener('click', function () { if (state.playing) play(state.playing); });

  // What a test reads.
  window.site = {
    state: function () {
      var v = $('video');
      return { user: state.user, playing: state.playing, url: state.url, expires: state.expires, refused: state.refused,
        error: state.error, currentTime: v.currentTime, duration: v.duration, paused: v.paused, seeking: v.seeking };
    },
    seek: function (to) { $('video').currentTime = to; },
  };

  (async function () {
    var r = await call('GET', '/api/me');
    if (r.status === 200) state.user = r.data.name;
    show();
    if (state.user) loadCatalog();
  })();
})();
