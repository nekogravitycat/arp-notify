package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
)

var filePath = "monitor_config.json"

var (
	_config     *MonitorConfig
	_onceConfig sync.Once
)

type MonitorConfig struct {
	Targets []TargetInfo `json:"targets"`
}

type TargetInfo struct {
	Mac       string   `json:"mac"`
	Message   string   `json:"message"`
	Receivers []string `json:"receivers"`
}

func GetMonitorConfig() *MonitorConfig {
	_onceConfig.Do(func() {
		var err error
		_config, err = loadMonitorConfig()
		if err != nil {
			log.Fatalf("Error loading monitor config: %v", err)
		}
	})

	return _config
}

func loadMonitorConfig() (*MonitorConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File does not exist, create an empty config file.
			if err := createEmptyMonitorConfig(); err != nil {
				return nil, fmt.Errorf("failed to create empty config file %q: %w", filePath, err)
			}
		}
		return nil, fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer file.Close()

	var cfg MonitorConfig

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &cfg, nil
}

func createEmptyMonitorConfig() error {
	emptyCfg := MonitorConfig{
		Targets: []TargetInfo{},
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", filePath, err)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(emptyCfg); err != nil {
		return fmt.Errorf("failed to write JSON to file %q: %w", filePath, err)
	}

	return nil
}
