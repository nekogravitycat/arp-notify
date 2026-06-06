package arpscan

import (
	"context"
	"os/exec"
)

// RunArpScan runs arp-scan with a context timeout and returns output and error.
func RunArpScan(ctx context.Context, bin string, iface string) (string, error) {
	// Construct full path and args (no shell).
	// -q and -x makes parsing easier (no header/footer, minimal output)
	args := []string{"-l", "-q", "-x"}

	if iface != "" {
		args = append(args, "-I", iface)
	}

	// Create command with context (to enforce timeout/cancellation).
	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.CombinedOutput()

	return string(out), err
}

func RunArpScanOnIp(ctx context.Context, bin string, iface string, ip string) (string, error) {
	// Construct full path and args (no shell).
	// -q and -x makes parsing easier (no header/footer, minimal output)
	args := []string{ip, "-q", "-x"}

	if iface != "" {
		args = append(args, "-I", iface)
	}

	// Create command with context (to enforce timeout/cancellation).
	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.CombinedOutput()

	return string(out), err
}
