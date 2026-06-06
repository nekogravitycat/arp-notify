package config

import (
	"os/exec"
	"testing"
)

// availableBin returns a binary that is guaranteed to be on PATH so that
// checkBin-dependent validation succeeds, or skips the test if none is found.
func availableBin(t *testing.T) string {
	t.Helper()
	for _, c := range []string{"go", "sh", "cmd", "ls", "where"} {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
	}
	t.Skip("no known binary on PATH for checkBin")
	return ""
}

func TestApplySystemDefaults(t *testing.T) {
	var cfg SystemConfig
	applySystemDefaults(&cfg)

	if cfg.ArpScan.Bin != "arp-scan" {
		t.Errorf("Bin = %q, want arp-scan", cfg.ArpScan.Bin)
	}
	if cfg.ArpScan.IntervalSec != 60 {
		t.Errorf("IntervalSec = %d, want 60", cfg.ArpScan.IntervalSec)
	}
	if cfg.ArpScan.BroadcastTimeoutSec != 15 {
		t.Errorf("BroadcastTimeoutSec = %d, want 15", cfg.ArpScan.BroadcastTimeoutSec)
	}
	if cfg.ArpScan.IndividualTimeoutSec != 2 {
		t.Errorf("IndividualTimeoutSec = %d, want 2", cfg.ArpScan.IndividualTimeoutSec)
	}
	if cfg.Monitor.AbsenceResetMin != 1440 {
		t.Errorf("AbsenceResetMin = %d, want 1440", cfg.Monitor.AbsenceResetMin)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Server.Port != 5000 {
		t.Errorf("Port = %d, want 5000", cfg.Server.Port)
	}
}

func TestApplySystemDefaultsKeepsExisting(t *testing.T) {
	cfg := SystemConfig{
		ArpScan: ArpScanConfig{Bin: "custom", IntervalSec: 5},
		Server:  ServerConfig{Host: "0.0.0.0", Port: 8080},
	}
	applySystemDefaults(&cfg)

	if cfg.ArpScan.Bin != "custom" {
		t.Errorf("Bin overwritten: %q", cfg.ArpScan.Bin)
	}
	if cfg.ArpScan.IntervalSec != 5 {
		t.Errorf("IntervalSec overwritten: %d", cfg.ArpScan.IntervalSec)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Host overwritten: %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Port overwritten: %d", cfg.Server.Port)
	}
}

func validSystemConfig(bin string) SystemConfig {
	return SystemConfig{
		ArpScan: ArpScanConfig{
			Bin:                  bin,
			IntervalSec:          60,
			BroadcastTimeoutSec:  15,
			IndividualTimeoutSec: 2,
		},
		Monitor: MonitorConfig{AbsenceResetMin: 1440},
		Server:  ServerConfig{Host: "127.0.0.1", Port: 5000},
	}
}

func TestValidateSystemConfig(t *testing.T) {
	bin := availableBin(t)

	if err := validateSystemConfig(ptr(validSystemConfig(bin))); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*SystemConfig)
	}{
		{"missing bin", func(c *SystemConfig) { c.ArpScan.Bin = "definitely-not-a-real-binary-xyz" }},
		{"bad iface", func(c *SystemConfig) { c.ArpScan.Iface = "bad iface!" }},
		{"zero interval", func(c *SystemConfig) { c.ArpScan.IntervalSec = 0 }},
		{"zero broadcast timeout", func(c *SystemConfig) { c.ArpScan.BroadcastTimeoutSec = 0 }},
		{"zero individual timeout", func(c *SystemConfig) { c.ArpScan.IndividualTimeoutSec = 0 }},
		{"zero absence", func(c *SystemConfig) { c.Monitor.AbsenceResetMin = 0 }},
		{"bad host", func(c *SystemConfig) { c.Server.Host = "not-an-ip" }},
		{"empty host", func(c *SystemConfig) { c.Server.Host = "" }},
		{"port zero", func(c *SystemConfig) { c.Server.Port = 0 }},
		{"port too large", func(c *SystemConfig) { c.Server.Port = 70000 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validSystemConfig(bin)
			tt.mutate(&cfg)
			if err := validateSystemConfig(&cfg); err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestValidateIface(t *testing.T) {
	valid := []string{"", "eth0", "wlan0", "en0", "br-lan", "eth0.100"}
	for _, s := range valid {
		if err := validateIface(s); err != nil {
			t.Errorf("validateIface(%q) = %v, want nil", s, err)
		}
	}
	invalid := []string{"0eth", "bad iface", "has/slash", "way-too-long-interface-name"}
	for _, s := range invalid {
		if err := validateIface(s); err == nil {
			t.Errorf("validateIface(%q) = nil, want error", s)
		}
	}
}

func ptr[T any](v T) *T { return &v }
