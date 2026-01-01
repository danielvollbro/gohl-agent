package binary

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	api "github.com/danielvollbro/gohl-api"
)

type BinaryProvider struct {
	Name string
	Path string
}

func New(name, path string) *BinaryProvider {
	return &BinaryProvider{Name: name, Path: path}
}

func (p *BinaryProvider) Info() api.PluginInfo {
	return api.PluginInfo{
		ID:   "external-" + p.Name,
		Name: "External: " + p.Name,
	}
}

func (p *BinaryProvider) Analyze(ctx context.Context, config map[string]string) (*api.ScanReport, error) {
	cmd := exec.CommandContext(ctx, p.Path)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute provider %s: %w", p.Path, err)
	}

	var report api.ScanReport
	if err := json.Unmarshal(output, &report); err != nil {
		return nil, fmt.Errorf("invalid json from provider %s: %w", p.Path, err)
	}

	return &report, nil
}
