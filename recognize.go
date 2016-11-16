package main

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/recognizer"
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
	// check results, extract succeeded high scored results
	success := 0
	succeededFaces := make([]recognizer.RecognizedFace, 0, 5)
	sort.Sort(recognizer.ByTopValue(result.Faces))
	for _, face := range result.Faces {
		top := face.Recognize[0]
		if !(top.Label.ID > 0 && top.Value > 0.5) {
			continue
		}
		success++
		if len(succeededFaces) >= 5 {
			continue
		}
		succeededFaces = append(succeededFaces, face)
	}

	var messages []linebot.Message
	if success > 0 {
		// success
		thumbnailImageURL, err := url.Parse(appURL)
		if err != nil {
			return err
		}
		thumbnailImageURL.Path = path.Join(thumbnailImageURL.Path, "image")
		columns := makeRecognizedCarousel(succeededFaces, key, thumbnailImageURL.String())

		text := fmt.Sprintf("%d件の顔を識別しました\xf0\x9f\x98\x80", success)
		if len(result.Faces) > len(columns) {
			text = fmt.Sprintf("%d件中 %s", len(result.Faces), text)
		}
		altTextLines := []string{}
		for _, column := range columns {
			altTextLines = append(altTextLines, fmt.Sprintf("%s [%s]", column.Title, column.Text))
		}
		messages = []linebot.Message{
			linebot.NewTextMessage(text),
			linebot.NewTemplateMessage(
				strings.Join(altTextLines, "\n"),
				linebot.NewCarouselTemplate(columns...),
			),
		}
	} else {
		// failure
		var text string
		if len(result.Faces) > 0 {
			text = fmt.Sprintf("%d件の顔を検出しましたが、識別対象の人物ではなさそうです", len(result.Faces))
		} else {
			text = "顔を検出できませんでした"
		}
		messages = []linebot.Message{
			linebot.NewTextMessage(text + "\xf0\x9f\x98\x9e"),
		}
	}
	_, err = a.linebot.ReplyMessage(replyToken, messages...).Do()
	if err != nil {
		return err
	}
	return nil
}

func makeRecognizedCarousel(faces []recognizer.RecognizedFace, key, thumbnailImageURL string) []*linebot.CarouselColumn {
	columns := make([]*linebot.CarouselColumn, 0, 5)
	for _, face := range faces {
		top := face.Recognize[0]
		name := top.Label.Name
		if len(top.Label.Description) > 0 {
			name += " (" + strings.Split(top.Label.Description, "\r\n")[0] + ")"
		}
		// thumbnailImageURL query parameters
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
		columns = append(columns, linebot.NewCarouselColumn(
			thumbnailImageURL+"?"+values.Encode(),
			name,
			fmt.Sprintf("%.2f", top.Value*100.0),
			linebot.NewURITemplateAction(
				"@"+top.Label.Twitter,
				"https://twitter.com/"+top.Label.Twitter,
			),
		))
	}
	return columns
}
