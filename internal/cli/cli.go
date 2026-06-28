package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"git.domi.ninja/domi-ninja/infra-meta-forgejo/internal/deploy"
)

func Main(name string, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		args = []string{"deploy"}
	}

	switch args[0] {
	case "deploy":
		fmt.Fprintf(stderr, "%s: full deployment is not implemented yet; run `%s plan` first\n", name, name)
		return 2
	case "plan":
		return runPlan(stdout, stderr)
	case "status":
		fmt.Fprintf(stderr, "%s status: not implemented yet\n", name)
		return 2
	case "rollback":
		fmt.Fprintf(stderr, "%s rollback: not implemented yet\n", name)
		return 2
	case "-h", "--help", "help":
		printHelp(name, stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printHelp(name, stderr)
		return 2
	}
}

func runPlan(stdout io.Writer, stderr io.Writer) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: read working directory: %v\n", err)
		return 1
	}

	plan, err := deploy.LoadPlan(wd, "deploy.yml")
	if err != nil {
		var validationErr deploy.ValidationError
		if errors.As(err, &validationErr) {
			fmt.Fprintf(stderr, "invalid deploy config:\n%s", validationErr.Error())
			return 1
		}

		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	deploy.PrintPlan(stdout, plan)
	return 0
}

func printHelp(name string, w io.Writer) {
	fmt.Fprintf(w, "usage: %s <command>\n", name)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  plan      validate deploy.yml and print host/service placement")
	fmt.Fprintln(w, "  status    show observed host state (not implemented)")
	fmt.Fprintln(w, "  rollback  restore previous code and DB state (not implemented)")
}
