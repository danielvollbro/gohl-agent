package ui

import (
	"encoding/json"
	"fmt"

	"github.com/danielvollbro/gohl/internal/game"
	"github.com/pterm/pterm"

	api "github.com/danielvollbro/gohl-api"
)

type Console struct {
	Silent bool
}

func New(silent bool) *Console {
	return &Console{Silent: silent}
}

func (c *Console) RenderLogo() {
	if c.Silent {
		return
	}
	pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("GO", pterm.NewStyle(pterm.FgCyan)),
		pterm.NewLettersFromStringWithStyle("HL", pterm.NewStyle(pterm.FgLightMagenta)),
	).Render()
	pterm.Println(pterm.Cyan("Game of Homelab") + " - " + pterm.LightMagenta("v0.1.0"))
	fmt.Println()
}

func (c *Console) StartSpinner(text string) (*pterm.SpinnerPrinter, error) {
	if c.Silent {
		return nil, nil
	}
	return pterm.DefaultSpinner.Start(text)
}

func (c *Console) PrintSuccess(format string, a ...interface{}) {
	if !c.Silent {
		pterm.Success.Printf(format+"\n", a...)
	}
}

func (c *Console) PrintError(format string, a ...interface{}) {
	if !c.Silent {
		pterm.Error.Printf(format+"\n", a...)
	}
}

func (c *Console) PrintWarning(format string, a ...interface{}) {
	if !c.Silent {
		pterm.Warning.Printf(format+"\n", a...)
	}
}

func (c *Console) RenderReport(report *api.ScanReport) {
	if c.Silent {
		return
	}

	pterm.DefaultSection.Println("Analysis Report: " + report.PluginID)

	tableData := pterm.TableData{
		{"ID", "CHECK", "STATUS", "SCORE"},
	}

	var failures []api.CheckResult

	for _, check := range report.Checks {
		status := pterm.FgGreen.Sprint("PASS")
		if !check.Passed {
			status = pterm.FgRed.Sprint("FAIL")
			failures = append(failures, check)
		}

		tableData = append(tableData, []string{
			check.ID,
			check.Name,
			status,
			fmt.Sprintf("%d/%d", check.Score, check.MaxScore),
		})
	}

	pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()

	if len(failures) > 0 {
		fmt.Println()
		pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).Println("🛑 ACTIVE QUESTS (Fix these to level up!)")

		for _, fail := range failures {
			pterm.Println(pterm.Red("● " + fail.Name))
			pterm.Println("  " + pterm.Yellow("Objective: ") + fail.Remediation)
			fmt.Println()
		}
	}
}

func (c *Console) RenderGrandTotal(score, maxScore int, rankName string, rankColor pterm.Color, previousScore int) {
	if c.Silent {
		return
	}

	style := pterm.NewStyle(rankColor)
	text := fmt.Sprintf("TOTAL SCORE: %d / %d", score, maxScore)

	if previousScore != -1 {
		diff := score - previousScore
		if diff > 0 {
			text += pterm.Green(fmt.Sprintf(" (+%d XP 📈)", diff))
		} else if diff < 0 {
			text += pterm.Red(fmt.Sprintf(" (%d XP 📉)", diff))
		} else {
			text += pterm.Gray(" (+0 XP)")
		}
	}

	text += fmt.Sprintf("\n\nRANK: %s", rankName)

	panel := pterm.DefaultBox.
		WithTitle("🏆 GAME OVER SUMMARY").
		WithTitleBottomCenter().
		WithBoxStyle(style).
		Sprint(text)

	pterm.DefaultCenter.Println(panel)
}

func (c *Console) Spacer() {
	if !c.Silent {
		fmt.Println()
	}
}

func (c *Console) PrintFinalResults(report game.GrandReport, asJson bool, previousScore int) {
	if asJson {
		jsonData, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Println("Error generating JSON:", err)
			return
		}
		fmt.Println(string(jsonData))
	} else {
		fmt.Println()

		for _, pluginReport := range report.PluginReports {
			c.RenderReport(pluginReport)
			fmt.Println()
		}

		_, rankColor := game.GetRank(report.TotalScore, report.MaxScore)
		c.RenderGrandTotal(report.TotalScore, report.MaxScore, report.Rank, rankColor, previousScore)
	}
}
