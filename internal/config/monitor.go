package config

import (
	"fmt"
	"regexp"
)

// Detection modes.
const (
	ModeIP        = "ip"        // individual-scan the configured IP only
	ModeBroadcast = "broadcast" // broadcast-scan only
	ModeAuto      = "auto"      // individual-scan the IP first, broadcast as fallback
)

// TargetsConfig holds the monitoring targets loaded from targets.yaml.
type TargetsConfig struct {
	DefaultMessage string    `yaml:"default_message" json:"default_message"`
	Contacts       []Contact `yaml:"contacts" json:"contacts"`
	Targets        []Target  `yaml:"targets" json:"targets"`
}

// Contact maps a LINE user ID to a human-friendly name, reusable across targets.
type Contact struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
}

type Target struct {
	Name      string     `yaml:"name" json:"name"`
	Mac       string     `yaml:"mac" json:"mac"`
	Enabled   bool       `yaml:"enabled" json:"enabled"`
	Detection Detection  `yaml:"detection" json:"detection"`
	Message   string     `yaml:"message,omitempty" json:"message"`
	Receivers []Receiver `yaml:"receivers" json:"receivers"`
}

type Detection struct {
	Mode string `yaml:"mode" json:"mode"`
	IP   string `yaml:"ip,omitempty" json:"ip"`
}

type Receiver struct {
	ID      string `yaml:"id" json:"id"`
	Message string `yaml:"message,omitempty" json:"message"`
}

// MessageFor resolves the message a given receiver should get, applying the
// precedence: receiver.Message -> target.Message -> defaultMessage.
func (t Target) MessageFor(r Receiver, defaultMessage string) string {
	if r.Message != "" {
		return r.Message
	}
	if t.Message != "" {
		return t.Message
	}
	return defaultMessage
}

var macRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`)

func validateTargetsConfig(cfg *TargetsConfig) error {
	for i, t := range cfg.Targets {
		label := t.Name
		if label == "" {
			label = fmt.Sprintf("#%d", i+1)
		}

		if t.Mac == "" {
			return fmt.Errorf("target %s: mac is required", label)
		}
		if !macRegex.MatchString(t.Mac) {
			return fmt.Errorf("target %s: invalid mac %q (expected aa:bb:cc:dd:ee:ff)", label, t.Mac)
		}

		switch t.Detection.Mode {
		case ModeIP, ModeAuto:
			if t.Detection.IP == "" {
				return fmt.Errorf("target %s: detection mode %q requires an ip", label, t.Detection.Mode)
			}
		case ModeBroadcast:
			// no IP required
		case "":
			return fmt.Errorf("target %s: detection mode is required (ip|broadcast|auto)", label)
		default:
			return fmt.Errorf("target %s: invalid detection mode %q (expected ip|broadcast|auto)", label, t.Detection.Mode)
		}

		for j, r := range t.Receivers {
			if r.ID == "" {
				return fmt.Errorf("target %s: receiver #%d has an empty id", label, j+1)
			}
		}
	}

	for i, c := range cfg.Contacts {
		if c.ID == "" {
			return fmt.Errorf("contact #%d has an empty id", i+1)
		}
	}

	return nil
}
