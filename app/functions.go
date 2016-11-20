package app

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/recognizer"
)

func columnsFromInferences(inferences []recognizer.Inference) []*linebot.CarouselColumn {
	columns := make([]*linebot.CarouselColumn, 0, 5)
	ids := rand.Perm(len(inferences))
	num := 5
	if len(ids) < num {
		num = len(ids)
	}
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
		thumbnailImageURL, _ := url.Parse(os.Getenv("APP_URL") + "/thumbnail")
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
	return columns
}

func columnsFromRecognizedFaces(faces []recognizer.RecognizedFace, key, thumbnailImageURL string) []*linebot.CarouselColumn {
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
