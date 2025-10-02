package arpscan

import (
	"context"
	"log"
	"os/exec"
	"time"
)

func PeriodicScan(ctx context.Context) {
	config := getFromEnv()

	// Validate config before starting.
	if err := ValidateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Binary semaphore to allow only one scan at a time.
	semaphore := make(chan struct{}, 1)

	ticker := time.NewTicker(time.Duration(config.TimeoutSec) * time.Second)
	defer ticker.Stop()

	tryRun := func() {
		select {
		case semaphore <- struct{}{}:
			// Acquired semaphore, run scan.
			defer func() { <-semaphore }() // Release semaphore when done.

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSec)*time.Second)
			defer cancel()

			output, err := runArpScan(ctx, config.Bin, config.Iface)
			if err != nil {
				log.Printf("Error running arp-scan: %v", err)
				return
			}

			log.Printf("arp-scan output:\n%s", output)
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

// runArpScan runs arp-scan with a context timeout and returns output and error.
func runArpScan(ctx context.Context, bin string, iface string) (string, error) {
	if err := validateIface(iface); err != nil {
		return "", err
	}

	// Construct full path and args (no shell).
	// -x makes parsing easier (no header/footer)
	args := []string{"-I", iface, "--localnet", "-x"}

	// Create command with context (to enforce timeout/cancellation).
	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.CombinedOutput()

	return string(out), err
}
