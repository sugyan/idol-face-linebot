package app

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/app/message"
	"github.com/sugyan/idol-face-linebot/recognizer"
)

func (app *BotApp) sendRecognized(key, replyToken string) error {
	appURL := os.Getenv("APP_URL")
	imageURL, err := url.Parse(appURL)
	if err != nil {
		return err
	}
	imageURL.Path = path.Join(imageURL.Path, "image")
	values := url.Values{}
	values.Set("key", key)
	imageURL.RawQuery = values.Encode()
	result, err := app.recognizerAdmin.RecognizeFaces(imageURL.String())
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
		columns := message.FromRecognizedFaces(succeededFaces, key, thumbnailImageURL.String())

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
	_, err = app.linebot.ReplyMessage(replyToken, messages...).Do()
	if err != nil {
		return err
	}
	return nil
}
