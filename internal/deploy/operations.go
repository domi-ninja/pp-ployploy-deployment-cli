package deploy

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

type Deployer struct {
	Root   string
	Out    io.Writer
	Err    io.Writer
	Runner Runner
}

func NewDeployer(root string, out io.Writer, errOut io.Writer) Deployer {
	return Deployer{
		Root: root,
		Out:  out,
		Err:  errOut,
		Runner: Runner{
			Stdout: out,
			Stderr: errOut,
		},
	}
}

func (d Deployer) Deploy() error {
	plan, err := LoadPlan(d.Root, "deploy.yml")
	if err != nil {
		return err
	}
	state, err := LoadState(d.Root)
	if err != nil {
		return err
	}
	bundle, err := RenderBundle(d.Root, plan)
	if err != nil {
		return err
	}

	record := recordFromPlan(plan, bundle, state.CurrentReleaseID)
	if err := SaveRelease(d.Root, record); err != nil {
		return err
	}

	fmt.Fprintf(d.Out, "release: %s\n", plan.ReleaseID)
	fmt.Fprintln(d.Out, "building image")
	if err := d.BuildImages(plan, bundle); err != nil {
		record.Status = "failed"
		record.Apply = StepRecord{Status: "failed", Error: err.Error(), At: time.Now().UTC()}
		_ = SaveRelease(d.Root, record)
		return err
	}

	migrationRan := false
	if plan.Config.Migrations != nil {
		fmt.Fprintln(d.Out, "running migration")
		if err := d.RunMigration(plan, false); err != nil {
			record.Status = "failed"
			record.Migration = StepRecord{Status: "failed", Error: err.Error(), At: time.Now().UTC()}
			_ = SaveRelease(d.Root, record)
			return err
		}
		migrationRan = true
		record.Migration = StepRecord{Status: "ok", At: time.Now().UTC()}
		_ = SaveRelease(d.Root, record)
	}

	fmt.Fprintln(d.Out, "applying hosts")
	if err := d.ApplyBundle(plan, bundle); err != nil {
		record.Status = "failed"
		record.Apply = StepRecord{Status: "failed", Error: err.Error(), At: time.Now().UTC()}
		_ = SaveRelease(d.Root, record)
		rollbackErrors := []string{}
		if state.CurrentReleaseID != "" {
			fmt.Fprintf(d.Err, "apply failed; re-applying previous release %s\n", state.CurrentReleaseID)
			previous, loadErr := LoadRelease(d.Root, state.CurrentReleaseID)
			if loadErr != nil {
				rollbackErrors = append(rollbackErrors, loadErr.Error())
			} else if rollbackErr := d.ApplyBundle(planForRecord(plan, previous), bundleFromRecord(previous)); rollbackErr != nil {
				rollbackErrors = append(rollbackErrors, rollbackErr.Error())
			}
		}
		if migrationRan {
			fmt.Fprintln(d.Err, "apply failed after migration; running migration rollback")
			if rollbackErr := d.RunMigration(plan, true); rollbackErr != nil {
				rollbackErrors = append(rollbackErrors, "migration rollback: "+rollbackErr.Error())
			}
		}
		if len(rollbackErrors) > 0 {
			return fmt.Errorf("%w; rollback also had errors: %s", err, strings.Join(rollbackErrors, "; "))
		}
		return err
	}

	now := time.Now().UTC()
	record.Status = "ok"
	record.Apply = StepRecord{Status: "ok", At: now}
	for i := range record.Hosts {
		record.Hosts[i].Status = "ok"
	}
	record.UpdatedAt = now
	if err := SaveRelease(d.Root, record); err != nil {
		return err
	}
	if err := SaveState(d.Root, State{
		Project:           plan.Config.Project.Name,
		Environment:       plan.Config.Project.Environment,
		CurrentReleaseID:  plan.ReleaseID,
		PreviousReleaseID: state.CurrentReleaseID,
	}); err != nil {
		return err
	}

	fmt.Fprintf(d.Out, "deployed %s\n", plan.ReleaseID)
	return nil
}

