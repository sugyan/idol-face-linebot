package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
)

func (app *BotApp) callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := app.linebot.ParseRequest(r)
	if err != nil {
		log.Printf("parse request error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, event := range events {
		go func(event *linebot.Event) {
			switch event.Type {
			case linebot.EventTypeFollow:
				token, err := app.retrieveUserToken(event.Source.UserID)
				if err != nil {
					log.Print(err)
				}
				log.Printf("token: %v", token)
			case linebot.EventTypeMessage:
				if err := app.handleMessage(event); err != nil {
					log.Print(err)
				}
			default:
				log.Printf("not message event: %v (source: %v)", event, *event.Source)
			}
		}(event)
	}
}

func (app *BotApp) handleMessage(event *linebot.Event) error {
	switch message := event.Message.(type) {
	case *linebot.ImageMessage:
		log.Printf("image message from %v: %s", event.Source, message.ID)
		if err := app.sendRecognized(message.ID, event.ReplyToken); err != nil {
			return fmt.Errorf("recognize image error: %v", err)
		}
	}
	return nil
}
