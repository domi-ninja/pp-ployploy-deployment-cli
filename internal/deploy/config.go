package deploy

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version    int                `yaml:"version"`
	Project    Project            `yaml:"project"`
	Build      Build              `yaml:"build"`
	Hosts      map[string]Host    `yaml:"hosts"`
	Services   map[string]Service `yaml:"services"`
	Volumes    map[string]Volume  `yaml:"volumes"`
	Migrations *Migration         `yaml:"migrations"`
	Checks     map[string][]Check `yaml:"checks"`
}

type Project struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
}

type Build struct {
	Context    string   `yaml:"context"`
	Dockerfile string   `yaml:"dockerfile"`
	Target     string   `yaml:"target"`
	Platforms  []string `yaml:"platforms"`
	Tags       []string `yaml:"tags"`
}

type Host struct {
	SSH   string    `yaml:"ssh"`
	Agent HostAgent `yaml:"agent"`
	Roles []string  `yaml:"roles"`
}

type HostAgent struct {
	LocalURL string `yaml:"local_url"`
}

type Service struct {
	Image      string        `yaml:"image"`
	Hosts      []string      `yaml:"hosts"`
	Command    []string      `yaml:"command"`
	Entrypoint []string      `yaml:"entrypoint"`
	Env        EnvSpec       `yaml:"env"`
	Ports      []Port        `yaml:"ports"`
	Volumes    []VolumeMount `yaml:"volumes"`
	Health     Health        `yaml:"health"`
}

type EnvSpec struct {
	Source   string   `yaml:"source"`
	Required []string `yaml:"required"`
}

type Port struct {
	HostIP    string `yaml:"host_ip"`
	Published int    `yaml:"published"`
	Target    int    `yaml:"target"`
}

type VolumeMount struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
}

type Health struct {
	HTTP           string   `yaml:"http"`
	Command        []string `yaml:"command"`
	TimeoutSeconds int      `yaml:"timeout_seconds"`
}

type Volume struct {
	Driver   string `yaml:"driver"`
	External bool   `yaml:"external"`
}

type Migration struct {
	Image           string   `yaml:"image"`
	Command         []string `yaml:"command"`
	RollbackCommand []string `yaml:"rollback_command"`
	RestorePlan     string   `yaml:"restore_plan"`
	Env             EnvSpec  `yaml:"env"`
	Run             string   `yaml:"run"`
	TimeoutSeconds  int      `yaml:"timeout_seconds"`
}

type Check struct {
	Name         string `yaml:"name"`
	URL          string `yaml:"url"`
	ExpectStatus int    `yaml:"expect_status"`
}

type ValidationError struct {
	Problems []string
}

