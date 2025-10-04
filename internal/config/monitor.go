package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

var filePath = "monitor_targets.json"

type TargetsFile struct {
	Targets []Target `json:"targets"`
}

type Target struct {
	Mac       string   `json:"mac"`
	Ip        *string  `json:"ip,omitempty"`
	Message   string   `json:"message"`
	Receivers []string `json:"receivers"`
}

type MonitorConfig struct {
	Targets         []Target
	AbsenceResetMin int
}

var _monitorConfig *MonitorConfig

func loadMonitorConfig() (MonitorConfig, error) {
	targets, err := loadTargetsFromFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File does not exist, create an empty config file.
			if err := createEmptyTargetsFile(); err != nil {
				return MonitorConfig{}, fmt.Errorf("failed to create empty config file %q: %w", filePath, err)
			} else {
				return MonitorConfig{}, fmt.Errorf("created empty config file %q. please populate it and restart the application", filePath)
			}
		}
		return MonitorConfig{}, fmt.Errorf("failed to open monitor target file %q: %w", filePath, err)
	}

	absenceResetMin := getEnvAsInt("MONITOR_ABSENCE_RESET_MIN", 1440) // Default to 24 hours

	_monitorConfig = &MonitorConfig{
		Targets:         targets,
		AbsenceResetMin: absenceResetMin,
	}

	return *_monitorConfig, nil
}

func GetMonitorConfig() MonitorConfig {
	if _monitorConfig == nil {
		log.Fatal("MonitorConfig not loaded. Call LoadMonitorConfig() first.")
	}
	return *_monitorConfig
}

func loadTargetsFromFile() ([]Target, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	var cfg TargetsFile
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return cfg.Targets, nil
}

func createEmptyTargetsFile() error {
	emptyCfg := TargetsFile{
		Targets: []Target{
			{
				Mac:       "",
				Ip:        nil,
				Message:   "",
				Receivers: []string{},
			},
		},
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
