package binary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

	cmd.Env = os.Environ()
	for k, v := range config {
		envKey := fmt.Sprintf("GOHL_CONFIG_%s", strings.ToUpper(k))
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envKey, v))
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()

	if err != nil {
		errorMsg := strings.TrimSpace(stderr.String())

		if errorMsg != "" {
			return nil, fmt.Errorf("%s", errorMsg)
		}

		return nil, fmt.Errorf("failed to execute provider %s: %w", p.Path, err)
	}

	var report api.ScanReport
	if err := json.Unmarshal(output, &report); err != nil {
		return nil, fmt.Errorf("invalid json from provider %s: %w", p.Path, err)
	}

	return &report, nil
}
