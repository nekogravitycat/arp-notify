package linebot

import (
	"errors"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

// Singleton instance of the LINE bot client
var (
	_bot     *messaging_api.MessagingApiAPI
	_onceBot sync.Once
)

func getBot() *messaging_api.MessagingApiAPI {
	channelAccessToken, ok := os.LookupEnv("LINE_BOT_CHANNEL_ACCESS_TOKEN")
	if !ok {
		log.Fatal("LINE_BOT_CHANNEL_ACCESS_TOKEN not set")
		return nil
	}

	_onceBot.Do(func() {
		var err error
		_bot, err = messaging_api.NewMessagingApiAPI(channelAccessToken)
		if err != nil || _bot == nil {
			log.Fatal("Error creating LINE bot client: ", err)
		}
	})

	return _bot
}

func SendMessage(to string, message string) error {
	bot := getBot()

	_, err := bot.PushMessage(
		&messaging_api.PushMessageRequest{
			To: to,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessageV2{Text: message},
			},
		},
		"", // x-line-retry-key
	)

	return err
}

func OnCallback(writer http.ResponseWriter, request *http.Request) {
	channelSecret, ok := os.LookupEnv("LINE_BOT_CHANNEL_SECRET")
	if !ok {
		log.Fatal("LINE_BOT_CHANNEL_SECRET not set")
		writer.WriteHeader(500)
		return
	}

	callback, err := webhook.ParseRequest(channelSecret, request)
	if err != nil {
		log.Printf("Cannot parse request: %+v\n", err)
		if errors.Is(err, webhook.ErrInvalidSignature) {
			writer.WriteHeader(400)
		} else {
			writer.WriteHeader(500)
		}
		return
	}

	for _, event := range callback.Events {
		switch e := event.(type) {
		case webhook.MessageEvent:
			onMessageEvent(e)
		}
	}
}

// Send user's ID when receiving a text message "whoami" from a user
func onMessageEvent(event webhook.MessageEvent) {
	message, ok := event.Message.(webhook.TextMessageContent)
	if !ok {
		return
	}

	if message.Text != "whoami" {
		return
	}

	source, ok := event.Source.(webhook.UserSource)
	if !ok {
		return
	}

	userId := source.UserId

	_, err := getBot().ReplyMessage(
		&messaging_api.ReplyMessageRequest{
			ReplyToken: event.ReplyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessageV2{Text: userId},
			},
		},
	)
	if err != nil {
		log.Printf("Error replying message to %s: %v\n", userId, err)
	}
}
