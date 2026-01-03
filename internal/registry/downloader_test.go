package registry

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func mockGitHubResponse(downloadURL, tagName string) string {
	assetName := fmt.Sprintf("provider-test_%s_%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	return fmt.Sprintf(`{
		"tag_name": "%s",
		"assets": [
			{
				"name": "%s",
				"browser_download_url": "%s"
			},
			{
				"name": "provider-test_checksums.txt",
				"browser_download_url": "http://ignore.me"
			}
		]
	}`, tagName, assetName, downloadURL)
}

func TestEnsureProvider_NewInstall(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gohl-plugins")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	originalPluginDir := PluginDir
	PluginDir = tempDir
	defer func() { PluginDir = originalPluginDir }()

	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/binary" {
			w.Write([]byte("I AM A BINARY CONTENT"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockGitHubResponse(ts.URL+"/download/binary", "v1.0.0")))
	}))
	defer ts.Close()

	originalBaseURL := GitHubBaseURL
	GitHubBaseURL = ts.URL
	defer func() { GitHubBaseURL = originalBaseURL }()

	path, err := EnsureProvider("test", "github.com/fake/repo", "latest")
	if err != nil {
		t.Fatalf("EnsureProvider failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Binary was not created at %s", path)
	}

	content, _ := os.ReadFile(path)
	if string(content) != "I AM A BINARY CONTENT" {
		t.Errorf("Wrong content. Got: %s", string(content))
	}

	versionPath := path + ".version"
	vContent, _ := os.ReadFile(versionPath)
	if string(vContent) != "v1.0.0" {
		t.Errorf("Wrong version file content. Got: %s", string(vContent))
	}
}

func TestEnsureProvider_AlreadyUpToDate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gohl-plugins")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	PluginDir = tempDir
	defer func() { PluginDir = "./plugins" }()

	binaryName := "provider-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	localPath := filepath.Join(tempDir, binaryName)

	os.WriteFile(localPath, []byte("OLD BINARY"), 0755)
	os.WriteFile(localPath+".version", []byte("v1.0.0"), 0644)

	downloadHit := false

	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/binary" {
			downloadHit = true
			w.Write([]byte("NEW BINARY"))
			return
		}
		w.Write([]byte(mockGitHubResponse(ts.URL+"/download/binary", "v1.0.0")))
	}))
	defer ts.Close()

	GitHubBaseURL = ts.URL
	defer func() { GitHubBaseURL = "https://api.github.com" }()

	_, err = EnsureProvider("test", "github.com/fake/repo", "latest")
	if err != nil {
		t.Fatal(err)
	}

	if downloadHit {
		t.Error("Code downloaded file even though version was up to date")
	}
}

func TestEnsureProvider_UpdateNeeded(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gohl-plugins")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	PluginDir = tempDir
	defer func() { PluginDir = "./plugins" }()

	binaryName := "provider-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	localPath := filepath.Join(tempDir, binaryName)

	os.WriteFile(localPath, []byte("OLD BINARY"), 0755)
	os.WriteFile(localPath+".version", []byte("v0.9.0"), 0644)

	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/binary" {
			w.Write([]byte("NEW BINARY"))
			return
		}
		w.Write([]byte(mockGitHubResponse(ts.URL+"/download/binary", "v1.0.0")))
	}))
	defer ts.Close()

	GitHubBaseURL = ts.URL
	defer func() { GitHubBaseURL = "https://api.github.com" }()

	_, err = EnsureProvider("test", "github.com/fake/repo", "latest")
	if err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(localPath)
	if string(content) != "NEW BINARY" {
		t.Error("Binary was not updated")
	}

	vContent, _ := os.ReadFile(localPath + ".version")
	if string(vContent) != "v1.0.0" {
		t.Error("Version file was not updated")
	}
}
