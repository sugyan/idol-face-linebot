package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/face-manager-linebot/recognizer"
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
	http.HandleFunc("/crop", cropImageHandler)
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
		if event.Source.Type != linebot.EventSourceTypeUser {
			log.Printf("not from user: %v", event)
			continue
		}
		userID := event.Source.UserID
		switch event.Type {
		case linebot.EventTypeFollow:
			token, err := a.retrieveUserToken(userID)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("token: %v", token)
		case linebot.EventTypeMessage:
			if message, ok := event.Message.(*linebot.TextMessage); ok {
				log.Printf("text message from %s: %v", event.Source.UserID, message.Text)
				query := message.Text
				if message.Text == "all" {
					query = ""
				}
				if err := a.sendInferences(event.Source.UserID, event.ReplyToken, query); err != nil {
					log.Printf("send error: %v", err)
				}
			}
		case linebot.EventTypePostback:
			log.Printf("got postback: %s", event.Postback.Data)
			token, err := a.retrieveUserToken(userID)
			if err != nil {
				log.Fatal(err)
			}
			client, err := recognizer.NewClient(userID+"@line.me", token)
			if err != nil {
				log.Fatal(err)
			}
			// <face-id>,<inference-id>
			ids := strings.Split(event.Postback.Data, ",")
			resultURL, err := client.AcceptInference(ids[1])
			if err != nil {
				log.Printf("accept error: %v", err)
				continue
			}
			messageText := fmt.Sprintf("id:%s を更新しました！", ids[0])
			if _, err := a.linebot.ReplyMessage(
				event.ReplyToken,
				linebot.NewTemplateMessage(
					messageText,
					linebot.NewConfirmTemplate(
						messageText,
						linebot.NewMessageTemplateAction("やっぱちがう", "やっぱちがう"),
						linebot.NewURITemplateAction("確認する", resultURL),
					),
				),
			).Do(); err != nil {
				log.Printf("send message error: %v", err)
				continue
			}
		default:
			log.Printf("not message/postback event: %v", event)
			continue
		}
	}
}

func (a *app) sendInferences(userID, replyToken, query string) error {
	token, err := a.retrieveUserToken(userID)
	if err != nil {
		return err
	}
	client, err := recognizer.NewClient(userID+"@line.me", token)
	if err != nil {
		return err
	}

	labelIDs := []int{}
	if query != "" {
		labels, err := client.Labels(query)
		if err != nil {
			return err
		}
		if len(labels) == 0 {
			log.Println("empty labels")
			_, err := a.linebot.ReplyMessage(
				replyToken,
				linebot.NewTextMessage("識別対象のアイドルの名前ではないようです\xf0\x9f\x98\x9e"),
			).Do()
			return err
		}
		for _, label := range labels {
			labelIDs = append(labelIDs, label.ID)
		}
	}
	result, err := client.Inferences(labelIDs)
	if err != nil {
		return err
	}
	inferences := result.Inferences
	totalCount := result.Page.TotalCount
	ids := rand.Perm(len(inferences))
	num := 5
	if len(ids) < num {
		num = len(ids)
	}
	columns := make([]*linebot.CarouselColumn, 0, 5)
	for i := 0; i < num; i++ {
		inference := inferences[ids[i]]
		title := fmt.Sprintf("%d:[%.2f] %s", inference.Face.ID, inference.Score*100.0, inference.Label.Name)
		if inference.Label.Description != "" {
			title += " (" + strings.Replace(inference.Label.Description, "\r\n", ", ", -1) + ")"
		}
		if len([]rune(title)) > 40 {
			title = string([]rune(title)[0:39]) + "…"
		}
		text := strings.Replace(inference.Face.Photo.Caption, "\n", " ", -1)
		if len([]rune(text)) > 60 {
			text = string([]rune(text)[0:59]) + "…"
		}
		thumbnailImageURL, err := url.Parse(os.Getenv("APP_URL") + "/thumbnail")
		if err != nil {
			return err
		}
		values := url.Values{}
		values.Set("image_url", inference.Face.ImageURL)
		thumbnailImageURL.RawQuery = values.Encode()
		columns = append(
			columns,
			linebot.NewCarouselColumn(
				thumbnailImageURL.String(),
				title,
				text,
				linebot.NewURITemplateAction(
					"\xf0\x9f\x94\x8d くわしく",
					inference.Face.Photo.SourceURL,
				),
				linebot.NewPostbackTemplateAction(
					"\xf0\x9f\x99\x86 あってる",
					strings.Join(
						[]string{
							strconv.FormatUint(uint64(inference.Face.ID), 10),
							strconv.FormatUint(uint64(inference.ID), 10),
						},
						",",
					),
					"",
				),
				linebot.NewMessageTemplateAction(
					"\xf0\x9f\x99\x85 ちがうよ", "ちがうよ",
				),
			),
		)
	}
	titles := []string{}
	for _, column := range columns {
		titles = append(titles, column.Title)
	}
	if _, err = a.linebot.ReplyMessage(
		replyToken,
		linebot.NewTextMessage(
			fmt.Sprintf("%d件の候補があります\xf0\x9f\x98\x80", totalCount),
		),
		linebot.NewTemplateMessage(
			strings.Join(titles, "\n"),
			linebot.NewCarouselTemplate(columns...),
		),
	).Do(); err != nil {
		return err
	}
	return nil
}
