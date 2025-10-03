package config

import (
	"fmt"
	"os"
	"strconv"
)

func Load() error {
	if cfg, err := loadArpScanConfig(); err != nil {
		return fmt.Errorf("failed to load arp-scan config: %w", err)
	} else {
		fmt.Printf("Using arp-scan config: %+v\n", cfg)
	}

	if cfg, err := loadMonitorConfig(); err != nil {
		return fmt.Errorf("failed to load monitor config: %w", err)
	} else {
		fmt.Printf("Using monitor config: %+v\n", cfg)
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
