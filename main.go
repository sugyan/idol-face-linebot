package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/face-manager-linebot/inferences"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}
	app := &app{bot: bot}
	http.HandleFunc(os.Getenv("CALLBACK_PATH"), app.handler)
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}

type app struct {
	bot *linebot.Client
}

func (a *app) handler(w http.ResponseWriter, r *http.Request) {
	events, err := a.bot.ParseRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, event := range events {
		if event.Type != linebot.EventTypeMessage {
			log.Printf("not message event: %v", event)
			continue
		}
		if event.Source.Type != linebot.EventSourceTypeUser {
			log.Printf("not from user: %v", event)
			continue
		}
		if message, ok := event.Message.(*linebot.TextMessage); ok {
			log.Printf("text message: %v", message.Text)
			inferences, err := inferences.BulkFetch(event.Source.UserID)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(inferences) < 1 {
				log.Println("empty inferences")
				continue
			}
			ids := rand.Perm(len(inferences))
			inference := inferences[ids[0]]
			_, err = a.bot.ReplyMessage(
				event.ReplyToken,
				linebot.NewImageMessage(inference.Face.ImageURL, inference.Face.ImageURL),
			).Do()
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
