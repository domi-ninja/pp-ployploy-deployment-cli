package deploy

import (
	"strings"
	"testing"
)

func TestAutoPortScriptUsesStableKeyAndRange(t *testing.T) {
	script := autoPortScript("quotes:web:80")
	for _, wanted := range []string{
		"/etc/pp/ports.tsv",
		"/etc/pp/ports.lock",
		"quotes:web:80",
		"seq 18000 19999",
	} {
		if !strings.Contains(script, wanted) {
			t.Fatalf("script missing %q:\n%s", wanted, script)
		}
	}
}
