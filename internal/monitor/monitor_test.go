package monitor

import (
	"os/exec"
	"testing"
	"time"

	"github.com/nekogravitycat/arp-notify/internal/config"
)

func TestContainsMac(t *testing.T) {
	// Typical arp-scan -x output: IP\tMAC\tvendor, one host per line.
	output := "192.168.0.2\taa:bb:cc:dd:ee:ff\tAcme Corp\n" +
		"192.168.0.3\t11:22:33:44:55:66\tWidgets Inc\n"

	if !containsMac(output, "aa:bb:cc:dd:ee:ff") {
		t.Error("expected to find aa:bb:cc:dd:ee:ff")
	}
	if !containsMac(output, "AA:BB:CC:DD:EE:FF") {
		t.Error("match should be case-insensitive")
	}
	if !containsMac(output, "11:22:33:44:55:66") {
		t.Error("expected to find second host")
	}
	if containsMac(output, "00:00:00:00:00:00") {
		t.Error("did not expect to find absent MAC")
	}
}

func TestContainsMacNoSubstringFalsePositive(t *testing.T) {
	// The MAC appears only as a substring of a larger token, never as a field.
	output := "192.168.0.2\tffaa:bb:cc:dd:ee:ffff\tVendor\n"
	if containsMac(output, "aa:bb:cc:dd:ee:ff") {
		t.Error("substring inside a larger field should not count as a match")
	}
}

// configureMonitor points config at a fresh temp dir with the given re-notify
// window so updateStateAndShouldNotify reads deterministic settings.
func configureMonitor(t *testing.T, absenceMin int) {
	t.Helper()
	t.Chdir(t.TempDir())

	bin := ""
	for _, c := range []string{"go", "sh", "cmd", "ls", "where"} {
		if _, err := exec.LookPath(c); err == nil {
			bin = c
			break
		}
	}
	if bin == "" {
		t.Skip("no known binary on PATH for config validation")
	}

	cfg := config.SystemConfig{
		ArpScan: config.ArpScanConfig{
			Bin:                  bin,
			IntervalSec:          60,
			BroadcastTimeoutSec:  15,
			IndividualTimeoutSec: 2,
		},
		Monitor: config.MonitorConfig{AbsenceResetMin: absenceMin},
		Server:  config.ServerConfig{Host: "127.0.0.1", Port: 5000},
	}
	if err := config.SaveSystemConfig(cfg); err != nil {
		t.Fatalf("SaveSystemConfig: %v", err)
	}
}

func resetState() {
	stateMu.Lock()
	defer stateMu.Unlock()
	state = make(map[string]deviceState)
}

func TestUpdateStateFirstSightingNotifies(t *testing.T) {
	configureMonitor(t, 1440)
	resetState()

	const mac = "aa:bb:cc:dd:ee:ff"
	if !updateStateAndShouldNotify(mac) {
		t.Error("first sighting should notify")
	}

	stateMu.Lock()
	ds := state[mac]
	stateMu.Unlock()
	if !ds.notified {
		t.Error("first sighting should mark device as notified")
	}
}

func TestUpdateStateSecondSightingSkips(t *testing.T) {
	configureMonitor(t, 1440)
	resetState()

	const mac = "aa:bb:cc:dd:ee:ff"
	updateStateAndShouldNotify(mac) // first: notifies
	if updateStateAndShouldNotify(mac) {
		t.Error("second sighting within the re-notify window should not notify")
	}
}

func TestUpdateStateReNotifiesAfterAbsence(t *testing.T) {
	configureMonitor(t, 60) // re-notify after 60 minutes of absence
	resetState()

	const mac = "aa:bb:cc:dd:ee:ff"

	// Seed a device last seen well beyond the absence window, already notified.
	stateMu.Lock()
	state[mac] = deviceState{lastSeen: time.Now().Add(-2 * time.Hour), notified: true}
	stateMu.Unlock()

	if !updateStateAndShouldNotify(mac) {
		t.Error("device reappearing after the absence window should notify again")
	}
}
