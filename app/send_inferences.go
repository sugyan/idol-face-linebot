package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/recognizer"
)

func (app *BotApp) sendInferences(userID, replyToken, query string) error {
	token, err := app.retrieveUserToken(userID)
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
			_, err = app.linebot.ReplyMessage(
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
	totalCount := result.Page.TotalCount
	columns := columnsFromInferences(result.Inferences)
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
	_, err = app.linebot.ReplyMessage(replyToken, messages...).Do()
	if err != nil {
		return err
	}
	return nil
}
