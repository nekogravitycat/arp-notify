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

// StartPeriodicScan runs arp-scan periodically.
func StartPeriodicScan(ctx context.Context) {
	arpCfg := config.GetArpScanConfig()
	targets := config.GetMonitorConfig().Targets

	// Binary semaphore to allow only one scan at a time.
	semaphore := make(chan struct{}, 1)

	ticker := time.NewTicker(time.Duration(arpCfg.IntervalSec) * time.Second)
	defer ticker.Stop()

	tryRun := func() {
		select {
		case semaphore <- struct{}{}:
			// Acquired semaphore, run scan.
			defer func() { <-semaphore }() // Release semaphore when done.

			if allTargetsHaveIPs(targets) {
				individualScan(targets)
			} else {
				broadcastScan(targets)
			}

		default:
			// Semaphore not acquired, scan already running.
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
			tryRun()
		}
	}
}

func allTargetsHaveIPs(targets []config.Target) bool {
	for _, target := range targets {
		if target.Ip == nil || *target.Ip == "" {
			return false
		}
	}
	return true
}

// broadcastScan performs a broadcast arp-scan and processes the output.
func broadcastScan(targets []config.Target) {
	// Run broadcast arp-scan
	log.Println("Starting broadcast arp-scan...")

	arpCfg := config.GetArpScanConfig()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(arpCfg.TimeoutSec)*time.Second)
	defer cancel()

	output, err := arpscan.RunArpScan(ctx, arpCfg.Bin, arpCfg.Iface)
	if err != nil {
		log.Printf("Error running arp-scan: %v", err)
		return
	}

	// Handle broadcast output
	notfoundHasIp := []config.Target{}

	for _, info := range targets {
		if strings.Contains(output, info.Mac) {
			// MAC found in broadcast scan
			onFound(info)
		} else if info.Ip != nil && *info.Ip != "" {
			// Schedule individual scan for this target
			log.Printf("MAC %s not found in broadcast scan, scheduling individual scan.", info.Mac)
			notfoundHasIp = append(notfoundHasIp, info)
		}
	}

	// Perform individual scans for targets not found in broadcast scan
	if len(notfoundHasIp) != 0 {
		individualScan(notfoundHasIp)
	}
}

// individualScan performs an individual arp-scan for each target with a specified IP.
func individualScan(targets []config.Target) {
	// Run individual arp-scan for each target with specified IP
	log.Println("Starting individual arp-scan for specified IPs...")

	arpCfg := config.GetArpScanConfig()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(arpCfg.TimeoutSec)*time.Second)
	defer cancel()

	// Iterate over targets and run individual scans
	for _, info := range targets {
		if info.Ip == nil || *info.Ip == "" {
			log.Printf("Skipping MAC %s as no IP is specified.", info.Mac)
			continue
		}

		log.Printf("Running individual scan for MAC %s with IP %v", info.Mac, info.Ip)
		output, err := arpscan.RunArpScanOnIp(ctx, arpCfg.Bin, arpCfg.Iface, *info.Ip)
		if err != nil {
			log.Printf("Error running individual arp-scan for IP %v: %v", *info.Ip, err)
			continue
		}

		// Handle individual scan output
		if strings.Contains(output, info.Mac) {
			onFound(info)
		} else {
			log.Printf("MAC %s not found in individual scan for IP %v", info.Mac, *info.Ip)
		}
	}
}

// onFound handles the event when a target MAC is found in a scan.
func onFound(target config.Target) {
	log.Printf("Target MAC %s found in scan output.", target.Mac)

	if !updateStateAndShouldNotify(target.Mac) {
		log.Printf("Already notified for MAC %s, skipping notification.", target.Mac)
		return
	}

	log.Printf("Sending notification for MAC %s.", target.Mac)

	// Mark as notified
	markNotified(target.Mac)

	// Send notification asynchronously
	go sendNotification(target.Receivers, target.Message)
}

func sendNotification(receivers []string, message string) {
	for _, receiver := range receivers {
		if err := linebot.SendMessage(receiver, message); err != nil {
			log.Printf("Error sending notification to %s: %v", receiver, err)
		} else {
			log.Printf("Notification sent to %s", receiver)
		}
	}
}
