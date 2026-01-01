package plugin

import (
	"context"

	api "github.com/danielvollbro/gohl-api"
)

type Scanner interface {
	Info() api.PluginInfo
	Analyze(ctx context.Context, config map[string]string) (*api.ScanReport, error)
}
