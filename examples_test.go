// Package examples holds fit's config examples, one directory each; its
// test checks them with fit validate, as a user would.
package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
