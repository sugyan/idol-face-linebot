package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/app/message"
	"github.com/sugyan/idol-face-linebot/recognizer"
)

func (app *BotApp) callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := app.linebot.ParseRequest(r)
	if err != nil {
		log.Printf("parse request error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeFollow:
			token, err := app.retrieveUserToken(event.Source.UserID)
			if err != nil {
				log.Print(err)
				continue
			}
			log.Printf("token: %v", token)
		case linebot.EventTypeMessage:
			if err := app.handleMessage(event); err != nil {
				log.Print(err)
				continue
			}
		case linebot.EventTypePostback:
			if err := app.handlePostback(event); err != nil {
				log.Print(err)
				continue
			}
		default:
			log.Printf("not message/postback event: %v", event)
			continue
		}
	}
}

func (app *BotApp) handleMessage(event *linebot.Event) error {
	switch msg := event.Message.(type) {
	case *linebot.TextMessage:
		if event.Source.Type != linebot.EventSourceTypeUser {
			// not from user (ignore)
			return nil
		}
		log.Printf("text message from %s: %s", event.Source.UserID, msg.Text)
		query := msg.Text
		if msg.Text == "all" {
			query = ""
		}
		if err := app.sendInferences(event.Source.UserID, event.ReplyToken, query); err != nil {
			return fmt.Errorf("send error: %v", err)
		}
	case *linebot.ImageMessage:
		log.Printf("image message from %v: %s", event.Source, msg.ID)
		// encrypt message ID and pass URL
		key, err := app.encrypt(msg.ID)
		if err != nil {
			return err
		}
		if err := app.sendRecognized(key, event.ReplyToken); err != nil {
			return fmt.Errorf("recognize image error: %v", err)
		}
	}
	return nil
}

func (app *BotApp) handlePostback(event *linebot.Event) error {
	if event.Source.Type != linebot.EventSourceTypeUser {
		return fmt.Errorf("not from user: %v", event)
	}

	userID := event.Source.UserID
	log.Printf("got postback: %s", event.Postback.Data)
	token, err := app.retrieveUserToken(userID)
	if err != nil {
		return err
	}
	client, err := recognizer.NewClient(userID+"@line.me", token)
	if err != nil {
		return err
	}
	// unmarshal data
	data := &message.PostbackData{}
	if err := json.Unmarshal([]byte(event.Postback.Data), data); err != nil {
		return err
	}
	// accept or reject
	var text string
	switch data.Action {
	case message.PostbackActionAccept:
		if err := client.AcceptInference(data.InferenceID); err != nil {
			log.Printf("accept error: %v", err)
			text = "処理できませんでした\xf0\x9f\x98\x9e"
		} else {
			text = fmt.Sprintf("ID:%d を更新しました \xf0\x9f\x99\x86", data.FaceID)
		}
	case message.PostbackActionReject:
		if err := client.RejectInference(data.InferenceID); err != nil {
			log.Printf("reject error: %v", err)
			text = "処理できませんでした\xf0\x9f\x98\x9e"
		} else {
			text = fmt.Sprintf("ID:%d を更新しました \xf0\x9f\x99\x85", data.FaceID)
		}
	}
	if _, err := app.linebot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(text),
	).Do(); err != nil {
		return fmt.Errorf("send message error: %v", err)
	}
	return nil
}