func (d Deployer) BuildImages(plan Plan, bundle Bundle) error {
	args := []string{"build", "-f", plan.Config.Build.Dockerfile}
	if plan.Config.Build.Target != "" {
		args = append(args, "--target", plan.Config.Build.Target)
	}
	if len(plan.Config.Build.Platforms) > 0 {
		args = append(args, "--platform", plan.Config.Build.Platforms[0])
	}
	for _, tag := range bundle.ImageTags {
		args = append(args, "-t", tag)
	}
	args = append(args, plan.Config.Build.Context)
	if err := d.Runner.Run(d.Root, "docker", args...); err != nil {
		return err
	}

	saveArgs := append([]string{"save", "-o", bundle.ImageTar}, bundle.ImageTags...)
	return d.Runner.Run(d.Root, "docker", saveArgs...)
}

func (d Deployer) RunMigration(plan Plan, rollback bool) error {
	migration := plan.Config.Migrations
	if migration == nil {
		return nil
	}
	return d.runMigration(migration, RenderTemplate(migration.Image, VarsForPlan(plan)), rollback)
}

func (d Deployer) runMigration(migration *Migration, image string, rollback bool) error {
	command := migration.Command
	if rollback {
		command = migration.RollbackCommand
	}
	if len(command) == 0 {
		if rollback && migration.RestorePlan != "" {
			return fmt.Errorf("restore plan requires manual action: %s", migration.RestorePlan)
		}
		return fmt.Errorf("migration command is empty")
	}

	args := []string{"run", "--rm"}
	if migration.Env.Source != "" {
		args = append(args, "--env-file", filepath.Join(d.Root, migration.Env.Source))
	}
	args = append(args, image)
	args = append(args, command...)
	return d.Runner.Run(d.Root, "docker", args...)
}

func (d Deployer) ApplyBundle(plan Plan, bundle Bundle) error {
	for _, host := range bundle.Hosts {
		fmt.Fprintf(d.Out, "host %s: upload\n", host.ID)
		if err := d.UploadHostBundle(host, bundle.ImageTar); err != nil {
			return fmt.Errorf("host %s upload: %w", host.ID, err)
		}
		fmt.Fprintf(d.Out, "host %s: docker load\n", host.ID)
		if err := d.Remote(host.SSH, "docker load -i "+shellQuote(host.RemoteDir+"/images/"+filepath.Base(bundle.ImageTar))); err != nil {
			return fmt.Errorf("host %s docker load: %w", host.ID, err)
		}
		fmt.Fprintf(d.Out, "host %s: compose up\n", host.ID)
		composeCmd := "cd " + shellQuote(host.RemoteDir) + " && docker compose -f compose.yml -p " + shellQuote(plan.Config.Project.Name) + " up -d"
		if err := d.Remote(host.SSH, composeCmd); err != nil {
			return fmt.Errorf("host %s compose up: %w", host.ID, err)
		}
	}
	return nil
}

func (d Deployer) UploadHostBundle(host HostBundle, imageTar string) error {
	if err := d.Remote(host.SSH, "mkdir -p "+shellQuote(host.RemoteDir)+"/env "+shellQuote(host.RemoteDir)+"/images"); err != nil {
		return err
	}
	if err := d.Copy(host.Compose, host.SSH, host.RemoteDir+"/compose.yml"); err != nil {
		return err
	}
	for _, envFile := range host.EnvFiles {
		if err := d.Copy(envFile, host.SSH, host.RemoteDir+"/env/"+filepath.Base(envFile)); err != nil {
			return err
		}
	}
	return d.Copy(imageTar, host.SSH, host.RemoteDir+"/images/"+filepath.Base(imageTar))
}

func (d Deployer) Remote(sshTarget string, command string) error {
	return d.Runner.Run(d.Root, "ssh", sshTarget, command)
}

func (d Deployer) Copy(local string, sshTarget string, remotePath string) error {
	return d.Runner.Run(d.Root, "scp", local, sshTarget+":"+remotePath)
}

