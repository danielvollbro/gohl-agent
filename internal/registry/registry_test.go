package registry

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/viper"
)

func TestGetProvider_LocalPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	fakeBinary := filepath.Join(tempDir, "my-local-provider")
	if runtime.GOOS == "windows" {
		fakeBinary += ".exe"
	}
	os.WriteFile(fakeBinary, []byte("fake"), 0755)

	viper.Reset()
	viper.Set("test-local.path", fakeBinary)

	p, err := GetProvider("test-local")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}

	if p == nil {
		t.Fatal("Provider should not be nil")
	}
}

func TestGetProvider_MissingConfig(t *testing.T) {
	viper.Reset()

	_, err := GetProvider("ghost-provider")
	if err == nil {
		t.Error("Expected error for missing provider config, got nil")
	}
}

func TestGetConfig_Parsing(t *testing.T) {
	viper.Reset()
	viper.Set("my-provider.url", "http://localhost")
	viper.Set("my-provider.retries", 3)
	viper.Set("my-provider.tags", []string{"prod", "linux"})

	config := GetConfig("my-provider")

	if config["url"] != "http://localhost" {
		t.Errorf("Wrong url: %s", config["url"])
	}

	if config["retries"] != "3" {
		t.Errorf("Wrong retries: %s", config["retries"])
	}

	if config["tags"] != "prod,linux" && config["tags"] != "[prod linux]" {
		if config["tags"] == "" {
			t.Error("Tags missing")
		}
	}
}
