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

// StartPeriodicScan starts a goroutine that runs arp-scan periodically.
func StartPeriodicScan(ctx context.Context) {
	cfg := config.GetArpScanConfig()

	// Binary semaphore to allow only one scan at a time.
	semaphore := make(chan struct{}, 1)

	ticker := time.NewTicker(time.Duration(cfg.TimeoutSec) * time.Second)
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

func handleOutput(output string) {
	config := config.GetMonitorConfig()

	for mac, info := range config.Targets {
		if !strings.Contains(output, mac) {
			continue
		}

		log.Printf("Target MAC %s found in scan output.", mac)

		if !updateStateAndShouldNotify(mac) {
			log.Printf("Already notified for MAC %s, skipping notification.", mac)
			continue
		}

		log.Printf("Sending notification for MAC %s.", mac)
		// Notify receivers
		sendNotification(info.Receivers, info.Message)

		// Mark as notified
		markNotified(mac)
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