func (d Deployer) Status() error {
	state, err := LoadState(d.Root)
	if err != nil {
		return err
	}
	if state.CurrentReleaseID == "" {
		fmt.Fprintln(d.Out, "no deployments recorded")
		return nil
	}
	record, err := LoadRelease(d.Root, state.CurrentReleaseID)
	if err != nil {
		return err
	}
	fmt.Fprintf(d.Out, "project: %s\n", state.Project)
	fmt.Fprintf(d.Out, "environment: %s\n", state.Environment)
	fmt.Fprintf(d.Out, "current: %s\n", state.CurrentReleaseID)
	fmt.Fprintf(d.Out, "previous: %s\n", emptyDash(state.PreviousReleaseID))
	fmt.Fprintf(d.Out, "status: %s\n", record.Status)
	fmt.Fprintf(d.Out, "git_sha: %s\n", record.Git.SHA)
	fmt.Fprintf(d.Out, "dirty: %t\n", record.Git.Dirty)
	fmt.Fprintln(d.Out, "hosts:")
	for _, host := range record.Hosts {
		fmt.Fprintf(d.Out, "  - %s (%s): %s\n", host.ID, host.SSH, host.Status)
	}
	return nil
}

func (d Deployer) Rollback() error {
	state, err := LoadState(d.Root)
	if err != nil {
		return err
	}
	if state.PreviousReleaseID == "" {
		return fmt.Errorf("no previous release recorded")
	}
	current, _ := LoadRelease(d.Root, state.CurrentReleaseID)
	previous, err := LoadRelease(d.Root, state.PreviousReleaseID)
	if err != nil {
		return err
	}
	plan, err := LoadPlan(d.Root, "deploy.yml")
	if err != nil {
		return err
	}

	if plan.Config.Migrations != nil && current.ReleaseID != "" {
		fmt.Fprintln(d.Out, "running DB rollback")
		image := plan.Config.Migrations.Image
		if len(current.ImageTags) > 0 {
			image = current.ImageTags[0]
		}
		if err := d.runMigration(plan.Config.Migrations, image, true); err != nil {
			return err
		}
	}

	bundle := bundleFromRecord(previous)
	fmt.Fprintf(d.Out, "rolling back to %s\n", previous.ReleaseID)
	if err := d.ApplyBundle(planForRecord(plan, previous), bundle); err != nil {
		return err
	}

	previous.Status = "ok"
	previous.Rollback = StepRecord{Status: "ok", At: time.Now().UTC()}
	if err := SaveRelease(d.Root, previous); err != nil {
		return err
	}
	return SaveState(d.Root, State{
		Project:           previous.Project,
		Environment:       previous.Environment,
		CurrentReleaseID:  previous.ReleaseID,
		PreviousReleaseID: current.ReleaseID,
	})
}

func recordFromPlan(plan Plan, bundle Bundle, previousReleaseID string) ReleaseRecord {
	hosts := make([]HostRecord, 0, len(bundle.Hosts))
	for _, host := range bundle.Hosts {
		hosts = append(hosts, HostRecord{
			ID:        host.ID,
			SSH:       host.SSH,
			Services:  host.ServiceIDs,
			RemoteDir: host.RemoteDir,
			Status:    "planned",
		})
	}
	now := time.Now().UTC()
	return ReleaseRecord{
		Project:           plan.Config.Project.Name,
		Environment:       plan.Config.Project.Environment,
		ReleaseID:         plan.ReleaseID,
		PreviousReleaseID: previousReleaseID,
		Git:               plan.Git,
		ImageTags:         bundle.ImageTags,
		ImageTar:          bundle.ImageTar,
		BundlePath:        bundle.Root,
		RemoteBase:        ".pp/" + plan.Config.Project.Name,
		Hosts:             hosts,
		Migration:         StepRecord{Status: "skipped"},
		Apply:             StepRecord{Status: "planned"},
		Status:            "planned",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func bundleFromRecord(record ReleaseRecord) Bundle {
	hosts := make([]HostBundle, 0, len(record.Hosts))
	for _, host := range record.Hosts {
		hostPath := filepath.Join(record.BundlePath, "hosts", host.ID)
		envFiles, _ := filepath.Glob(filepath.Join(hostPath, "env", "*.env"))
		hosts = append(hosts, HostBundle{
			ID:         host.ID,
			SSH:        host.SSH,
			Path:       hostPath,
			Compose:    filepath.Join(hostPath, "compose.yml"),
			EnvFiles:   envFiles,
			RemoteDir:  host.RemoteDir,
			ServiceIDs: host.Services,
		})
	}
	return Bundle{
		Root:      record.BundlePath,
		ImageTar:  record.ImageTar,
		ImageTags: record.ImageTags,
		Hosts:     hosts,
	}
}

func planForRecord(base Plan, record ReleaseRecord) Plan {
	base.ReleaseID = record.ReleaseID
	base.Git = record.Git
	return base
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
