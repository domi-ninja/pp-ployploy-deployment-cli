package deploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitConfigCreatesStarterDeployYAML(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "My_App.Name")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}

	result, err := InitConfig(root, "deploy.yml")
	if err != nil {
		t.Fatal(err)
	}
	if result.ProjectName != "my-app-name" {
		t.Fatalf("unexpected project name %q", result.ProjectName)
	}

	body, err := os.ReadFile(filepath.Join(root, "deploy.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, wanted := range []string{
		"version: 1",
		"name: my-app-name",
		"ssh: deploy@example.com",
		"local_url: \"http://127.0.0.1:7468\"",
	} {
		if !strings.Contains(text, wanted) {
			t.Fatalf("starter config missing %q:\n%s", wanted, text)
		}
	}
}

func TestInitConfigRefusesOverwrite(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "deploy.yml", "version: 1\n")

	_, err := InitConfig(root, "deploy.yml")
	if err == nil {
		t.Fatal("expected overwrite error")
	}
	if !strings.Contains(err.Error(), "deploy.yml already exists") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"My App":       "my-app",
		"__API..Prod":  "api-prod",
		"---":          "",
		"quotes_2026!": "quotes-2026",
	}
	for input, want := range cases {
		if got := slugify(input); got != want {
			t.Fatalf("slugify(%q) = %q, want %q", input, got, want)
		}
	}
}
