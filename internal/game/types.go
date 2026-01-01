package game

import api "github.com/danielvollbro/gohl-api"

type GrandReport struct {
	Timestamp     string            `json:"timestamp"`
	TotalScore    int               `json:"total_score"`
	MaxScore      int               `json:"max_score"`
	Rank          string            `json:"rank"`
	PluginReports []*api.ScanReport `json:"details"`
}
