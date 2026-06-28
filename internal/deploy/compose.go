package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type ComposeFile struct {
	Services map[string]ComposeService `yaml:"services"`
	Volumes  map[string]ComposeVolume  `yaml:"volumes,omitempty"`
}

type ComposeService struct {
	Image       string            `yaml:"image"`
	Restart     string            `yaml:"restart,omitempty"`
	Command     []string          `yaml:"command,omitempty"`
	Entrypoint  []string          `yaml:"entrypoint,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	EnvFile     []string          `yaml:"env_file,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Healthcheck *Healthcheck      `yaml:"healthcheck,omitempty"`
}

type Healthcheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty"`
	Retries  int      `yaml:"retries,omitempty"`
}

type ComposeVolume struct {
	Driver   string `yaml:"driver,omitempty"`
	External bool   `yaml:"external,omitempty"`
}

type Bundle struct {
	Root      string
	ImageTar  string
	ImageTags []string
	Hosts     []HostBundle
}

type HostBundle struct {
	ID         string
	SSH        string
	Path       string
	Compose    string
	Routes     string
	EnvFiles   []string
	RemoteDir  string
	ServiceIDs []string
}

func RenderBundle(root string, plan Plan) (Bundle, error) {
	vars := VarsForPlan(plan)
	bundleRoot := filepath.Join(root, ".deploy", "releases", plan.ReleaseID)
	imagesDir := filepath.Join(bundleRoot, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return Bundle{}, fmt.Errorf("create image dir: %w", err)
	}

	imageTags := RenderTemplates(plan.Config.Build.Tags, vars)
	if len(imageTags) == 0 {
		imageTags = []string{fmt.Sprintf("%s:%s", plan.Config.Project.Name, plan.Git.SHA)}
	}
	imageTar := filepath.Join(imagesDir, plan.Config.Project.Name+"-"+plan.ReleaseID+".tar")

	bundle := Bundle{
		Root:      bundleRoot,
		ImageTar:  imageTar,
		ImageTags: imageTags,
	}

	for _, hostPlan := range plan.Hosts {
		hostDir := filepath.Join(bundleRoot, "hosts", hostPlan.ID)
		envDir := filepath.Join(hostDir, "env")
		if err := os.MkdirAll(envDir, 0700); err != nil {
			return Bundle{}, fmt.Errorf("create host bundle dir: %w", err)
		}

		compose := ComposeFile{
			Services: map[string]ComposeService{},
			Volumes:  map[string]ComposeVolume{},
		}
		envFiles := []string{}

		for _, serviceID := range hostPlan.Services {
			service := plan.Config.Services[serviceID]
			envFile, err := renderServiceEnv(root, envDir, serviceID, service.Env)
			if err != nil {
				return Bundle{}, err
			}
			if envFile != "" {
				envFiles = append(envFiles, envFile)
			}

			composeService := renderComposeService(plan, serviceID, service, vars, envFile)
			compose.Services[serviceID] = composeService
			for _, mount := range service.Volumes {
				volume := plan.Config.Volumes[mount.Name]
				compose.Volumes[mount.Name] = ComposeVolume{
					Driver:   volume.Driver,
					External: volume.External,
				}
			}
		}
		if len(compose.Volumes) == 0 {
			compose.Volumes = nil
		}

		composePath := filepath.Join(hostDir, "compose.yml")
		if err := writeYAML(composePath, compose); err != nil {
			return Bundle{}, err
		}
		routesPath, err := renderCaddyRoutes(hostDir, plan, hostPlan.ID)
		if err != nil {
			return Bundle{}, err
		}

		bundle.Hosts = append(bundle.Hosts, HostBundle{
			ID:         hostPlan.ID,
			SSH:        hostPlan.SSH,
			Path:       hostDir,
			Compose:    composePath,
			Routes:     routesPath,
			EnvFiles:   envFiles,
			RemoteDir:  remoteReleaseDir(plan.Config.Project.Name, plan.ReleaseID),
			ServiceIDs: hostPlan.Services,
		})
	}

	return bundle, nil
}

