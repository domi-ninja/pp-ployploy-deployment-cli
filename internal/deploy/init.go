package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type InitResult struct {
	Path        string
	ProjectName string
}

func InitConfig(root string, configPath string) (InitResult, error) {
	fullPath := filepath.Join(root, configPath)
	if _, err := os.Stat(fullPath); err == nil {
		return InitResult{}, fmt.Errorf("%s already exists", configPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return InitResult{}, fmt.Errorf("check %s: %w", configPath, err)
	}

	projectName := slugify(filepath.Base(root))
	if projectName == "" {
		projectName = "app"
	}

	body := starterConfig(projectName)
	if err := os.WriteFile(fullPath, []byte(body), 0644); err != nil {
		return InitResult{}, fmt.Errorf("write %s: %w", configPath, err)
	}

	return InitResult{Path: fullPath, ProjectName: projectName}, nil
}

func starterConfig(projectName string) string {
	return fmt.Sprintf(`version: 1

project:
  name: %s
  environment: prod

build:
  context: .
  dockerfile: Dockerfile
  tags:
    - "%s:${git_sha}"

hosts:
  app-01:
    ssh: deploy@example.com
    agent:
      local_url: "http://127.0.0.1:7468"
    roles: [web]

services:
  web:
    image: "%s:${git_sha}"
    hosts: [app-01]
    ports:
      - published: 8080
        target: 3000
    health:
      command: ["true"]
      timeout_seconds: 60
`, projectName, projectName, projectName)
}

func slugify(value string) string {
	value = strings.ToLower(value)
	var out strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out.WriteRune(r)
			lastDash = false
		case unicode.IsSpace(r) || r == '-' || r == '_' || r == '.':
			if out.Len() > 0 && !lastDash {
				out.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(out.String(), "-")
}
