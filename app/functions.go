package app

import (
	"fmt"
	"image"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/recognizer"
)

type cropTarget struct {
	rect  image.Rectangle
	angle float64
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
			linebot.NewURIAction(
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