func (e ValidationError) Error() string {
	out := ""
	for _, problem := range e.Problems {
		out += fmt.Sprintf("- %s\n", problem)
	}
	return out
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func LoadConfig(root string, configPath string) (Config, error) {
	fullPath := filepath.Join(root, configPath)
	body, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("missing %s in %s", configPath, root)
		}
		return Config{}, fmt.Errorf("read %s: %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(body, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", configPath, err)
	}

	if err := ValidateConfig(root, cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func ValidateConfig(root string, cfg Config) error {
	var problems []string

	if cfg.Version != 1 {
		problems = append(problems, "version must be 1")
	}
	validateSlug(&problems, "project.name", cfg.Project.Name)
	validateSlug(&problems, "project.environment", cfg.Project.Environment)
	require(&problems, "build.context", cfg.Build.Context)
	require(&problems, "build.dockerfile", cfg.Build.Dockerfile)

	if len(cfg.Hosts) == 0 {
		problems = append(problems, "hosts must contain at least one host")
	}
	if len(cfg.Services) == 0 {
		problems = append(problems, "services must contain at least one service")
	}

	for _, item := range sortedMap(cfg.Hosts) {
		hostID := item.Key
		host := item.Value
		validateSlug(&problems, "hosts."+hostID, hostID)
		require(&problems, "hosts."+hostID+".ssh", host.SSH)
		if host.Agent.LocalURL == "" {
			problems = append(problems, "hosts."+hostID+".agent.local_url is required")
		} else if err := validateLocalURL(host.Agent.LocalURL); err != nil {
			problems = append(problems, "hosts."+hostID+".agent.local_url "+err.Error())
		}
	}

	hostPorts := map[string]map[int]string{}
	for _, item := range sortedMap(cfg.Services) {
		serviceID := item.Key
		service := item.Value
		validateSlug(&problems, "services."+serviceID, serviceID)
		require(&problems, "services."+serviceID+".image", service.Image)
		if len(service.Hosts) == 0 {
			problems = append(problems, "services."+serviceID+".hosts must contain at least one host")
		}
		for _, hostID := range service.Hosts {
			if _, ok := cfg.Hosts[hostID]; !ok {
				problems = append(problems, "services."+serviceID+".hosts references unknown host "+hostID)
				continue
			}
			if _, ok := hostPorts[hostID]; !ok {
				hostPorts[hostID] = map[int]string{}
			}
			for _, port := range service.Ports {
				if port.Published <= 0 || port.Target <= 0 {
					problems = append(problems, "services."+serviceID+".ports must use positive published and target ports")
					continue
				}
				if owner, exists := hostPorts[hostID][port.Published]; exists {
					problems = append(problems, fmt.Sprintf("host %s published port %d conflicts between services %s and %s", hostID, port.Published, owner, serviceID))
				}
				hostPorts[hostID][port.Published] = serviceID
			}
		}
		validateEnv(&problems, root, "services."+serviceID+".env", service.Env)
		validateHealth(&problems, "services."+serviceID+".health", service.Health, len(service.Ports) > 0)
		for _, mount := range service.Volumes {
			if mount.Name == "" {
				problems = append(problems, "services."+serviceID+".volumes.name is required")
				continue
			}
			if _, ok := cfg.Volumes[mount.Name]; !ok {
				problems = append(problems, "services."+serviceID+".volumes references unknown volume "+mount.Name)
			}
		}
	}

	for volumeID := range cfg.Volumes {
		validateSlug(&problems, "volumes."+volumeID, volumeID)
	}

	if cfg.Migrations != nil {
		validateMigration(&problems, root, *cfg.Migrations)
	}

	if len(problems) > 0 {
		return ValidationError{Problems: problems}
	}
	return nil
}

func validateMigration(problems *[]string, root string, migration Migration) {
	require(problems, "migrations.image", migration.Image)
	if len(migration.Command) == 0 {
		*problems = append(*problems, "migrations.command must contain at least one argument")
	}
	if len(migration.RollbackCommand) == 0 && migration.RestorePlan == "" {
		*problems = append(*problems, "migrations.rollback_command or migrations.restore_plan is required")
	}
	if migration.Run != "" && migration.Run != "before_services" {
		*problems = append(*problems, "migrations.run must be before_services when set")
	}
	validateEnv(problems, root, "migrations.env", migration.Env)
}

func validateEnv(problems *[]string, root string, field string, env EnvSpec) {
	if env.Source == "" && len(env.Required) == 0 {
		return
	}
	if env.Source == "" {
		*problems = append(*problems, field+".source is required when required env names are declared")
		return
	}
	values, err := LoadEnvFile(filepath.Join(root, env.Source))
	if err != nil {
		*problems = append(*problems, field+".source "+err.Error())
		return
	}
	for _, name := range env.Required {
		if _, ok := values[name]; !ok {
			*problems = append(*problems, field+".required missing "+name+" in "+env.Source)
		}
	}
}

func validateHealth(problems *[]string, field string, health Health, externallyReachable bool) {
	if !externallyReachable {
		return
	}
	if health.HTTP == "" && len(health.Command) == 0 {
		*problems = append(*problems, field+" is required for services with published ports")
	}
}

func validateLocalURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("must be a valid URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("must use http or https")
	}
	if parsed.Hostname() != "127.0.0.1" && parsed.Hostname() != "localhost" {
		return fmt.Errorf("must point to localhost or 127.0.0.1")
	}
	return nil
}

func require(problems *[]string, field string, value string) {
	if value == "" {
		*problems = append(*problems, field+" is required")
	}
}

func validateSlug(problems *[]string, field string, value string) {
	if value == "" {
		*problems = append(*problems, field+" is required")
		return
	}
	if !slugPattern.MatchString(value) {
		*problems = append(*problems, field+" must be a lowercase DNS-safe slug")
	}
}

func sortedMap[T any](items map[string]T) []struct {
	Key   string
	Value T
} {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]struct {
		Key   string
		Value T
	}, 0, len(keys))
	for _, key := range keys {
		out = append(out, struct {
			Key   string
			Value T
		}{Key: key, Value: items[key]})
	}
	return out
}
