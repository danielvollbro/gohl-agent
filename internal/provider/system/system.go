package system

import (
	"context"
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	
	"github.com/danielvollbro/gohl/pkg/plugin"
)

type StatsGatherer interface {
	GetDiskUsage(path string) (*disk.UsageStat, error)
	GetLoadAvg() (*load.AvgStat, error)
	GetNumCPU() int
}

type RealStatsGatherer struct{}

func (r *RealStatsGatherer) GetDiskUsage(path string) (*disk.UsageStat, error) {
	return disk.Usage(path)
}

func (r *RealStatsGatherer) GetLoadAvg() (*load.AvgStat, error) {
	return load.Avg()
}

func (r *RealStatsGatherer) GetNumCPU() int {
	return runtime.NumCPU()
}

type SystemProvider struct{
	Gatherer StatsGatherer
}

func New() *SystemProvider {
	return &SystemProvider{
		Gatherer: &RealStatsGatherer{},
	}
}

func (p *SystemProvider) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:          "provider-system",
		Name:        "System/OS Scanner",
		Version:     "0.1.0",
		Description: "Checks basic OS health metrics (Disk, CPU, etc)",
		Author:      "GOHL Core",
	}
}

func (p *SystemProvider) Analyze(ctx context.Context, config map[string]string) (*plugin.ScanReport, error) {
	var checks []plugin.CheckResult

	// --- CHECK 1: DISK USAGE ---
	usage, err := p.Gatherer.GetDiskUsage("/")
	if err == nil {
		passed := usage.UsedPercent < 90.0
		score := 0
		if passed {
			score = 10
		}

		checks = append(checks, plugin.CheckResult{
			ID:          "SYS-001",
			Name:        "Root Disk Usage",
			Description: fmt.Sprintf("Checking if disk usage (%.1f%%) is below 90%%", usage.UsedPercent),
			Passed:      passed,
			Score:       score,
			MaxScore:    10,
			Remediation: "Clean up disk space or expand the volume.",
		})
	}

	// --- CHECK 2: SYSTEM LOAD ---
	avg, err := p.Gatherer.GetLoadAvg()
	if err == nil {
		cores := float64(p.Gatherer.GetNumCPU())
		isOverloaded := avg.Load5 > (cores * 2)
		
		score := 0
		if !isOverloaded {
			score = 10
		}

		checks = append(checks, plugin.CheckResult{
			ID:          "SYS-002",
			Name:        "System Load (5m)",
			Description: fmt.Sprintf("Load is %.2f (Cores: %d)", avg.Load5, int(cores)),
			Passed:      !isOverloaded,
			Score:       score,
			MaxScore:    10,
			Remediation: "Check running processes, upgrade CPU or optimize workloads.",
		})
	}

	return &plugin.ScanReport{
		PluginID: "provider-system",
		Checks:   checks,
	}, nil
}