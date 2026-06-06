package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/nekogravitycat/arp-notify/internal/config"
	"github.com/nekogravitycat/arp-notify/internal/linebot"
	"github.com/nekogravitycat/arp-notify/internal/monitor"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// handleSystem reads (GET) or saves (PUT) the system config.
func handleSystem(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, config.GetSystemConfig())
	case http.MethodPut:
		var cfg config.SystemConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if err := config.SaveSystemConfig(cfg); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, config.GetSystemConfig())
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleTargets reads (GET) or saves (PUT) the targets config.
func handleTargets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, config.GetTargetsConfig())
	case http.MethodPut:
		var cfg config.TargetsConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if err := config.SaveTargetsConfig(cfg); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, config.GetTargetsConfig())
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleContacts reads (GET) or saves (PUT) just the contacts registry.
func handleContacts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, config.GetTargetsConfig().Contacts)
	case http.MethodPut:
		var contacts []config.Contact
		if err := json.NewDecoder(r.Body).Decode(&contacts); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		cfg := config.GetTargetsConfig()
		cfg.Contacts = contacts
		if err := config.SaveTargetsConfig(cfg); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, config.GetTargetsConfig().Contacts)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type statusRow struct {
	Mac      string    `json:"mac"`
	Name     string    `json:"name"`
	LastSeen time.Time `json:"lastSeen"`
	Notified bool      `json:"notified"`
}

// handleStatus returns the live device states joined with target names.
func handleStatus(w http.ResponseWriter, r *http.Request) {
	nameByMac := make(map[string]string)
	for _, t := range config.GetTargetsConfig().Targets {
		nameByMac[strings.ToLower(t.Mac)] = t.Name
	}

	snapshot := monitor.Snapshot()
	rows := make([]statusRow, 0, len(snapshot))
	for _, s := range snapshot {
		rows = append(rows, statusRow{
			Mac:      s.Mac,
			Name:     nameByMac[strings.ToLower(s.Mac)],
			LastSeen: s.LastSeen,
			Notified: s.Notified,
		})
	}
	writeJSON(w, http.StatusOK, rows)
}

// handleSeenUsers returns recently-seen LINE users, preferring the friendly
// name from the contacts registry over the fetched profile name.
func handleSeenUsers(w http.ResponseWriter, r *http.Request) {
	nameByID := make(map[string]string)
	for _, c := range config.GetTargetsConfig().Contacts {
		nameByID[c.ID] = c.Name
	}

	users := linebot.SeenUsers()
	for i := range users {
		if n := nameByID[users[i].ID]; n != "" {
			users[i].Name = n
		}
	}
	writeJSON(w, http.StatusOK, users)
}

// handleTestNotify sends a one-off LINE push so a receiver ID can be verified.
func handleTestNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.ID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	if req.Message == "" {
		req.Message = "arp-notify test notification"
	}

	if err := linebot.SendMessage(req.ID, req.Message); err != nil {
		writeError(w, http.StatusBadGateway, "failed to send: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}
