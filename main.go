package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

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
		log.Printf("parse request error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, event := range events {
		if event.Source.Type != linebot.EventSourceTypeUser {
			log.Printf("not from user: %v", event)
			continue
		}
		switch event.Type {
		case linebot.EventTypeMessage:
			if message, ok := event.Message.(*linebot.TextMessage); ok {
				log.Printf("text message from %s: %v", event.Source.UserID, message.Text)
				if err := a.sendCarousel(event.Source.UserID, event.ReplyToken); err != nil {
					log.Printf("send error: %v", err)
				}
			}
		case linebot.EventTypePostback:
			log.Printf("got postback: %s", event.Postback.Data)
			if err := inferences.Accept(event.Source.UserID, event.Postback.Data); err != nil {
				log.Printf("accept error: %v", err)
				continue
			}
			if err := a.bot.ReplyMessage(
				event.ReplyToken,
				linebot.NewTextMessage("更新しました！"),
			); err != nil {
				log.Printf("send message error: %v", err)
				continue
			}
		default:
			log.Printf("not message or postback event: %v", event)
			continue
		}
	}
}

func (a *app) sendCarousel(userID, replyToken string) error {
	inferences, err := inferences.BulkFetch(userID)
	if err != nil {
		return err
	}
	if len(inferences) < 1 {
		return errors.New("empty inferences")
	}
	ids := rand.Perm(len(inferences))
	log.Printf("%d, %v", len(ids), ids)
	num := 5
	if len(ids) < num {
		num = len(ids)
	}
	columns := make([]*linebot.CarouselColumn, 0, 5)
	for i := 0; i < num; i++ {
		inference := inferences[ids[i]]
		name := inference.Label.Name
		if inference.Label.Description != "" {
			name += " (" + inference.Label.Description + ")"
		}
		columns = append(
			columns,
			linebot.NewCarouselColumn(
				inference.Face.ImageURL,
				fmt.Sprintf("id: %d [%.4f]", inference.Face.ID, inference.Score),
				name,
				linebot.NewURITemplateAction(
					"くわしく",
					inference.Face.Photo.SourceURL,
				),
				linebot.NewPostbackTemplateAction(
					"あってる",
					strconv.FormatUint(uint64(inference.ID), 10),
					"",
				),
			),
		)
	}
	if _, err = a.bot.ReplyMessage(
		replyToken,
		linebot.NewTemplateMessage("template message", linebot.NewCarouselTemplate(columns...)),
	).Do(); err != nil {
		return err
	}
	return nil
}
