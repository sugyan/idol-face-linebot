package main

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

func (a *app) recognizeFaces(key, replyToken string) error {
	appURL := os.Getenv("APP_URL")
	imageURL, err := url.Parse(appURL)
	if err != nil {
		return err
	}
	imageURL.Path = path.Join(imageURL.Path, "image")
	values := url.Values{}
	values.Set("key", key)
	imageURL.RawQuery = values.Encode()
	result, err := a.recognizerAdmin.RecognizeFaces(imageURL.String())
	if err != nil {
		return err
	}
	columns := make([]*linebot.CarouselColumn, 0, 5)
	for _, face := range result.Faces {
		top := face.Recognize[0]
		if !(top.Label.ID > 0 && top.Value > 0.5) {
			continue
		}
		name := top.Label.Name
		if len(top.Label.Description) > 0 {
			name += " (" + strings.Split(top.Label.Description, "\r\n")[0] + ")"
		}
		xMin := math.MaxInt32
		xMax := math.MinInt32
		yMin := math.MaxInt32
		yMax := math.MinInt32
		for _, bounding := range face.Bounding {
			if bounding.X < xMin {
				xMin = bounding.X
			}
			if bounding.X > xMax {
				xMax = bounding.X
			}
			if bounding.Y < yMin {
				yMin = bounding.Y
			}
			if bounding.Y > yMax {
				yMax = bounding.Y
			}
		}
		xSize := float64(xMax-xMin) * 1.2
		ySize := float64(yMax-yMin) * 1.2
		srt := strings.Join([]string{
			fmt.Sprintf("%.2f,%.2f", float64(xMin+xMax)*0.5, float64(yMin+yMax)*0.5),
			"1.0",
			fmt.Sprintf("%.2f", -face.Angle.Roll),
			fmt.Sprintf("%.2f,%.2f", float64(xSize)*0.5, float64(ySize)*0.5),
		}, " ")
		values := url.Values{}
		values.Set("key", key)
		values.Set("srt", srt)
		values.Set("w", strconv.Itoa(int(xSize+0.5)))
		values.Set("h", strconv.Itoa(int(ySize+0.5)))
		thumbnailImageURL, _ := url.Parse(appURL)
		thumbnailImageURL.Path = path.Join(thumbnailImageURL.Path, "image")
		thumbnailImageURL.RawQuery = values.Encode()
		if err != nil {
			return err
		}
		columns = append(columns, linebot.NewCarouselColumn(
			thumbnailImageURL.String(),
			name,
			fmt.Sprintf("%.2f", top.Value*100.0),
			linebot.NewURITemplateAction(
				"@"+top.Label.Twitter,
				"https://twitter.com/"+top.Label.Twitter,
			),
		))
	}
	_, err = a.linebot.ReplyMessage(
		replyToken,
		linebot.NewTemplateMessage(
			"altText",
			linebot.NewCarouselTemplate(columns...),
		),
	).Do()
	if err != nil {
		return err
	}
	return nil
}
