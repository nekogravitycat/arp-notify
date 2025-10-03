package config

import (
	"fmt"
	"os"
	"strconv"
)

func LoadConfigs() error {
	if err := LoadArpScanConfig(); err != nil {
		return fmt.Errorf("failed to load arp-scan config: %w", err)
	}
	if err := LoadMonitorConfig(); err != nil {
		return fmt.Errorf("failed to load monitor config: %w", err)
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	if value, exists := os.LookupEnv(name); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultVal
}
