package app

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/jpeg" // for decode
	_ "image/png"  // for decode
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gopkg.in/redis.v5"
)

func (app *BotApp) imageHandler(w http.ResponseWriter, r *http.Request) {
	// return 304 if "If-Modified-Since" header exists.
	// if len(r.Header.Get("If-Modified-Since")) > 0 {
	// 	w.WriteHeader(http.StatusNotModified)
	// 	return
	// }

	var bytes []byte
	bytes, err := app.redis.Get(cacheKey(r.URL)).Bytes()
	if err == redis.Nil {
		bytes, err = app.getImageData(r.URL.Query())
		if err != nil {
			log.Printf("get image error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err = app.redis.Set(cacheKey(r.URL), bytes, time.Hour*24).Err(); err != nil {
			return
		}
	} else if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	// w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
	_, err = w.Write(bytes)
	if err != nil {
		log.Print(err.Error())
	}
}

func (app *BotApp) getImageData(query url.Values) ([]byte, error) {
	// get content image from encrypted key
	key := query.Get("key")
	messageID, err := app.decrypt(key)
	if err != nil {
		return nil, err
	}
	log.Printf("get content %s", messageID)
	res, err := app.linebot.GetMessageContent(messageID).Do()
	if err != nil {
		return nil, err
	}
	defer res.Content.Close()

	// generate thumbnailImage
	target, err := cropTargetFromQuery(query)
	if err != nil {
		return nil, err
	}
	src, _, err := image.Decode(res.Content)
	if err != nil {
		return nil, err
	}
	dst := padForThumbnailImage(rotateAndCropImage(src, target.rect, target.angle))

	buf := bytes.NewBuffer([]byte{})
	if err = jpeg.Encode(buf, dst, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func thumbnailImageHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	// fetch original image
	res, err := http.Get(query.Get("image_url"))
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()
	// decode to image
	face, err := jpeg.Decode(res.Body)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// generate new image, draw
	img := image.NewRGBA(image.Rect(0, 0, 168, 112))
	draw.Draw(img, image.Rect(0, 0, 112, 112).Add(image.Pt(28, 0)), face, image.Pt(0, 0), draw.Src)

	jpeg.Encode(w, img, &jpeg.Options{Quality: 95})
}

func cacheKey(u *url.URL) string {
	return "image:" + u.RawQuery
}
