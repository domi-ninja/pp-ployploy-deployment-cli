package deploy

import (
	"fmt"
	"io"
	"sort"
	"time"
)

type Plan struct {
	Config    Config
	Git       GitMetadata
	ReleaseID string
	Hosts     []HostPlan
}

type HostPlan struct {
	ID       string
	SSH      string
	Services []string
}

func LoadPlan(root string, configPath string) (Plan, error) {
	cfg, err := LoadConfig(root, configPath)
	if err != nil {
		return Plan{}, err
	}
	git, err := ReadGitMetadata(root)
	if err != nil {
		return Plan{}, fmt.Errorf("read git metadata: %w", err)
	}

	return BuildPlan(cfg, git, time.Now()), nil
}

func BuildPlan(cfg Config, git GitMetadata, now time.Time) Plan {
	hostPlans := make([]HostPlan, 0, len(cfg.Hosts))
	for _, hostItem := range sortedMap(cfg.Hosts) {
		hostID := hostItem.Key
		host := hostItem.Value
		services := make([]string, 0)
		for _, serviceItem := range sortedMap(cfg.Services) {
			serviceID := serviceItem.Key
			service := serviceItem.Value
			if contains(service.Hosts, hostID) {
				services = append(services, serviceID)
			}
		}
		hostPlans = append(hostPlans, HostPlan{
			ID:       hostID,
			SSH:      host.SSH,
			Services: services,
		})
	}

	return Plan{
		Config:    cfg,
		Git:       git,
		ReleaseID: ReleaseID(now, git.ShortSHA),
		Hosts:     hostPlans,
	}
}

func PrintPlan(w io.Writer, plan Plan) {
	fmt.Fprintf(w, "project: %s\n", plan.Config.Project.Name)
	fmt.Fprintf(w, "environment: %s\n", plan.Config.Project.Environment)
	fmt.Fprintf(w, "release: %s\n", plan.ReleaseID)
	fmt.Fprintf(w, "git_sha: %s\n", plan.Git.SHA)
	fmt.Fprintf(w, "dirty: %t\n", plan.Git.Dirty)
	if plan.Git.Dirty {
		fmt.Fprintf(w, "worktree_diff_digest: %s\n", plan.Git.DiffDigest)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "hosts:")
	for _, host := range plan.Hosts {
		fmt.Fprintf(w, "  - %s (%s)\n", host.ID, host.SSH)
		if len(host.Services) == 0 {
			fmt.Fprintln(w, "    services: []")
			continue
		}
		fmt.Fprintf(w, "    services: %s\n", joinComma(host.Services))
	}
	if plan.Config.Migrations != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "migration:")
		fmt.Fprintf(w, "  image: %s\n", plan.Config.Migrations.Image)
		fmt.Fprintf(w, "  rollback: %t\n", len(plan.Config.Migrations.RollbackCommand) > 0 || plan.Config.Migrations.RestorePlan != "")
	}
}

func contains(items []string, wanted string) bool {
	for _, item := range items {
		if item == wanted {
			return true
		}
	}
	return false
}

func joinComma(items []string) string {
	copied := append([]string(nil), items...)
	sort.Strings(copied)
	out := ""
	for i, item := range copied {
		if i > 0 {
			out += ", "
		}
		out += item
	}
	return out
}
