//go:build wasip1

// Command plugin is the sync-wasm-transform example's transform: it drops
// the records of test accounts, whose email is at the domain the config
// names, and gives every other record a domain field and its email with
// the local part hidden. Build it from the repository's root:
//
//	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o sync-wasm-transform/users.wasm ./sync-wasm-transform/plugin
package main

import (
	"encoding/json"
	"strings"

	"github.com/lesismal/fit/plugin/wasm/guest"
)

// settings is the config the pipeline gives the plugin.
type settings struct {
	DropDomain string `json:"drop_domain"`
}

var (
	conf   settings
	loaded bool
)

func cfg() settings {
	if !loaded {
		_ = json.Unmarshal([]byte(guest.Config()), &conf)
		loaded = true
	}
	return conf
}

func init() { guest.OnTransform(transform) }

// transform emits each record of the batch changed, or drops it by
// emitting nothing for it. A record whose row is not a map fails alone,
// and the pipeline's on_error takes it.
func transform(batch []guest.Record) error {
	drop := cfg().DropDomain
	for i := range batch {
		r := &batch[i]
		row, err := guest.ParseRow(r.After)
		if err != nil {
			guest.Fail(i, "the row is not a map: "+err.Error())
			continue
		}
		v, ok := row.Get("email")
		email, isString := guest.AsString(v)
		if !ok || !isString {
			guest.Emit(i, r)
			continue
		}
		local, domain, found := strings.Cut(email, "@")
		if found && drop != "" && strings.EqualFold(domain, drop) {
			continue
		}
		if found && len(local) > 1 {
			row.Set("email", guest.String(local[:1]+"***@"+domain))
		}
		row.Set("domain", guest.String(strings.ToLower(domain)))
		r.After = row.Value()
		guest.Emit(i, r)
	}
	return nil
}

func main() {}
