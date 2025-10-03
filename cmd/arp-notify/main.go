package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/nekogravitycat/arp-notify/internal/config"
	"github.com/nekogravitycat/arp-notify/internal/linebot"
	"github.com/nekogravitycat/arp-notify/internal/monitor"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	if err := config.LoadConfigs(); err != nil {
		log.Fatalf("Failed to load configs: %v", err)
	}

	monitor.StartPeriodicScan(context.Background())

	linebot.StartLinebotServer()
}
