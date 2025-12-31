package game

import "github.com/pterm/pterm"

func GetRank(score, maxScore int) (string, pterm.Color) {
	if maxScore == 0 {
		return "Unranked", pterm.FgGray
	}

	percentage := float64(score) / float64(maxScore) * 100

	switch {
	case percentage == 100:
		return "HOMELAB GOD ⚡", pterm.FgLightMagenta
	case percentage >= 90:
		return "System Architect 🏗️", pterm.FgCyan
	case percentage >= 75:
		return "DevOps Engineer 🚀", pterm.FgGreen
	case percentage >= 50:
		return "Junior Sysadmin 🛠️", pterm.FgYellow
	case percentage >= 25:
		return "Script Kiddie 💻", pterm.FgRed
	default:
		return "Intern checking logs 📄", pterm.FgGray
	}
}