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
	_botErr  error
)

// CheckEnv verifies the required LINE credentials are present. Call it once at
// startup so missing configuration fails fast instead of crashing a request.
func CheckEnv() error {
	if _, ok := os.LookupEnv("LINE_BOT_CHANNEL_ACCESS_TOKEN"); !ok {
		return errors.New("LINE_BOT_CHANNEL_ACCESS_TOKEN not set")
	}
	if _, ok := os.LookupEnv("LINE_BOT_CHANNEL_SECRET"); !ok {
		return errors.New("LINE_BOT_CHANNEL_SECRET not set")
	}
	return nil
}

func getBot() (*messaging_api.MessagingApiAPI, error) {
	channelAccessToken, ok := os.LookupEnv("LINE_BOT_CHANNEL_ACCESS_TOKEN")
	if !ok {
		return nil, errors.New("LINE_BOT_CHANNEL_ACCESS_TOKEN not set")
	}

	_onceBot.Do(func() {
		_bot, _botErr = messaging_api.NewMessagingApiAPI(channelAccessToken)
		if _botErr == nil && _bot == nil {
			_botErr = errors.New("failed to create LINE bot client")
		}
	})

	return _bot, _botErr
}

func SendMessage(to string, message string) error {
	bot, err := getBot()
	if err != nil {
		return err
	}

	_, err = bot.PushMessage(
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

func onCallback(writer http.ResponseWriter, request *http.Request) {
	channelSecret, ok := os.LookupEnv("LINE_BOT_CHANNEL_SECRET")
	if !ok {
		log.Println("LINE_BOT_CHANNEL_SECRET not set")
		writer.WriteHeader(http.StatusInternalServerError)
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

// onMessageEvent records the sender (for the web UI's receiver picker) and,
// when the text is "whoami", replies with the user's ID.
func onMessageEvent(event webhook.MessageEvent) {
	source, ok := event.Source.(webhook.UserSource)
	if !ok {
		return
	}
	userId := source.UserId

	// Remember the user; fetch the LINE profile display name once.
	recordSeenUser(userId, "")
	if seenUserNeedsName(userId) {
		if bot, err := getBot(); err == nil {
			if profile, err := bot.GetProfile(userId); err == nil && profile != nil {
				recordSeenUser(userId, profile.DisplayName)
			}
		}
	}

	message, ok := event.Message.(webhook.TextMessageContent)
	if !ok {
		return
	}
	if message.Text != "whoami" {
		return
	}

	bot, err := getBot()
	if err != nil {
		log.Printf("Cannot reply to %s: %v\n", userId, err)
		return
	}
	_, err = bot.ReplyMessage(
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
