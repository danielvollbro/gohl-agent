package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	api "github.com/danielvollbro/gohl-api"
)

const historyDirName = ".gohl/history"

func getHistoryDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, historyDirName)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return "", err
		}
	}
	return path, nil
}

func Save(report api.GrandReport) error {
	dir, err := getHistoryDir()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("report_%s.json", time.Now().Format("2006-01-02T15-04-05"))
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func LoadLatest() (*api.GrandReport, error) {
	dir, err := getHistoryDir()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var jsonFiles []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "report_") && strings.HasSuffix(f.Name(), ".json") {
			jsonFiles = append(jsonFiles, filepath.Join(dir, f.Name()))
		}
	}

	if len(jsonFiles) == 0 {
		return nil, nil
	}

	sort.Strings(jsonFiles)
	latestFile := jsonFiles[len(jsonFiles)-1]

	data, err := os.ReadFile(latestFile)
	if err != nil {
		return nil, err
	}

	var report api.GrandReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return &report, nil
}
