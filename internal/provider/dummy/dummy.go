package dummy

import (
	"context"

	api "github.com/danielvollbro/gohl-api"
)

type DummyProvider struct{}

func New() *DummyProvider {
	return &DummyProvider{}
}

func (p *DummyProvider) Info() api.PluginInfo {
	return api.PluginInfo{
		ID:          "provider-dummy",
		Name:        "Dummy System Scanner",
		Version:     "0.1.0",
		Description: "A fake scanner to test the game engine",
		Author:      "GOHL Team",
	}
}

func (p *DummyProvider) Analyze(ctx context.Context, config map[string]string) (*api.ScanReport, error) {
	checks := []api.CheckResult{
		{
			ID:          "DUMMY-001",
			Name:        "Check if Homelab is cool",
			Description: "Analyzes coolness factor of the lab",
			Passed:      true,
			Score:       100,
			MaxScore:    100,
			Remediation: "",
		},
		{
			ID:          "DUMMY-002",
			Name:        "Root Password Check",
			Description: "Checking if root password is 'password123'",
			Passed:      false,
			Score:       0,
			MaxScore:    50,
			Remediation: "Change root password using 'passwd' command immediately.",
		},
		{
			ID:          "DUMMY-003",
			Name:        "Backup Strategy",
			Description: "Checking for backup.sh in /root",
			Passed:      true,
			Score:       25,
			MaxScore:    25,
			Remediation: "",
		},
	}

	return &api.ScanReport{
		PluginID: "provider-dummy",
		Checks:   checks,
	}, nil
}
