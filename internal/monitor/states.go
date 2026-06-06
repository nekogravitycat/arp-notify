package monitor

import (
	"sync"
	"time"

	"github.com/nekogravitycat/arp-notify/internal/config"
)

type deviceState struct {
	lastSeen time.Time
	notified bool
}

var (
	stateMu sync.Mutex
	state   = make(map[string]deviceState)
)

// DeviceStatus is a read-only snapshot of a tracked device, exposed to the web UI.
type DeviceStatus struct {
	Mac      string    `json:"mac"`
	LastSeen time.Time `json:"lastSeen"`
	Notified bool      `json:"notified"`
}

// Snapshot returns the current device states for the status view.
func Snapshot() []DeviceStatus {
	stateMu.Lock()
	defer stateMu.Unlock()

	out := make([]DeviceStatus, 0, len(state))
	for mac, ds := range state {
		out = append(out, DeviceStatus{
			Mac:      mac,
			LastSeen: ds.lastSeen,
			Notified: ds.notified,
		})
	}
	return out
}

func markNotified(mac string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if ds, exists := state[mac]; exists {
		ds.notified = true
		state[mac] = ds
	}
}

// updateStateAndShouldNotify updates the state for the given MAC address
// and returns true if a notification should be sent.
func updateStateAndShouldNotify(mac string) bool {
	cfg := config.GetSystemConfig()

	stateMu.Lock()
	defer stateMu.Unlock()

	ds, exists := state[mac]
	if !exists {
		// First time seeing this MAC.
		state[mac] = deviceState{
			lastSeen: time.Now(),
			notified: false,
		}
		return true // Notify on first sighting.
	}

	// Reset notified status if last seen was long ago.
	if time.Since(ds.lastSeen) > time.Duration(cfg.Monitor.AbsenceResetMin)*time.Minute {
		ds.notified = false
	}

	ds.lastSeen = time.Now()
	state[mac] = ds
	return !ds.notified
}
