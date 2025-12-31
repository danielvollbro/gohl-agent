package main

import (
	"context"
	"fmt"
	"os"
	"time"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/danielvollbro/gohl/internal/client"
	"github.com/danielvollbro/gohl/internal/game"
	"github.com/danielvollbro/gohl/internal/registry"
	"github.com/danielvollbro/gohl/internal/storage"
	"github.com/danielvollbro/gohl/internal/ui"
	"github.com/danielvollbro/gohl/pkg/plugin"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gohl",
	Short: "Gamify your Homelab infrastructure",
	Long:  `GOHL (Game of Homelab) - Level up your infrastructure!`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.New(false).RenderLogo()
		cmd.Help()
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Start analyzing the local environment",
	Run: func(cmd *cobra.Command, args []string) {
		useJson, _ := cmd.Flags().GetBool("json")
		console := ui.New(useJson)

		console.RenderLogo()

		spinner, _ := console.StartSpinner("Initializing sensors...")
		if spinner != nil {
			time.Sleep(time.Second * 1)
			spinner.Success("Sensors initialized")
		}
		
		enabledProviders := viper.GetStringSlice("providers")
		if len(enabledProviders) == 0 {
			pterm.Warning.Println("No providers defined in gohl.yaml, running default: system")
			enabledProviders = []string{"system"}
		}

		console.Spacer()

		ctx := context.Background()
		var allReports []*plugin.ScanReport
		for _, name := range enabledProviders {
			scanner, err := registry.GetProvider(name)
			if err != nil {
				console.PrintError("Unknown provider in config: '%s' (skipping)", name)
				continue
			}
			
			console.PrintSuccess("Enabled provider: %s\n", name)

			cfg := registry.GetConfig(name)

			info := scanner.Info()
			scanSpinner, _ := console.StartSpinner(fmt.Sprintf("Running %s...", info.Name))
			
			report, err := scanner.Analyze(ctx, cfg)
			if err != nil {
				if scanSpinner != nil {
					scanSpinner.Fail(fmt.Sprintf("%s failed: %v", info.Name, err))
				}
				continue
			}

			if scanSpinner != nil {
				scanSpinner.Success(fmt.Sprintf("%s complete", info.Name))
			}

			allReports = append(allReports, report)
		}

		console.Spacer()

		grandReport := game.CompileReport(allReports)

		previousScore := -1
		lastReport, err := storage.LoadLatest()
		if err == nil && lastReport != nil {
			previousScore = lastReport.TotalScore
		}

		console.PrintFinalResults(grandReport, useJson, previousScore)

		if err := storage.Save(grandReport); err != nil {
			console.PrintWarning("Could not save history: %v", err)
		}

		// --- CLOUD UPLOAD ---
		shouldSubmit, _ := cmd.Flags().GetBool("submit")
		
		if shouldSubmit {
			console.Spacer()
			
			serverURL := viper.GetString("server_url")
			if serverURL == "" {
				console.PrintError("Cannot submit: 'server_url' is missing in gohl.yaml")
				return
			}

			spinner, _ := console.StartSpinner("Uploading results to cloud...")
			
			err := client.UploadReport(serverURL, grandReport)
			if err != nil {
				if spinner != nil { spinner.Fail("Upload failed: " + err.Error()) }
			} else {
				if spinner != nil { spinner.Success("Successfully uploaded to leaderboard!") }
			}
		}
	},
}

func getDockerConfig() map[string]string {
	cfg := make(map[string]string)
	ignoredContainers := viper.GetStringSlice("docker.ignore")
	if len(ignoredContainers) > 0 {
		cfg["ignore"] = strings.Join(ignoredContainers, ",")
	}
	return cfg
}

func init() {
	cobra.OnInitialize(initConfig)
	scanCmd.Flags().Bool("json", false, "Output results as JSON for integrations")
	scanCmd.Flags().Bool("submit", false, "Upload results to the configured server")
	rootCmd.AddCommand(scanCmd)
}

func initConfig() {
    viper.SetConfigName("gohl")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".") 

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            fmt.Println("Config file error:", err)
        }
    }
}
