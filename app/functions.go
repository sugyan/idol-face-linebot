package app

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/recognizer"
)

type cropTarget struct {
	rect  image.Rectangle
	angle float64
}

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
		faceImageURL, _ := url.Parse(os.Getenv("APP_URL") + "/face")
		values := url.Values{}
		values.Set("id", strconv.Itoa(inference.Face.ID))
		faceImageURL.RawQuery = values.Encode()
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
				faceImageURL.String(),
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
			name += " (" + strings.Replace(top.Label.Description, "\r\n", ", ", -1) + ")"
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
		values := url.Values{}
		values.Set("key", key)
		values.Set("x_min", strconv.Itoa(xMin))
		values.Set("x_max", strconv.Itoa(xMax))
		values.Set("y_min", strconv.Itoa(yMin))
		values.Set("y_max", strconv.Itoa(yMax))
		values.Set("angle", fmt.Sprintf("%.3f", face.Angle.Roll))
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

func cropTargetFromQuery(query url.Values) (*cropTarget, error) {
	xMinStr := query.Get("x_min")
	xMaxStr := query.Get("x_max")
	yMinStr := query.Get("y_min")
	yMaxStr := query.Get("y_max")
	angleStr := query.Get("angle")
	if xMinStr == "" || xMaxStr == "" || yMinStr == "" || yMaxStr == "" || angleStr == "" {
		return nil, fmt.Errorf("missing parameters")
	}
	xMin, _ := strconv.Atoi(xMinStr)
	xMax, _ := strconv.Atoi(xMaxStr)
	yMin, _ := strconv.Atoi(yMinStr)
	yMax, _ := strconv.Atoi(yMaxStr)
	angle, _ := strconv.ParseFloat(angleStr, 32)
	return &cropTarget{
		rect:  image.Rect(xMin, yMin, xMax, yMax),
		angle: angle,
	}, nil
}
