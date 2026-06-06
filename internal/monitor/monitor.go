package monitor

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/nekogravitycat/arp-notify/internal/arpscan"
	"github.com/nekogravitycat/arp-notify/internal/config"
	"github.com/nekogravitycat/arp-notify/internal/linebot"
)

// StartPeriodicScan runs arp-scan periodically. Config (targets and interval)
// is re-read on every cycle, so edits made through the web UI take effect
// without a restart.
func StartPeriodicScan(ctx context.Context) {
	interval := config.GetSystemConfig().ArpScan.IntervalSec

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Binary semaphore to allow only one scan at a time.
	semaphore := make(chan struct{}, 1)

	tryRun := func() {
		select {
		case semaphore <- struct{}{}:
			defer func() { <-semaphore }()
			runScanCycle()
		default:
			log.Println("Scan already in progress, skipping this interval.")
		}
	}

	// Initial run.
	tryRun()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping periodic scan due to context cancellation.")
			return
		case <-ticker.C:
			// Hot-reload the interval if it changed.
			if newInterval := config.GetSystemConfig().ArpScan.IntervalSec; newInterval > 0 && newInterval != interval {
				interval = newInterval
				ticker.Reset(time.Duration(interval) * time.Second)
				log.Printf("Scan interval updated to %d seconds.", interval)
			}
			tryRun()
		}
	}
}

// runScanCycle performs one detection pass honoring each target's detection mode.
// At most one broadcast scan runs per cycle.
func runScanCycle() {
	arpCfg := config.GetSystemConfig().ArpScan
	targetsCfg := config.GetTargetsConfig()

	// Collect enabled targets.
	active := make([]config.Target, 0, len(targetsCfg.Targets))
	for _, t := range targetsCfg.Targets {
		if t.Enabled {
			active = append(active, t)
		}
	}
	if len(active) == 0 {
		return
	}

	found := make(map[string]bool) // mac -> found this cycle

	// 1. Individual pass for ip / auto targets.
	for _, t := range active {
		if t.Detection.Mode != config.ModeIP && t.Detection.Mode != config.ModeAuto {
			continue
		}
		if t.Detection.IP == "" {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(arpCfg.IndividualTimeoutSec)*time.Second)
		log.Printf("Individual scan for %q (MAC %s, IP %s)", t.Name, t.Mac, t.Detection.IP)
		output, err := arpscan.RunArpScanOnIp(ctx, arpCfg.Bin, arpCfg.Iface, t.Detection.IP)
		cancel()
		if err != nil {
			log.Printf("Error running individual arp-scan for IP %s: %v", t.Detection.IP, err)
			continue
		}

		if containsMac(output, t.Mac) {
			found[t.Mac] = true
			onFound(t, targetsCfg.DefaultMessage)
		}
	}

	// 2. Broadcast pass for broadcast targets + auto targets not yet found.
	needBroadcast := make([]config.Target, 0)
	for _, t := range active {
		switch t.Detection.Mode {
		case config.ModeBroadcast:
			needBroadcast = append(needBroadcast, t)
		case config.ModeAuto:
			if !found[t.Mac] {
				needBroadcast = append(needBroadcast, t)
			}
		}
	}
	if len(needBroadcast) == 0 {
		return
	}

	log.Println("Starting broadcast arp-scan...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(arpCfg.BroadcastTimeoutSec)*time.Second)
	output, err := arpscan.RunArpScan(ctx, arpCfg.Bin, arpCfg.Iface)
	cancel()
	if err != nil {
		log.Printf("Error running broadcast arp-scan: %v", err)
		return
	}

	for _, t := range needBroadcast {
		if found[t.Mac] {
			continue
		}
		if containsMac(output, t.Mac) {
			found[t.Mac] = true
			onFound(t, targetsCfg.DefaultMessage)
		} else {
			log.Printf("MAC %s (%q) not found.", t.Mac, t.Name)
		}
	}
}

// containsMac reports whether the scan output contains the MAC (case-insensitive).
func containsMac(output, mac string) bool {
	return strings.Contains(strings.ToLower(output), strings.ToLower(mac))
}

// onFound handles the event when a target MAC is found in a scan.
func onFound(target config.Target, defaultMessage string) {
	log.Printf("Target %q (MAC %s) found in scan output.", target.Name, target.Mac)

	if !updateStateAndShouldNotify(target.Mac) {
		log.Printf("Already notified for MAC %s, skipping notification.", target.Mac)
		return
	}

	log.Printf("Sending notification for MAC %s.", target.Mac)
	markNotified(target.Mac)

	// Send each receiver its resolved message asynchronously.
	for _, r := range target.Receivers {
		go sendNotification(r.ID, target.MessageFor(r, defaultMessage))
	}
}

func sendNotification(receiverID, message string) {
	if err := linebot.SendMessage(receiverID, message); err != nil {
		log.Printf("Error sending notification to %s: %v", receiverID, err)
	} else {
		log.Printf("Notification sent to %s", receiverID)
	}
}
