package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/face-manager-linebot/recognizer"
)

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
		accept, _ := json.Marshal(postbackData{
			Action:      postbackActionAccept,
			FaceID:      inference.Face.ID,
			InferenceID: inference.ID,
		})
		reject, _ := json.Marshal(postbackData{
			Action:      postbackActionReject,
			FaceID:      inference.Face.ID,
			InferenceID: inference.ID,
		})
		if err != nil {
			return err
		}
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
				linebot.NewPostbackTemplateAction("\xe2\xad\x95 あってる", string(accept), ""),
				linebot.NewPostbackTemplateAction("\xe2\x9d\x8c ちがうよ", string(reject), ""),
			),
		)
	}
	var messages []linebot.Message
	if len(columns) > 0 {
		altTextLines := []string{}
		for _, column := range columns {
			altTextLines = append(altTextLines, column.Title)
		}
		messages = []linebot.Message{
			linebot.NewTextMessage(
				fmt.Sprintf("%d件の候補があります\xf0\x9f\x98\x80", totalCount),
			),
			linebot.NewTemplateMessage(
				strings.Join(altTextLines, "\n"),
				linebot.NewCarouselTemplate(columns...),
			),
		}
	} else {
		messages = []linebot.Message{
			linebot.NewTextMessage("候補が見つかりませんでした\xf0\x9f\x98\x9e"),
		}
	}
	_, err = a.linebot.ReplyMessage(replyToken, messages...).Do()
	if err != nil {
		return err
	}
	return nil
}
