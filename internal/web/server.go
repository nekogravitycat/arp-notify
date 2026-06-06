package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var staticFS embed.FS

// RegisterRoutes mounts the admin UI and its JSON API on the given mux.
func RegisterRoutes(mux *http.ServeMux) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	mux.Handle("/", fileServer)

	mux.HandleFunc("/api/system", handleSystem)
	mux.HandleFunc("/api/targets", handleTargets)
	mux.HandleFunc("/api/contacts", handleContacts)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/seen-users", handleSeenUsers)
	mux.HandleFunc("/api/test-notify", handleTestNotify)
}
