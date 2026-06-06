package config

import "testing"

func validTarget() Target {
	return Target{
		Name:      "Phone",
		Mac:       "aa:bb:cc:dd:ee:ff",
		Enabled:   true,
		Detection: Detection{Mode: ModeAuto, IP: "192.168.0.2"},
		Receivers: []Receiver{{ID: "U123"}},
	}
}

func TestValidateTargetsConfigValid(t *testing.T) {
	cfg := &TargetsConfig{
		Contacts: []Contact{{ID: "U1", Name: "Mom"}},
		Targets:  []Target{validTarget()},
	}
	if err := validateTargetsConfig(cfg); err != nil {
		t.Fatalf("valid targets rejected: %v", err)
	}
}

func TestValidateTargetsConfigBroadcastNeedsNoIP(t *testing.T) {
	tgt := validTarget()
	tgt.Detection = Detection{Mode: ModeBroadcast}
	cfg := &TargetsConfig{Targets: []Target{tgt}}
	if err := validateTargetsConfig(cfg); err != nil {
		t.Errorf("broadcast target without ip rejected: %v", err)
	}
}

func TestValidateTargetsConfigErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Target)
	}{
		{"empty mac", func(t *Target) { t.Mac = "" }},
		{"bad mac", func(t *Target) { t.Mac = "aa-bb-cc-dd-ee-ff" }},
		{"ip mode without ip", func(t *Target) { t.Detection = Detection{Mode: ModeIP} }},
		{"auto mode without ip", func(t *Target) { t.Detection = Detection{Mode: ModeAuto} }},
		{"invalid ip", func(t *Target) { t.Detection = Detection{Mode: ModeIP, IP: "999.1.1.1"} }},
		{"empty mode", func(t *Target) { t.Detection = Detection{Mode: "", IP: "192.168.0.2"} }},
		{"unknown mode", func(t *Target) { t.Detection = Detection{Mode: "magic", IP: "192.168.0.2"} }},
		{"empty receiver id", func(t *Target) { t.Receivers = []Receiver{{ID: ""}} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tgt := validTarget()
			tt.mutate(&tgt)
			cfg := &TargetsConfig{Targets: []Target{tgt}}
			if err := validateTargetsConfig(cfg); err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestValidateTargetsConfigEmptyContactID(t *testing.T) {
	cfg := &TargetsConfig{
		Contacts: []Contact{{ID: "", Name: "Nobody"}},
		Targets:  []Target{validTarget()},
	}
	if err := validateTargetsConfig(cfg); err == nil {
		t.Error("expected error for empty contact id, got nil")
	}
}

func TestMessageFor(t *testing.T) {
	tgt := Target{Message: "target-msg"}

	if got := tgt.MessageFor(Receiver{Message: "recv-msg"}, "default"); got != "recv-msg" {
		t.Errorf("receiver message should win: got %q", got)
	}
	if got := tgt.MessageFor(Receiver{}, "default"); got != "target-msg" {
		t.Errorf("target message should win over default: got %q", got)
	}

	empty := Target{}
	if got := empty.MessageFor(Receiver{}, "default"); got != "default" {
		t.Errorf("default should be used: got %q", got)
	}
}
