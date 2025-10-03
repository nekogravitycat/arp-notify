package linebot

import (
	"log"
	"net/http"
	"os"
)

// StartLinebotServer initializes and starts the HTTP server for handling LINE bot callbacks and health checks.
func StartLinebotServer() {
	http.HandleFunc("/callback", OnCallback)
	http.HandleFunc("/health", onHealthCheck)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "5000"
	}

	log.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}

func onHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
