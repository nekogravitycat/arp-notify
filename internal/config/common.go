package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	systemConfigPath  = "config.yaml"
	targetsConfigPath = "targets.yaml"
)

var (
	mu        sync.RWMutex
	systemCfg SystemConfig
	targetCfg TargetsConfig
)

// Load reads config.yaml and targets.yaml into the in-memory store. Missing
// files are seeded with commented templates; a freshly-created targets.yaml
// returns an error asking the user to populate it and restart.
func Load() error {
	// System config: seed defaults if missing, then load.
	if _, err := os.Stat(systemConfigPath); errors.Is(err, os.ErrNotExist) {
		if err := writeFileAtomic(systemConfigPath, []byte(systemConfigTemplate)); err != nil {
			return fmt.Errorf("failed to create %q: %w", systemConfigPath, err)
		}
		log.Printf("Created default config file %q.", systemConfigPath)
	}

	var sys SystemConfig
	if err := loadYAMLFile(systemConfigPath, &sys); err != nil {
		return fmt.Errorf("failed to load %q: %w", systemConfigPath, err)
	}
	applySystemDefaults(&sys)
	if err := validateSystemConfig(&sys); err != nil {
		return fmt.Errorf("invalid %q: %w", systemConfigPath, err)
	}

	// Targets config: seed template if missing and ask the user to populate it.
	if _, err := os.Stat(targetsConfigPath); errors.Is(err, os.ErrNotExist) {
		if err := writeFileAtomic(targetsConfigPath, []byte(targetsConfigTemplate)); err != nil {
			return fmt.Errorf("failed to create %q: %w", targetsConfigPath, err)
		}
		return fmt.Errorf("created empty config file %q. please populate it and restart the application", targetsConfigPath)
	}

	var targets TargetsConfig
	if err := loadYAMLFile(targetsConfigPath, &targets); err != nil {
		return fmt.Errorf("failed to load %q: %w", targetsConfigPath, err)
	}
	if err := validateTargetsConfig(&targets); err != nil {
		return fmt.Errorf("invalid %q: %w", targetsConfigPath, err)
	}

	mu.Lock()
	systemCfg = sys
	targetCfg = targets
	mu.Unlock()

	fmt.Printf("Using system config: %+v\n", sys)
	fmt.Printf("Using targets config: %+v\n", targets)
	return nil
}

// GetSystemConfig returns a copy of the current system config.
func GetSystemConfig() SystemConfig {
	mu.RLock()
	defer mu.RUnlock()
	return systemCfg
}

// GetTargetsConfig returns a copy of the current targets config.
func GetTargetsConfig() TargetsConfig {
	mu.RLock()
	defer mu.RUnlock()
	return targetCfg
}

// SaveSystemConfig validates, persists, and hot-swaps the system config.
func SaveSystemConfig(cfg SystemConfig) error {
	applySystemDefaults(&cfg)
	if err := validateSystemConfig(&cfg); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal system config: %w", err)
	}
	if err := writeFileAtomic(systemConfigPath, data); err != nil {
		return err
	}
	mu.Lock()
	systemCfg = cfg
	mu.Unlock()
	return nil
}

// SaveTargetsConfig validates, persists, and hot-swaps the targets config.
func SaveTargetsConfig(cfg TargetsConfig) error {
	if err := validateTargetsConfig(&cfg); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal targets config: %w", err)
	}
	if err := writeFileAtomic(targetsConfigPath, data); err != nil {
		return err
	}
	mu.Lock()
	targetCfg = cfg
	mu.Unlock()
	return nil
}

func loadYAMLFile(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}
	return nil
}

// writeFileAtomic writes data to a temp file in the same directory and renames
// it over the destination, so readers never observe a half-written file.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to replace %q: %w", path, err)
	}
	return nil
}

const systemConfigTemplate = `# arp-notify system configuration.
arp_scan:
  bin: arp-scan            # path to the arp-scan binary
  iface: ""               # network interface; empty = all interfaces
  interval_sec: 60         # how often to scan
  broadcast_timeout_sec: 15
  individual_timeout_sec: 2
monitor:
  absence_reset_min: 1440  # re-notify after a device has been absent this long (minutes)
server:
  port: 5000               # HTTP port for the LINE webhook and the /admin UI
`

const targetsConfigTemplate = `# arp-notify monitoring targets.
default_message: "Welcome home!"

# Reusable LINE user -> friendly name registry. Names are auto-filled from the
# LINE profile when a user messages the bot; you can also edit them here.
contacts: []
  # - id: "Uxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  #   name: "Mom"

targets:
  - name: "Example device"
    mac: "aa:bb:cc:dd:ee:ff"
    enabled: false
    detection:
      mode: auto            # ip | broadcast | auto
      ip: "192.168.0.100"   # required for ip / auto
    message: ""             # optional; overrides default_message for this device
    receivers:
      - id: "Uxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
        message: ""         # optional; overrides this device's message
`
