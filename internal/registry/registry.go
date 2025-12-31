package registry

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/danielvollbro/gohl/internal/provider/docker"
	"github.com/danielvollbro/gohl/internal/provider/linux"
	"github.com/danielvollbro/gohl/internal/provider/system"
	"github.com/danielvollbro/gohl/pkg/plugin"
)

type factoryFunc func() plugin.Scanner

var catalog = map[string]factoryFunc{
	"system": func() plugin.Scanner { return system.New() },
	"docker": func() plugin.Scanner { return docker.New() },
	"linux":  func() plugin.Scanner { return linux.New() },
}

func GetProvider(name string) (plugin.Scanner, error) {
	factory, exists := catalog[name]
	if !exists {
		return nil, fmt.Errorf("provider not found")
	}
	return factory(), nil
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