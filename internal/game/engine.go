package game

import (
	"os"
	"time"

	api "github.com/danielvollbro/gohl-api"
)

func CompileReport(reports []*api.ScanReport) api.GrandReport {
	var totalScore, maxScore int

	for _, report := range reports {
		for _, check := range report.Checks {
			totalScore += check.Score
			maxScore += check.MaxScore
		}
	}

	hostname, _ := os.Hostname()
	rankName, _ := GetRank(totalScore, maxScore)
	return api.GrandReport{
		Hostname:      hostname,
		Timestamp:     time.Now().Format(time.RFC3339),
		TotalScore:    totalScore,
		MaxScore:      maxScore,
		Rank:          rankName,
		PluginReports: reports,
	}
}
