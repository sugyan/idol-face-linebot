package app

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/recognizer"
)

func (app *BotApp) sendRecognized(messageID, replyToken string) error {
	// POST image data to recognition API
	res, err := app.linebot.GetMessageContent(messageID).Do()
	if err != nil {
		return err
	}
	defer res.Content.Close()
	data, err := ioutil.ReadAll(res.Content)
	if err != nil {
		return err
	}
	srcImage, format, err := image.Decode(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	log.Printf("%s image (%v)", format, srcImage.Bounds().Size())
	result, err := app.recognizerAdmin.RecognizeFaces(res.ContentType, data)
	if err != nil {
		return err
	}
	// check results, extract succeeded high scored results
	succeeded := make([]recognizer.RecognizedFace, 0, 10)
	sort.Sort(recognizer.ByTopValue(result.Faces))
	for _, face := range result.Faces {
		top := face.Recognize[0]
		if top.Label.ID > 0 && top.Value > RecognizedScoreThreshold {
			succeeded = append(succeeded, face)
		}
	}
	log.Printf("result: %s (succeeded: %d)", result.Message, len(succeeded))

	// encrypt message ID and pass URL
	key, err := app.encrypt(messageID)
	if err != nil {
		return err
	}
	var messages []linebot.Message
	if len(succeeded) > 0 {
		// success
		thumbnailImageURL, err := url.Parse(app.baseURL)
		if err != nil {
			return err
		}
		thumbnailImageURL.Path = path.Join(thumbnailImageURL.Path, "image")

		messages = make([]linebot.Message, 0)
		for i := 0; i < len(succeeded); i += 5 {
			j := i + 5
			if j > len(succeeded) {
				j = len(succeeded)
			}
			faces := succeeded[i:j]
			columns := columnsFromRecognizedFaces(faces, key, thumbnailImageURL.String())
			altTextLines := make([]string, 0, 5)
			for _, column := range columns {
				altTextLines = append(altTextLines, fmt.Sprintf("%s [%s]", column.Title, column.Text))
				if i == 0 {
					// create cache
					parsed, err := url.Parse(column.ThumbnailImageURL)
					if err != nil {
						return err
					}
					target, err := cropTargetFromQuery(parsed.Query())
					if err != nil {
						return err
					}
					dstImage := padForThumbnailImage(rotateAndCropImage(srcImage, target.rect, target.angle))
					buf := bytes.NewBuffer([]byte{})
					if err = jpeg.Encode(buf, dstImage, &jpeg.Options{Quality: 95}); err != nil {
						return err
					}
					if err = app.redis.Set(cacheKey(parsed), buf.Bytes(), time.Hour*24).Err(); err != nil {
						return err
					}
				}
			}
			messages = append(messages, linebot.NewTemplateMessage(
				strings.Join(altTextLines, "\n"),
				linebot.NewCarouselTemplate(columns...),
			))
		}
		text := fmt.Sprintf("%d件の顔を識別しました\xf0\x9f\x98\x80", len(succeeded))
		if len(result.Faces) > len(succeeded) {
			text = fmt.Sprintf("%d件中 %s", len(result.Faces), text)
		}
		messages = append([]linebot.Message{linebot.NewTextMessage(text)}, messages...)
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
