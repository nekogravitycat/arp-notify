package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	channelAccessToken := os.Getenv("LINE_BOT_CHANNEL_ACCESS_TOKEN")
	channelSecret := os.Getenv("LINE_BOT_CHANNEL_SECRET")

	bot, err := messaging_api.NewMessagingApiAPI(channelAccessToken)
	if err != nil {
		log.Print(err)
		log.Fatal("Error creating LINE bot client")
	}

	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		cb, err := webhook.ParseRequest(channelSecret, req)
		if err != nil {
			log.Printf("Cannot parse request: %+v\n", err)
			if errors.Is(err, webhook.ErrInvalidSignature) {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		for _, event := range cb.Events {
			switch e := event.(type) {
			case webhook.MessageEvent:
				// Handle message event
				switch message := e.Message.(type) {
				case webhook.TextMessageContent:
					// Handle text message content
					userId := "unknown"
					switch source := e.Source.(type) {
					case webhook.UserSource:
						// Get user ID from user source
						log.Printf("Message from user: %s\n", source.UserId)
						log.Printf("Message content: %s\n", message.Text)
						userId = source.UserId
					default:
						log.Printf("Message not from user: %T\n", source)
					}
					// Reply with the same message and user ID
					_, err := bot.ReplyMessage(
						&messaging_api.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages: []messaging_api.MessageInterface{
								messaging_api.TextMessageV2{
									Text: userId,
								},
								messaging_api.TextMessageV2{
									Text: message.Text,
								},
							},
						},
					)
					if err != nil {
						log.Print(err)
					}
				default:
					log.Printf("Unsupported message content: %T\n", e.Message)
				}
			default:
				log.Printf("Unsupported message: %T\n", event)
			}
		}
	})

	// Endpoint for health check
	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	fmt.Println("http://localhost:" + port + "/")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
