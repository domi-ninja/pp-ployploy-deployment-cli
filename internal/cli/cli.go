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
		return runDeploy(stdout, stderr)
	case "down":
		return runDown(stdout, stderr)
	case "init":
		return runInit(stdout, stderr)
	case "plan":
		return runPlan(stdout, stderr)
	case "status":
		return runStatus(stdout, stderr)
	case "rollback":
		return runRollback(stdout, stderr)
	case "-h", "--help", "help":
		printHelp(name, stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printHelp(name, stderr)
		return 2
	}
}

func runDown(stdout io.Writer, stderr io.Writer) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: read working directory: %v\n", err)
		return 1
	}
	if err := deploy.NewDeployer(wd, stdout, stderr).Down(); err != nil {
		printError(stderr, err)
		return 1
	}
	return 0
}

func runDeploy(stdout io.Writer, stderr io.Writer) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: read working directory: %v\n", err)
		return 1
	}
	if err := deploy.NewDeployer(wd, stdout, stderr).Deploy(); err != nil {
		printError(stderr, err)
		return 1
	}
	return 0
}

func runInit(stdout io.Writer, stderr io.Writer) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: read working directory: %v\n", err)
		return 1
	}

	result, err := deploy.InitConfig(wd, "deploy.yml")
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "created %s\n", result.Path)
	fmt.Fprintf(stdout, "project: %s\n", result.ProjectName)
	return 0
}

func runStatus(stdout io.Writer, stderr io.Writer) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: read working directory: %v\n", err)
		return 1
	}
	if err := deploy.NewDeployer(wd, stdout, stderr).Status(); err != nil {
		printError(stderr, err)
		return 1
	}
	return 0
}

func runRollback(stdout io.Writer, stderr io.Writer) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: read working directory: %v\n", err)
		return 1
	}
	if err := deploy.NewDeployer(wd, stdout, stderr).Rollback(); err != nil {
		printError(stderr, err)
		return 1
	}
	return 0
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

func printError(stderr io.Writer, err error) {
	var validationErr deploy.ValidationError
	if errors.As(err, &validationErr) {
		fmt.Fprintf(stderr, "invalid deploy config:\n%s", validationErr.Error())
		return
	}
	fmt.Fprintf(stderr, "error: %v\n", err)
}

func printHelp(name string, w io.Writer) {
	fmt.Fprintf(w, "usage: %s <command>\n", name)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  deploy    build, transfer, and apply the release")
	fmt.Fprintln(w, "  down      remove all remote containers for the project")
	fmt.Fprintln(w, "  init      create a starter deploy.yml")
	fmt.Fprintln(w, "  plan      validate deploy.yml and print host/service placement")
	fmt.Fprintln(w, "  status    show local deployment state")
	fmt.Fprintln(w, "  rollback  restore previous code and DB state")
}