func renderCaddyRoutes(hostDir string, plan Plan, hostID string) (string, error) {
	var body strings.Builder
	for _, route := range plan.Config.Routes {
		if !routeBelongsToHost(plan.Config, route, hostID) {
			continue
		}
		body.WriteString(route.Host)
		body.WriteString(" {\n")
		body.WriteString("\treverse_proxy ")
		body.WriteString(route.Target)
		body.WriteString("\n")
		body.WriteString("}\n\n")
	}
	if body.Len() == 0 {
		return "", nil
	}
	path := filepath.Join(hostDir, "routes.caddy")
	if err := os.WriteFile(path, []byte(body.String()), 0644); err != nil {
		return "", fmt.Errorf("write caddy routes: %w", err)
	}
	return path, nil
}

func routeBelongsToHost(cfg Config, route Route, hostID string) bool {
	if route.HostID != "" {
		return route.HostID == hostID
	}
	service, ok := cfg.Services[route.Service]
	return ok && contains(service.Hosts, hostID)
}

func renderComposeService(plan Plan, serviceID string, service Service, vars RenderVars, envFile string) ComposeService {
	labels := map[string]string{
		"pp.project":     plan.Config.Project.Name,
		"pp.environment": plan.Config.Project.Environment,
		"pp.release":     plan.ReleaseID,
		"pp.service":     serviceID,
		"pp.git_sha":     plan.Git.SHA,
	}

	out := ComposeService{
		Image:      RenderTemplate(service.Image, vars),
		Restart:    "unless-stopped",
		Command:    service.Command,
		Entrypoint: service.Entrypoint,
		Ports:      renderPorts(service.Ports),
		Volumes:    renderVolumeMounts(service.Volumes),
		Labels:     labels,
	}
	if envFile != "" {
		out.EnvFile = []string{filepath.ToSlash(filepath.Join("env", filepath.Base(envFile)))}
	}
	if service.Health.HTTP != "" {
		out.Healthcheck = &Healthcheck{
			Test:     []string{"CMD-SHELL", "wget -q --spider " + shellEscapeHealthURL(service.Health.HTTP)},
			Interval: "10s",
			Timeout:  "5s",
			Retries:  healthRetries(service.Health.TimeoutSeconds),
		}
	} else if len(service.Health.Command) > 0 {
		out.Healthcheck = &Healthcheck{
			Test:     append([]string{"CMD"}, service.Health.Command...),
			Interval: "10s",
			Timeout:  "5s",
			Retries:  healthRetries(service.Health.TimeoutSeconds),
		}
	}
	return out
}

func renderServiceEnv(root string, envDir string, serviceID string, env EnvSpec) (string, error) {
	if env.Source == "" {
		return "", nil
	}
	values, err := LoadEnvFile(filepath.Join(root, env.Source))
	if err != nil {
		return "", fmt.Errorf("load env for %s: %w", serviceID, err)
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var body strings.Builder
	for _, key := range keys {
		body.WriteString(key)
		body.WriteByte('=')
		body.WriteString(escapeEnvValue(values[key]))
		body.WriteByte('\n')
	}

	path := filepath.Join(envDir, serviceID+".env")
	if err := os.WriteFile(path, []byte(body.String()), 0600); err != nil {
		return "", fmt.Errorf("write env file for %s: %w", serviceID, err)
	}
	return path, nil
}

func renderPorts(ports []Port) []string {
	out := make([]string, 0, len(ports))
	for _, port := range ports {
		value := strconv.Itoa(port.Published) + ":" + strconv.Itoa(port.Target)
		if port.HostIP != "" {
			value = port.HostIP + ":" + value
		}
		out = append(out, value)
	}
	return out
}

func renderVolumeMounts(mounts []VolumeMount) []string {
	out := make([]string, 0, len(mounts))
	for _, mount := range mounts {
		out = append(out, mount.Name+":"+mount.Target)
	}
	return out
}

func healthRetries(timeoutSeconds int) int {
	if timeoutSeconds <= 0 {
		return 6
	}
	retries := timeoutSeconds / 10
	if retries < 1 {
		return 1
	}
	return retries
}

func shellEscapeHealthURL(value string) string {
	return strings.ReplaceAll(value, "'", "'\"'\"'")
}

func escapeEnvValue(value string) string {
	if strings.ContainsAny(value, "\n\r") {
		value = strings.NewReplacer("\n", "\\n", "\r", "\\r").Replace(value)
	}
	return value
}

func writeYAML(path string, value any) error {
	body, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	if err := os.WriteFile(path, body, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func remoteReleaseDir(project string, releaseID string) string {
	return ".pp/" + project + "/releases/" + releaseID
}
