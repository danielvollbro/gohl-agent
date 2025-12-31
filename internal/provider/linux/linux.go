package linux

import (
	"context"
	"os"
	"strings"

	"github.com/danielvollbro/gohl/pkg/plugin"
)

type LinuxProvider struct{
	SSHConfigPath string
	IPForwardPath string
}

func New() *LinuxProvider {
	return &LinuxProvider{
		SSHConfigPath: "/etc/ssh/sshd_config",
		IPForwardPath: "/proc/sys/net/ipv4/ip_forward",
	}
}

func (p *LinuxProvider) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:          "provider-linux",
		Name:        "Linux Security Auditor",
		Version:     "0.1.0",
		Description: "Audits SSH, Firewall and Kernel security settings",
		Author:      "GOHL Core",
	}
}

func (p *LinuxProvider) Analyze(ctx context.Context, config map[string]string) (*plugin.ScanReport, error) {
	var checks []plugin.CheckResult

	sshContent, err := os.ReadFile(p.SSHConfigPath)
	
	if err == nil {
		content := string(sshContent)

		// --- CHECK 1: SSH ROOT LOGIN ---
		hasNoRoot := strings.Contains(content, "PermitRootLogin no")
		
		checks = append(checks, plugin.CheckResult{
			ID:          "LNX-SSH-ROOT",
			Name:        "SSH Root Login Disabled",
			Description: "Checking if direct root login is disabled in sshd_config",
			Passed:      hasNoRoot,
			Score:       boolToScore(hasNoRoot, 20),
			MaxScore:    20,
			Remediation: "Set 'PermitRootLogin no' in \"" + p.SSHConfigPath + "\" and restart sshd.",
		})

		// --- CHECK 2: SSH PASSWORD AUTH ---
		hasNoPass := strings.Contains(content, "PasswordAuthentication no")
		checks = append(checks, plugin.CheckResult{
			ID:          "LNX-SSH-PASS",
			Name:        "SSH Key-Only Auth",
			Description: "Checking if PasswordAuthentication is disabled",
			Passed:      hasNoPass,
			Score:       boolToScore(hasNoPass, 15),
			MaxScore:    15,
			Remediation: "Use SSH keys! Set 'PasswordAuthentication no' in sshd_config.",
		})
	} else {
		checks = append(checks, plugin.CheckResult{
			ID:          "LNX-SSH-READ",
			Name:        "Read SSH Config",
			Description: "Attempting to read " + p.SSHConfigPath,
			Passed:      false,
			Score:       0,
			MaxScore:    0,
			Remediation: "Run GOHL with 'sudo' to scan protected config files.",
			Error:       err.Error(),
		})
	}

	// --- CHECK 3: IP FORWARDING ---
	fwdContent, err := os.ReadFile(p.IPForwardPath)
	if err == nil {
		val := strings.TrimSpace(string(fwdContent))
		isDisabled := val == "0"
		
		checks = append(checks, plugin.CheckResult{
			ID:          "LNX-NET-FWD",
			Name:        "IP Forwarding Disabled",
			Description: "Checking if kernel IP forwarding is off (0)",
			Passed:      isDisabled,
			Score:       boolToScore(isDisabled, 10),
			MaxScore:    10,
			Remediation: "Set net.ipv4.ip_forward=0 in /etc/sysctl.conf",
		})
	}

	return &plugin.ScanReport{
		PluginID: "provider-linux",
		Checks:   checks,
	}, nil
}

func boolToScore(passed bool, max int) int {
	if passed {
		return max
	}
	return 0
}