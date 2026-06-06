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

// updateStateAndShouldNotify records a sighting of the given MAC and atomically
// decides whether a notification should be sent. When it returns true it has
// already marked the device as notified, so callers need no second step.
func updateStateAndShouldNotify(mac string) bool {
	cfg := config.GetSystemConfig()

	stateMu.Lock()
	defer stateMu.Unlock()

	now := time.Now()

	ds, exists := state[mac]
	if !exists {
		// First sighting: notify and mark as notified in one go.
		state[mac] = deviceState{lastSeen: now, notified: true}
		return true
	}

	// Reset notified status if last seen was long ago.
	if time.Since(ds.lastSeen) > time.Duration(cfg.Monitor.AbsenceResetMin)*time.Minute {
		ds.notified = false
	}

	shouldNotify := !ds.notified
	ds.lastSeen = now
	if shouldNotify {
		ds.notified = true
	}
	state[mac] = ds
	return shouldNotify
}
