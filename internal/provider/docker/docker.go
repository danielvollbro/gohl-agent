package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	api "github.com/danielvollbro/gohl-api"
)

type DockerClient interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	Close() error
}

type DockerProvider struct {
	Client DockerClient
}

func New() *DockerProvider {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return &DockerProvider{Client: nil}
	}

	return &DockerProvider{
		Client: cli,
	}
}

func (p *DockerProvider) Info() api.PluginInfo {
	return api.PluginInfo{
		ID:          "provider-docker",
		Name:        "Docker Container Inspector",
		Version:     "0.1.0",
		Description: "Analyzes security and reliability of running containers",
		Author:      "GOHL Core",
	}
}

func isIgnored(name string, ignoreList string) bool {
	if ignoreList == "" {
		return false
	}
	items := strings.Split(ignoreList, ",")
	for _, item := range items {
		if strings.TrimSpace(item) == name {
			return true
		}
	}
	return false
}

func (p *DockerProvider) Analyze(ctx context.Context, config map[string]string) (*api.ScanReport, error) {
	if p.Client == nil {
		return nil, fmt.Errorf("docker client not initialized (is docker running?)")
	}

	containers, err := p.Client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var checks []api.CheckResult
	ignoreList := config["ignore"]

	for _, c := range containers {
		name := "unknown"
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		if isIgnored(name, ignoreList) {
			continue
		}

		inspect, err := p.Client.ContainerInspect(ctx, c.ID)
		if err == nil {
			policy := inspect.HostConfig.RestartPolicy.Name
			passedRestart := policy != "" && policy != "no"

			scoreRestart := 0
			if passedRestart {
				scoreRestart = 5
			}

			checks = append(checks, api.CheckResult{
				ID:          fmt.Sprintf("DKR-RESTART-%s", name),
				Name:        fmt.Sprintf("Restart Policy: %s", name),
				Description: fmt.Sprintf("Container %s has policy '%s'", name, policy),
				Passed:      passedRestart,
				Score:       scoreRestart,
				MaxScore:    5,
				Remediation: fmt.Sprintf("Run 'docker update --restart=unless-stopped %s'", name),
			})

			user := inspect.Config.User
			isRoot := user == "" || user == "0" || user == "root"
			passedRoot := !isRoot

			scoreRoot := 0
			if passedRoot {
				scoreRoot = 10
			}

			checks = append(checks, api.CheckResult{
				ID:          fmt.Sprintf("DKR-ROOT-%s", name),
				Name:        fmt.Sprintf("Non-Root User: %s", name),
				Description: fmt.Sprintf("Checking if %s runs as root", name),
				Passed:      passedRoot,
				Score:       scoreRoot,
				MaxScore:    10,
				Remediation: fmt.Sprintf("Set a generic user (PUID/PGID) in your compose file for %s", name),
			})
		}
	}

	if len(containers) == 0 {
		checks = append(checks, api.CheckResult{
			ID:          "DKR-000",
			Name:        "Docker Discovery",
			Description: "Checking for containers",
			Passed:      false,
			Score:       0,
			MaxScore:    0,
			Remediation: "No containers found. Start some services!",
		})
	}

	return &api.ScanReport{
		PluginID: "provider-docker",
		Checks:   checks,
	}, nil
}
