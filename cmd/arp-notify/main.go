package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/nekogravitycat/arp-notify/internal/config"
	"github.com/nekogravitycat/arp-notify/internal/linebot"
	"github.com/nekogravitycat/arp-notify/internal/monitor"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.GetMonitorConfig() // Load and validate monitor config early to avoid issues later.

	monitor.StartPeriodicScan(context.Background())

	http.HandleFunc("/callback", linebot.OnCallback)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "5000"
	}

	log.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
