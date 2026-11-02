// Package examples holds fit's config examples, one directory each; its
// test checks them with fit validate, as a user would.
package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	_ "github.com/lesismal/fit/app/standard"
	"github.com/lesismal/fit/cli"
	_ "github.com/lesismal/fit/modules/standard"
	_ "github.com/lesismal/fit/provider/standard"
)

// noValidate names the examples that fit validate cannot check here,
// because they name files a user supplies, which validate opens.
var noValidate = map[string]string{
	"https": "its certificate and key files are the user's own",
}

// TestConfigs checks every example directory's fit.conf with fit validate,
// from the repository's root, which the examples name their files from.
// One that cannot be validated still has to adapt.
func TestConfigs(t *testing.T) {
	confs, _ := filepath.Glob("*/fit.conf")
	if len(confs) < 50 {
		t.Fatalf("only %d examples", len(confs))
	}
	for _, f := range confs {
		name := filepath.Dir(f)
		cmd := "validate"
		if _, ok := noValidate[name]; ok {
			cmd = "adapt"
		}
		var stdout, stderr strings.Builder
		code := cli.Main([]string{cmd, "-c", f}, &stdout, &stderr)
		if code != 0 || stderr.Len() > 0 || cmd == "validate" && stdout.String() != f+": valid\n" {
			t.Errorf("fit %s -c %s: %d %q %q", cmd, f, code, stdout.String(), stderr.String())
		}
	}
	for name := range noValidate {
		if _, err := os.Stat(filepath.Join(name, "fit.conf")); err != nil {
			t.Errorf("unvalidated example %s: %v", name, err)
		}
	}
}

// TestBilingualComments checks that every block of comment lines in a
// fit.conf, those in a row at one indentation, is in English and then in
// Chinese: it has a line with Chinese in it, after one without.
func TestBilingualComments(t *testing.T) {
	confs, _ := filepath.Glob("*/fit.conf")
	for _, f := range confs {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var block []string
		start, indent := 0, ""
		check := func() {
			english, chinese := false, false
			for _, l := range block {
				if hasHan(l) {
					chinese = chinese || english
				} else {
					english = true
				}
			}
			if len(block) > 0 && (!english || !chinese) {
				t.Errorf("%s:%d: a comment not in English and then Chinese: %q", f, start, block)
			}
			block = block[:0]
		}
		for i, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimLeft(line, " \t")
			lead := line[:len(line)-len(trimmed)]
			if !strings.HasPrefix(trimmed, "#") {
				check()
				continue
			}
			if len(block) > 0 && lead != indent {
				check()
			}
			if len(block) == 0 {
				start, indent = i+1, lead
			}
			block = append(block, trimmed)
		}
		check()
	}
}

func hasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// TestEveryDirectory checks that every directory is an example: it has a
// fit.conf, or is one of those with a README of their own, or the module.
func TestEveryDirectory(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || e.Name() == "module" {
			continue
		}
		_, conf := os.Stat(filepath.Join(e.Name(), "fit.conf"))
		_, readme := os.Stat(filepath.Join(e.Name(), "README.md"))
		if conf != nil && readme != nil {
			t.Errorf("%s has neither fit.conf nor README.md", e.Name())
		}
	}
}
