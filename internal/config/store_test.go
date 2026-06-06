package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndGetSystemConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	bin := availableBin(t)

	cfg := validSystemConfig(bin)
	cfg.ArpScan.IntervalSec = 123
	if err := SaveSystemConfig(cfg); err != nil {
		t.Fatalf("SaveSystemConfig: %v", err)
	}

	got := GetSystemConfig()
	if got.ArpScan.IntervalSec != 123 {
		t.Errorf("IntervalSec = %d, want 123", got.ArpScan.IntervalSec)
	}

	// The file should exist and reload to the same values.
	if _, err := os.Stat(systemConfigPath); err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	var reloaded SystemConfig
	if err := loadYAMLFile(systemConfigPath, &reloaded); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.ArpScan.IntervalSec != 123 {
		t.Errorf("reloaded IntervalSec = %d, want 123", reloaded.ArpScan.IntervalSec)
	}
}

func TestSaveSystemConfigSignalsChange(t *testing.T) {
	t.Chdir(t.TempDir())
	bin := availableBin(t)

	// Drain any signal queued by earlier tests so the buffer starts empty.
	select {
	case <-SystemConfigChanged():
	default:
	}

	if err := SaveSystemConfig(validSystemConfig(bin)); err != nil {
		t.Fatalf("SaveSystemConfig: %v", err)
	}

	select {
	case <-SystemConfigChanged():
		// expected
	default:
		t.Error("SaveSystemConfig did not signal SystemConfigChanged")
	}
}

func TestSaveSystemConfigInvalidNotPersisted(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := validSystemConfig("definitely-not-a-real-binary-xyz")
	if err := SaveSystemConfig(cfg); err == nil {
		t.Fatal("expected SaveSystemConfig to reject invalid config")
	}
	if _, err := os.Stat(filepath.Join(".", systemConfigPath)); !os.IsNotExist(err) {
		t.Errorf("invalid config should not have been written, stat err = %v", err)
	}
}

func TestLoadSeedsTargetsTemplateAndAsksToPopulate(t *testing.T) {
	t.Chdir(t.TempDir())
	bin := availableBin(t)

	// Seed a valid system config so Load gets past system validation (the
	// shipped template defaults to bin "arp-scan", which may be absent here).
	if err := SaveSystemConfig(validSystemConfig(bin)); err != nil {
		t.Fatalf("seed system config: %v", err)
	}

	// targets.yaml is missing, so Load should create the template and ask the
	// user to populate it (returning an error).
	if err := Load(); err == nil {
		t.Fatal("expected Load to ask the user to populate targets.yaml")
	}
	if _, err := os.Stat(targetsConfigPath); err != nil {
		t.Errorf("targets config template not created: %v", err)
	}
}
