package registry

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/danielvollbro/gohl/internal/provider/binary"
	"github.com/danielvollbro/gohl/pkg/plugin"
)

func GetProvider(name string) (plugin.Scanner, error) {
	path := viper.GetString(name + ".path")
	source := viper.GetString(name + ".source")
	version := viper.GetString(name + ".version")

	if source != "" {
		if version == "" {
			version = "latest"
		}

		downloadedPath, err := EnsureProvider(name, source, version)
		if err != nil {
			return nil, fmt.Errorf("failed to download provider '%s': %w", name, err)
		}

		path = downloadedPath
	}

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			return binary.New(name, path), nil
		}
		return nil, fmt.Errorf("binary not found at path: %s", path)
	}

	return nil, fmt.Errorf("provider '%s' configuration missing 'source' or 'path'", name)
}

func GetConfig(providerName string) map[string]string {
	rawConfig := viper.GetStringMap(providerName)
	cleanConfig := make(map[string]string)

	for key, value := range rawConfig {
		switch v := value.(type) {
		case string:
			cleanConfig[key] = v
		case []interface{}:
			var items []string
			for _, item := range v {
				items = append(items, fmt.Sprint(item))
			}
			cleanConfig[key] = strings.Join(items, ",")
		default:
			cleanConfig[key] = fmt.Sprint(v)
		}
	}

	return cleanConfig
}
