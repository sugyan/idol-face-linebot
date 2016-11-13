package main

import (
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	app, err := newApp()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc(os.Getenv("CALLBACK_PATH"), app.handler)
	http.HandleFunc("/thumbnail", thumbnailImageHandler)
	http.HandleFunc("/image", app.imageHandler)
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}

func (a *app) handler(w http.ResponseWriter, r *http.Request) {
	events, err := a.linebot.ParseRequest(r)
	if err != nil {
		log.Printf("parse request error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeFollow:
			token, err := a.retrieveUserToken(event.Source.UserID)
			if err != nil {
				log.Print(err)
				continue
			}
			log.Printf("token: %v", token)
		case linebot.EventTypeMessage:
			if err := a.handleMessage(event); err != nil {
				log.Print(err)
				continue
			}
		case linebot.EventTypePostback:
			if err := a.handlePostback(event); err != nil {
				log.Print(err)
				continue
			}
		default:
			log.Printf("not message/postback event: %v", event)
			continue
		}
	}
}
