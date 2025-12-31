package game

import "github.com/danielvollbro/gohl/pkg/plugin"

type GrandReport struct {
	Timestamp    string               `json:"timestamp"`
	TotalScore   int                  `json:"total_score"`
	MaxScore     int                  `json:"max_score"`
	Rank         string               `json:"rank"`
	PluginReports []*plugin.ScanReport `json:"details"`
}