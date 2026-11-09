//go:build wasip1

// Command plugin is the sync-wasm-sink example's sink: it posts each batch
// to a webhook, through fit's HTTP client, as NDJSON, a record's canonical
// JSON a line, with a bearer token from its config. A record without a key
// fails alone, those before it written; an answer other than 2xx fails the
// batch, which fit tries again as the pipeline's retry says. Build it from
// the repository's root, by Go or by TinyGo:
//
//	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o sync-wasm-sink/webhook.wasm ./sync-wasm-sink/plugin
//	tinygo build -target=wasip1 -buildmode=c-shared -o sync-wasm-sink/webhook.wasm ./sync-wasm-sink/plugin
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lesismal/fit/plugin/wasm/guest"
)

// settings is the config the pipeline gives the plugin.
type settings struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func init() { guest.OnWrite(write) }

func write(batch []guest.Record) error {
	var conf settings
	if err := json.Unmarshal([]byte(guest.Config()), &conf); err != nil {
		return fmt.Errorf("the config: %w", err)
	}
	// Those before a record without a key are posted; it goes as on_error
	// says.
	bad := len(batch)
	var body strings.Builder
	for i := range batch {
		if len(batch[i].Key) == 0 {
			bad = i
			break
		}
		b, _ := batch[i].MarshalJSON()
		body.Write(b)
		body.WriteByte('\n')
	}
	if err := post(conf, body.String(), bad); err != nil {
		return err
	}
	if bad < len(batch) {
		return &guest.RecordError{Index: bad, Msg: "a record without a key"}
	}
	return nil
}

// post sends n records' lines to the webhook.
func post(conf settings, body string, n int) error {
	if n == 0 {
		return nil
	}
	headers := map[string]string{"Content-Type": "application/x-ndjson"}
	if conf.Token != "" {
		headers["Authorization"] = "Bearer " + conf.Token
	}
	resp, err := guest.Do(guest.Request{Method: "POST", URL: conf.URL, Headers: headers, Body: body})
	if err != nil {
		return err
	}
	if resp.Status < 200 || resp.Status > 299 {
		return fmt.Errorf("the webhook answered %d: %.200s", resp.Status, resp.Body)
	}
	guest.Log(fmt.Sprintf("posted %d records: %d", n, resp.Status))
	return nil
}

func main() {}
