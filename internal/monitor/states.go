package monitor

import (
	"time"

	"github.com/nekogravitycat/arp-notify/internal/config"
)

type deviceState struct {
	lastSeen time.Time
	notified bool
}

var state = make(map[string]deviceState)

func markNotified(mac string) {
	if ds, exists := state[mac]; exists {
		state[mac] = deviceState{
			lastSeen: ds.lastSeen,
			notified: true,
		}
	}
}

// updateStateAndShouldNotify updates the state for the given MAC address
// and returns true if a notification should be sent.
func updateStateAndShouldNotify(mac string) bool {
	cfg := config.GetMonitorConfig()

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
	if time.Since(ds.lastSeen) > time.Duration(cfg.AbsenceResetMin)*time.Minute {
		ds.notified = false
	}

	ds.lastSeen = time.Now()

	state[mac] = ds
	return !ds.notified
}
