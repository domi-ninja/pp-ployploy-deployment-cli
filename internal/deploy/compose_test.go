package deploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderBundleWritesComposeAndEnv(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".env.prod", "DATABASE_URL=postgres://example\nSESSION_SECRET=secret\nQUEUE_URL=redis://example\n")
	cfg := validConfig()
	git := GitMetadata{SHA: "abcdef1234567890", ShortSHA: "abcdef1", Dirty: true, DiffDigest: "digest"}
	plan := BuildPlan(cfg, git, time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC))

	bundle, err := RenderBundle(root, plan)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.ImageTar != filepath.Join(root, ".deploy", "releases", "20260628T100000Z-abcdef1", "images", "quotes-20260628T100000Z-abcdef1.tar") {
		t.Fatalf("unexpected image tar %s", bundle.ImageTar)
	}
	if len(bundle.Hosts) != 2 {
		t.Fatalf("expected 2 host bundles, got %d", len(bundle.Hosts))
	}

	compose, err := os.ReadFile(filepath.Join(bundle.Root, "hosts", "app-01", "compose.yml"))
	if err != nil {
		t.Fatal(err)
	}
	composeText := string(compose)
	for _, wanted := range []string{
		"image: quotes:abcdef1234567890",
		"- 443:3000",
		"- env/web.env",
		"pp.release: 20260628T100000Z-abcdef1",
	} {
		if !strings.Contains(composeText, wanted) {
			t.Fatalf("compose missing %q:\n%s", wanted, composeText)
		}
	}

	env, err := os.ReadFile(filepath.Join(bundle.Root, "hosts", "app-01", "env", "web.env"))
	if err != nil {
		t.Fatal(err)
	}
	envText := string(env)
	if !strings.Contains(envText, "DATABASE_URL=postgres://example\n") || !strings.Contains(envText, "SESSION_SECRET=secret\n") {
		t.Fatalf("unexpected env file:\n%s", envText)
	}
}
