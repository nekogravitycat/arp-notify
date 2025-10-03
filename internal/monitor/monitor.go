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
	cfg := config.GetArpScanConfig()

	// Binary semaphore to allow only one scan at a time.
	semaphore := make(chan struct{}, 1)

	ticker := time.NewTicker(time.Duration(cfg.IntervalSec) * time.Second)
	defer ticker.Stop()

	tryRun := func() {
		select {
		case semaphore <- struct{}{}:
			// Acquired semaphore, run scan.
			defer func() { <-semaphore }() // Release semaphore when done.

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSec)*time.Second)
			defer cancel()

			output, err := arpscan.RunArpScan(ctx, cfg.Bin, cfg.Iface)
			if err != nil {
				log.Printf("Error running arp-scan: %v", err)
				return
			}

			log.Printf("arp-scan output:\n%s", output)
			go handleOutput(output)
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

// handleOutput processes the output of boardcast arp-scan and checks for target MAC addresses.
func handleOutput(output string) {
	cfg := config.GetMonitorConfig()

	for mac, info := range cfg.Targets {
		if strings.Contains(output, mac) {
			onFound(mac, info)
		} else {
			onBoardcastNotFound(mac, info)
		}
	}
}

// onFound handles the event when a target MAC is found in a scan.
func onFound(mac string, info config.TargetInfo) {
	log.Printf("Target MAC %s found in scan output.", mac)

	if !updateStateAndShouldNotify(mac) {
		log.Printf("Already notified for MAC %s, skipping notification.", mac)
		return
	}

	log.Printf("Sending notification for MAC %s.", mac)
	// Notify receivers
	sendNotification(info.Receivers, info.Message)
	// Mark as notified
	markNotified(mac)
}

// onBoardcastNotFound performs an individual scan for the target IP if provided.
func onBoardcastNotFound(mac string, info config.TargetInfo) {
	log.Printf("Target MAC %s NOT found in scan output.", mac)

	if info.Ip != nil && *info.Ip != "" {
		log.Printf("Performing individual scan for IP %s of MAC %s.", *info.Ip, mac)
		individualScan(*info.Ip, mac, info)
	}
}

// individualScan performs an individual arp-scan for the given IP and checks for the MAC address.
func individualScan(ip string, mac string, info config.TargetInfo) {
	cfg := config.GetArpScanConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output, err := arpscan.RunArpScanOnIp(ctx, cfg.Bin, cfg.Iface, ip)
	if err != nil {
		log.Printf("Error running individual arp-scan on IP %s: %v", ip, err)
		return
	}

	log.Printf("Individual arp-scan output for IP %s:\n%s", ip, output)

	if strings.Contains(output, mac) {
		onFound(mac, info)
		return
	}
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
