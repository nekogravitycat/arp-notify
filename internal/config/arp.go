package config

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

type ArpScanConfig struct {
	Bin         string
	Iface       string
	IntervalSec int
	TimeoutSec  int
}

var _arpConfig *ArpScanConfig

func LoadArpScanConfig() error {
	_arpConfig = &ArpScanConfig{
		Bin:         getEnv("ARP_SCAN_BIN", "arp-scan"),
		Iface:       getEnv("ARP_SCAN_IFACE", "eno1"),
		IntervalSec: getEnvAsInt("ARP_SCAN_INTERVAL_SECS", 60),
		TimeoutSec:  getEnvAsInt("ARP_SCAN_TIMEOUT_SECS", 15),
	}

	return validateArpScanConfig(_arpConfig)
}

func GetArpScanConfig() ArpScanConfig {
	if _arpConfig == nil {
		log.Fatal("ArpScanConfig not loaded. Call LoadArpScanConfig() first.")
	}
	return *_arpConfig
}

func validateArpScanConfig(config *ArpScanConfig) error {
	if err := checkBin(config.Bin); err != nil {
		return err
	}
	if err := validateIface(config.Iface); err != nil {
		return err
	}
	if config.IntervalSec <= 0 {
		return errors.New("invalid interval (must be > 0)")
	}
	if config.TimeoutSec <= 0 {
		return errors.New("invalid timeout (must be > 0)")
	}
	return nil
}

// CheckBin checks if the arp-scan binary is available in PATH.
func checkBin(bin string) error {
	_, err := exec.LookPath(bin)
	if err != nil {
		return fmt.Errorf("binary %q not found in PATH: %w", bin, err)
	}
	return nil
}

// ValidateIface the interface name (alphanumeric, punctuation allowed limited).
func validateIface(iface string) error {
	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]{0,15}$`)
	if !re.MatchString(iface) {
		return errors.New("invalid interface name")
	}
	return nil
}
