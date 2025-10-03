package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

var filePath = "monitor_config.json"

type MonitorTargetsFile struct {
	Targets []MonitorTarget `json:"targets"`
}

type MonitorTarget struct {
	Mac       string   `json:"mac"`
	Message   string   `json:"message"`
	Receivers []string `json:"receivers"`
}

type TargetInfo struct {
	Message   string
	Receivers []string
}

type MonitorConfig struct {
	Targets         map[string]TargetInfo // Keyed by MAC address
	AbsenceResetMin int                   // Minutes after which absence notification resets
}

var _monitorConfig *MonitorConfig

func LoadMonitorConfig() error {
	targets, err := loadTargetsFromFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File does not exist, create an empty config file.
			if err := createEmptyMonitorTargets(); err != nil {
				return fmt.Errorf("failed to create empty config file %q: %w", filePath, err)
			} else {
				return fmt.Errorf("created empty config file %q. please populate it and restart the application", filePath)
			}
		}
		return fmt.Errorf("failed to open monitor target file %q: %w", filePath, err)
	}

	absenceResetMin := getEnvAsInt("MONITOR_ABSENCE_RESET_MIN", 60)

	_monitorConfig = &MonitorConfig{
		Targets:         targets,
		AbsenceResetMin: absenceResetMin,
	}

	return nil
}

func GetMonitorConfig() MonitorConfig {
	if _monitorConfig == nil {
		log.Fatal("MonitorConfig not loaded. Call LoadMonitorConfig() first.")
	}
	return *_monitorConfig
}

func loadTargetsFromFile() (map[string]TargetInfo, error) {
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

	var cfg MonitorTargetsFile
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	targets := make(map[string]TargetInfo)
	for _, entry := range cfg.Targets {
		targets[entry.Mac] = TargetInfo{
			Message:   entry.Message,
			Receivers: entry.Receivers,
		}
	}

	return targets, nil
}

func createEmptyMonitorTargets() error {
	emptyCfg := MonitorTargetsFile{
		Targets: []MonitorTarget{
			{
				Mac:       "",
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
