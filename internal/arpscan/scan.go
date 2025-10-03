package arpscan

import (
	"context"
	"os/exec"
)

// RunArpScan runs arp-scan with a context timeout and returns output and error.
func RunArpScan(ctx context.Context, bin string, iface string) (string, error) {
	// Construct full path and args (no shell).
	// -q and -x makes parsing easier (no header/footer, minimal output)
	args := []string{"-I", iface, "-l", "-q", "-x"}

	// Create command with context (to enforce timeout/cancellation).
	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.CombinedOutput()

	return string(out), err
}
