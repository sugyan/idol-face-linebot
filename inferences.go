package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
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
