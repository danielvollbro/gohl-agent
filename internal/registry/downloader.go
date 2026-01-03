package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	PluginDir      = "./plugins"
	GitHubBaseURL  = "https://api.github.com"
	ForceUserAgent = "gohl-agent-test" // Bra praxis
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func EnsureProvider(name, repoSource, requestedVersion string) (string, error) {
	if err := os.MkdirAll(PluginDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create plugin dir: %v", err)
	}

	binaryName := "provider-" + name
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	localPath := filepath.Join(PluginDir, binaryName)
	versionPath := localPath + ".version"
	downloadURL, resolvedVersion, err := resolveRemoteVersion(repoSource, requestedVersion)
	if err != nil {
		return "", fmt.Errorf("failed to resolve remote version: %v", err)
	}

	currentLocalVersion := ""
	if versionBytes, err := os.ReadFile(versionPath); err == nil {
		currentLocalVersion = strings.TrimSpace(string(versionBytes))
	}

	binExists := false
	if _, err := os.Stat(localPath); err == nil {
		binExists = true
	}

	if currentLocalVersion == resolvedVersion && binExists {
		// TODO: Check checksum or version to force update if needed
		return localPath, nil
	}

	if currentLocalVersion == "" {
		fmt.Printf("⬇️  Provider '%s' missing. Downloading from %s (%s)...\n", name, repoSource, resolvedVersion)
	} else {
		fmt.Printf("⬆️  Updating %s: %s -> %s\n", name, currentLocalVersion, resolvedVersion)
	}

	if err := downloadFile(localPath, downloadURL); err != nil {
		return "", fmt.Errorf("download failed: %v", err)
	}

	if err := os.WriteFile(versionPath, []byte(resolvedVersion), 0644); err != nil {
		fmt.Printf("⚠️ Warning: could not write version file: %v\n", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(localPath, 0755); err != nil {
			return "", fmt.Errorf("failed to chmod: %v", err)
		}
	}

	fmt.Printf("✅ Installed %s (%s) to %s\n", name, resolvedVersion, localPath)
	return localPath, nil
}

func resolveRemoteVersion(repoSource, version string) (string, string, error) {
	parts := strings.Split(repoSource, "/")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid source format")
	}
	owner, repo := parts[len(parts)-2], parts[len(parts)-1]

	apiURL := fmt.Sprintf("%s/repos/%s/%s/releases/latest", GitHubBaseURL, owner, repo)
	if version != "latest" && version != "" {
		apiURL = fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", GitHubBaseURL, owner, repo, version)
	}

	req, _ := http.NewRequest("GET", apiURL, nil)

	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("github api error: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	expectedSuffix := fmt.Sprintf("_%s_%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		expectedSuffix += ".exe"
	}
	expectedSuffix = strings.ToLower(expectedSuffix)

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".txt") {
			continue
		}
		if strings.HasSuffix(name, expectedSuffix) {
			return asset.BrowserDownloadURL, release.TagName, nil
		}
	}

	return "", "", fmt.Errorf("no binary ending in '%s' found in release %s", expectedSuffix, release.TagName)
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
