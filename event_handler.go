package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/face-manager-linebot/recognizer"
)

func (a *app) handleMessage(event *linebot.Event) error {
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		log.Printf("text message from %v: %v", event.Source, message.Text)
		if event.Source.Type != linebot.EventSourceTypeUser {
			return fmt.Errorf("not from user: %v", event)
		}
		query := message.Text
		if message.Text == "all" {
			query = ""
		}
		if err := a.sendInferences(event.Source.UserID, event.ReplyToken, query); err != nil {
			return fmt.Errorf("send error: %v", err)
		}
	case *linebot.ImageMessage:
		log.Printf("image message from %v: %s", event.Source, message.ID)
		imageURL := "" // TODO
		if err := a.recognizeFaces(imageURL, event.ReplyToken); err != nil {
			return fmt.Errorf("recognize image error: %v", err)
		}
	}
	return nil
}

func (a *app) handlePostback(event *linebot.Event) error {
	if event.Source.Type != linebot.EventSourceTypeUser {
		return fmt.Errorf("not from user: %v", event)
	}

	userID := event.Source.UserID
	log.Printf("got postback: %s", event.Postback.Data)
	token, err := a.retrieveUserToken(userID)
	if err != nil {
		return err
	}
	client, err := recognizer.NewClient(userID+"@line.me", token)
	if err != nil {
		return err
	}
	// <face-id>,<inference-id>
	ids := strings.Split(event.Postback.Data, ",")
	if err := client.AcceptInference(ids[1]); err != nil {
		return fmt.Errorf("accept error: %v", err)
	}
	if _, err := a.linebot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(fmt.Sprintf("id:%s を更新しました！ \xf0\x9f\x99\x86", ids[0])),
	).Do(); err != nil {
		return fmt.Errorf("send message error: %v", err)
	}
	return nil
}
