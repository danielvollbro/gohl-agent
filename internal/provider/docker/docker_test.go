package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

type MockDockerClient struct {
	Containers []types.Container
	Inspects   map[string]types.ContainerJSON
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	return m.Containers, nil
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if val, ok := m.Inspects[containerID]; ok {
		return val, nil
	}
	return types.ContainerJSON{}, nil
}

func (m *MockDockerClient) Close() error {
	return nil
}

func TestAnalyze_SecureContainer(t *testing.T) {
	mock := &MockDockerClient{
		Containers: []types.Container{
			{ID: "c1", Names: []string{"/secure-app"}},
		},
		Inspects: map[string]types.ContainerJSON{
			"c1": {
				ContainerJSONBase: &types.ContainerJSONBase{
					HostConfig: &container.HostConfig{
						RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
					},
				},
				Config: &container.Config{
					User: "1000",
				},
			},
		},
	}

	p := &DockerProvider{Client: mock}

	report, err := p.Analyze(context.Background(), nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(report.Checks) != 2 {
		t.Fatalf("Expected 2 checks, got %d", len(report.Checks))
	}

	for _, check := range report.Checks {
		if !check.Passed {
			t.Errorf("Check %s failed unexpectedly", check.ID)
		}
		if check.Score != check.MaxScore {
			t.Errorf("Check %s did not give max score", check.ID)
		}
	}
}

func TestAnalyze_InsecureContainer(t *testing.T) {
	mock := &MockDockerClient{
		Containers: []types.Container{
			{ID: "c2", Names: []string{"/yolo-app"}},
		},
		Inspects: map[string]types.ContainerJSON{
			"c2": {
				ContainerJSONBase: &types.ContainerJSONBase{
					HostConfig: &container.HostConfig{
						RestartPolicy: container.RestartPolicy{Name: "no"},
					},
				},
				Config: &container.Config{
					User: "root",
				},
			},
		},
	}

	p := &DockerProvider{Client: mock}
	report, _ := p.Analyze(context.Background(), nil)

	for _, check := range report.Checks {
		if check.Passed {
			t.Errorf("Check %s passed unexpectedly (should fail)", check.ID)
		}
		if check.Score != 0 {
			t.Errorf("Check %s gave points, expected 0", check.ID)
		}
	}
}

func TestAnalyze_IgnoreList(t *testing.T) {
	mock := &MockDockerClient{
		Containers: []types.Container{
			{ID: "c3", Names: []string{"/ignored-app"}},
		},
	}

	p := &DockerProvider{Client: mock}
	
	config := map[string]string{
		"ignore": "ignored-app,other-app",
	}

	report, _ := p.Analyze(context.Background(), config)

	if len(report.Checks) != 0 {
		if len(report.Checks) != 0 {
			t.Errorf("Expected 0 checks (ignored), got %d", len(report.Checks))
		}
	}
}