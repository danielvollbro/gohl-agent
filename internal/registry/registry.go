package registry

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/danielvollbro/gohl/internal/provider/binary"
	"github.com/danielvollbro/gohl/internal/provider/docker"
	"github.com/danielvollbro/gohl/pkg/plugin"
)

type factoryFunc func() plugin.Scanner

var catalog = map[string]factoryFunc{
	"docker": func() plugin.Scanner { return docker.New() },
}

func GetProvider(name string) (plugin.Scanner, error) {
	factory, exists := catalog[name]
	if exists {
		return factory(), nil
	}

	binaryPath := viper.GetString(name + ".path")
	if binaryPath != "" {
		// Kolla om filen finns
		if _, err := os.Stat(binaryPath); err == nil {
			return binary.New(name, binaryPath), nil
		}
	}

	return nil, fmt.Errorf("provider not found (and no path in config)")
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
