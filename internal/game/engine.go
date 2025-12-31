package game

import (
	"time"

	"github.com/danielvollbro/gohl/pkg/plugin"
)

func CompileReport(reports []*plugin.ScanReport) GrandReport {
	var totalScore, maxScore int
	
	for _, report := range reports {
		for _, check := range report.Checks {
			totalScore += check.Score
			maxScore += check.MaxScore
		}
	}

	rankName, _ := GetRank(totalScore, maxScore)
	return GrandReport{
		Timestamp:     time.Now().Format(time.RFC3339),
		TotalScore:    totalScore,
		MaxScore:      maxScore,
		Rank:          rankName,
		PluginReports: reports,
	}
}