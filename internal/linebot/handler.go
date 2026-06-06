package linebot

import (
	"net/http"
)

// RegisterRoutes registers the LINE webhook and health-check handlers on the
// given mux, so the bot can share one HTTP server with the web UI.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/callback", onCallback)
	mux.HandleFunc("/health", onHealthCheck)
}

func onHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
