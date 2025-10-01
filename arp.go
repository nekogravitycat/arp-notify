package main

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
)

// ValidateIface ensures the interface name is safe (alphanumeric, punctuation allowed limited).
// This prevents command injection via interface argument.
func validateIface(iface string) error {
	// allow common Linux interface names like eth0, eno1, enp0s3, wlan0, eno2, etc.
	// regex: start with letter, then letters/numbers/._- (reasonable and conservative)
	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]{0,15}$`)
	if !re.MatchString(iface) {
		return errors.New("invalid interface name")
	}
	return nil
}

// runArpScan runs arp-scan with a context timeout and returns output and error.
func runArpScan(ctx context.Context, bin string, iface string) (string, error) {
	// Validate interface name to prevent command injection.
	if err := validateIface(iface); err != nil {
		return "", err
	}

	// Construct full path and args (no shell).
	args := []string{"-I", iface, "--localnet", "-x"} // -x makes parsing easier (no header/footer)

	// Create command with context (to enforce timeout/cancellation).
	cmd := exec.CommandContext(ctx, bin, args...)

	// Optionally clear environment or set a minimal env; here we keep inherited env.
	// cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()

	return string(out), err
}
