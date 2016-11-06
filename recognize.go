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

func (a *app) recognizeFaces(photoURL, replyToken string) error {
	result, err := a.recognizerAdmin.RecognizeFaces(photoURL)
	if err != nil {
		return err
	}
	columns := make([]*linebot.CarouselColumn, 0, 5)
	for _, face := range result.Faces {
		top := face.Recognize[0]
		if top.Label.ID > 0 && top.Value > 0.5 {
			name := top.Label.Name
			if len(top.Label.Description) > 0 {
				name += " (" + strings.Split(top.Label.Description, "\r\n")[0] + ")"
			}
			text := fmt.Sprintf("%s [%.2f]", name, top.Value*100.0)
			values := url.Values{}
			values.Set("image_url", photoURL)
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
			values.Set("x_min", strconv.Itoa(xMin))
			values.Set("x_max", strconv.Itoa(xMax))
			values.Set("y_min", strconv.Itoa(yMin))
			values.Set("y_max", strconv.Itoa(yMax))
			values.Set("roll_angle", strconv.FormatFloat(face.Angle.Roll, 'f', 8, 64))
			thumbnailImageURL, err := url.Parse(os.Getenv("APP_URL"))
			thumbnailImageURL.Path = path.Join(thumbnailImageURL.Path, "crop")
			thumbnailImageURL.RawQuery = values.Encode()
			if err != nil {
				return err
			}
			columns = append(columns, linebot.NewCarouselColumn(
				thumbnailImageURL.String(),
				"",
				text,
			))
		}
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
