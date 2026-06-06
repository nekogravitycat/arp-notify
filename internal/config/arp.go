package config

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
)

// SystemConfig holds the system / scan behavior loaded from config.yaml.
type SystemConfig struct {
	ArpScan ArpScanConfig `yaml:"arp_scan" json:"arp_scan"`
	Monitor MonitorConfig `yaml:"monitor" json:"monitor"`
	Server  ServerConfig  `yaml:"server" json:"server"`
}

type ArpScanConfig struct {
	Bin                  string `yaml:"bin" json:"bin"`
	Iface                string `yaml:"iface" json:"iface"`
	IntervalSec          int    `yaml:"interval_sec" json:"interval_sec"`
	BroadcastTimeoutSec  int    `yaml:"broadcast_timeout_sec" json:"broadcast_timeout_sec"`
	IndividualTimeoutSec int    `yaml:"individual_timeout_sec" json:"individual_timeout_sec"`
}

type MonitorConfig struct {
	AbsenceResetMin int `yaml:"absence_reset_min" json:"absence_reset_min"`
}

type ServerConfig struct {
	Port int `yaml:"port" json:"port"`
}

// applySystemDefaults fills in sensible defaults for any zero-valued field so
// partially-specified config files still work.
func applySystemDefaults(cfg *SystemConfig) {
	if cfg.ArpScan.Bin == "" {
		cfg.ArpScan.Bin = "arp-scan"
	}
	if cfg.ArpScan.IntervalSec == 0 {
		cfg.ArpScan.IntervalSec = 60
	}
	if cfg.ArpScan.BroadcastTimeoutSec == 0 {
		cfg.ArpScan.BroadcastTimeoutSec = 15
	}
	if cfg.ArpScan.IndividualTimeoutSec == 0 {
		cfg.ArpScan.IndividualTimeoutSec = 2
	}
	if cfg.Monitor.AbsenceResetMin == 0 {
		cfg.Monitor.AbsenceResetMin = 1440 // 24 hours
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 5000
	}
}

func validateSystemConfig(cfg *SystemConfig) error {
	if err := checkBin(cfg.ArpScan.Bin); err != nil {
		return err
	}
	if err := validateIface(cfg.ArpScan.Iface); err != nil {
		return err
	}
	if cfg.ArpScan.IntervalSec <= 0 {
		return errors.New("arp_scan.interval_sec must be > 0")
	}
	if cfg.ArpScan.BroadcastTimeoutSec <= 0 {
		return errors.New("arp_scan.broadcast_timeout_sec must be > 0")
	}
	if cfg.ArpScan.IndividualTimeoutSec <= 0 {
		return errors.New("arp_scan.individual_timeout_sec must be > 0")
	}
	if cfg.Monitor.AbsenceResetMin <= 0 {
		return errors.New("monitor.absence_reset_min must be > 0")
	}
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return errors.New("server.port must be between 1 and 65535")
	}
	return nil
}

// checkBin checks if the arp-scan binary is available in PATH.
func checkBin(bin string) error {
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("binary %q not found in PATH: %w", bin, err)
	}
	return nil
}

// validateIface validates the interface name. An empty name is allowed and
// means "all interfaces".
func validateIface(iface string) error {
	if iface == "" {
		return nil
	}
	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]{0,15}$`)
	if !re.MatchString(iface) {
		return errors.New("invalid interface name")
	}
	return nil
}
