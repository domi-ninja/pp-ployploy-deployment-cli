package main

import (
	"os"

	"git.domi.ninja/domi-ninja/infra-meta-forgejo/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
