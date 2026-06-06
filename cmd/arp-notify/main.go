package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/nekogravitycat/arp-notify/internal/config"
	"github.com/nekogravitycat/arp-notify/internal/linebot"
	"github.com/nekogravitycat/arp-notify/internal/monitor"
	"github.com/nekogravitycat/arp-notify/internal/web"
)

func main() {
	// .env is optional: secrets may also be provided directly by the environment
	// (e.g. via systemd), so a missing file is only a warning.
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file loaded (%v); relying on the environment.", err)
	}

	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configs: %v", err)
	}

	if err := linebot.CheckEnv(); err != nil {
		log.Fatalf("LINE bot configuration error: %v", err)
	}

	go monitor.StartPeriodicScan(context.Background())

	mux := http.NewServeMux()
	linebot.RegisterRoutes(mux)
	web.RegisterRoutes(mux)

	srvCfg := config.GetSystemConfig().Server
	addr := net.JoinHostPort(srvCfg.Host, fmt.Sprintf("%d", srvCfg.Port))
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Printf("Starting server on %s (admin UI at /admin/)", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
