package arpscan

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

type ArpScanConfig struct {
	Bin         string
	Iface       string
	IntervalSec int
	TimeoutSec  int
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

func getFromEnv() ArpScanConfig {
	bin := getEnv("ARP_SCAN_BIN", "arp-scan")
	iface := getEnv("ARP_SCAN_IFACE", "eno1")
	interval := getEnvAsInt("ARP_SCAN_INTERVAL_SECS", 60)
	timeout := getEnvAsInt("ARP_SCAN_TIMEOUT_SECS", 15)

	return ArpScanConfig{
		Bin:         bin,
		Iface:       iface,
		IntervalSec: interval,
		TimeoutSec:  timeout,
	}
}

func ValidateConfig(config ArpScanConfig) error {
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
