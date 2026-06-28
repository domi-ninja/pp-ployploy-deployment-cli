package main

import (
	"os"
	"path/filepath"

	"git.domi.ninja/domi-ninja/infra-meta-forgejo/internal/cli"
)

func main() {
	os.Exit(cli.Main(filepath.Base(os.Args[0]), os.Args[1:], os.Stdout, os.Stderr))
}
