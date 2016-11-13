package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/face-manager-linebot/recognizer"
)

type postbackAction string

const (
	postbackActionAccept = "accept"
	postbackActionReject = "reject"
)

type postbackData struct {
	Action      postbackAction `json:"action"`
	FaceID      int            `json:"face_id"`
	InferenceID int            `json:"inference_id"`
}

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
		// encrypt message ID and pass URL
		key, err := a.encrypt(message.ID)
		if err != nil {
			return err
		}
		if err := a.recognizeFaces(key, event.ReplyToken); err != nil {
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
	// unmarshal data
	data := &postbackData{}
	if err := json.Unmarshal([]byte(event.Postback.Data), data); err != nil {
		return err
	}
	// accept or reject
	var message string
	switch data.Action {
	case postbackActionAccept:
		if err := client.AcceptInference(data.InferenceID); err != nil {
			log.Printf("accept error: %v", err)
			message = "処理できませんでした\xf0\x9f\x98\x9e"
		} else {
			message = fmt.Sprintf("ID:%d を更新しました \xf0\x9f\x99\x86", data.FaceID)
		}
	case postbackActionReject:
		if err := client.RejectInference(data.InferenceID); err != nil {
			log.Printf("reject error: %v", err)
			message = "処理できませんでした\xf0\x9f\x98\x9e"
		} else {
			message = fmt.Sprintf("ID:%d を更新しました \xf0\x9f\x99\x85", data.FaceID)
		}
	}
	if _, err := a.linebot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(message),
	).Do(); err != nil {
		return fmt.Errorf("send message error: %v", err)
	}
	return nil
}
