package app

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/redis.v5"
)

func (app *BotApp) imageHandler(w http.ResponseWriter, r *http.Request) {
	// return 304 if "If-Modified-Since" header exists.
	if len(r.Header.Get("If-Modified-Since")) > 0 {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	var bytes []byte
	bytes, err := app.redis.Get(cacheKey(r.URL)).Bytes()
	if err == redis.Nil {
		bytes, err = app.getImageData(r)
		if err != nil {
			log.Printf("get image error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
	_, err = w.Write(bytes)
	if err != nil {
		log.Print(err.Error())
	}
}

func (app *BotApp) getImageData(r *http.Request) ([]byte, error) {
	query := r.URL.Query()
	key := query.Get("key")
	messageID, err := app.decrypt(key)
	if err != nil {
		return nil, err
	}
	imagePath := filepath.Join(app.imageDir, messageID)
	if _, err := os.Stat(imagePath); err != nil {
		if err := app.downloadContentAsJpeg(messageID, imagePath); err != nil {
			return nil, err
		}
	}
	srt := query.Get("srt")
	w := query.Get("w")
	h := query.Get("h")
	if len(srt) > 0 && len(w) > 0 && len(h) > 0 {
		file, err := ioutil.TempFile("", "")
		if err != nil {
			return nil, err
		}
		defer os.Remove(file.Name())
		xSize, _ := strconv.Atoi(w)
		ySize, _ := strconv.Atoi(h)
		cmd := exec.Command(
			"convert",
			"-background", "black",
			"-virtual-pixel", "background",
			"-distort", "SRT", srt,
			"-crop", fmt.Sprintf("%dx%d+0+0", xSize, ySize),
			"-extent", fmt.Sprintf("%dx%d-%d+0", int(float64(xSize)*1.51+0.5), ySize, int(float64(xSize)*0.51*0.5+0.5)),
			"-resize", "302x200>",
			imagePath, file.Name(),
		)
		log.Print(cmd.Args)
		if err = cmd.Run(); err != nil {
			return nil, err
		}
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		// cache data 10 minutes
		if err := app.redis.Set(cacheKey(r.URL), bytes, time.Minute*10).Err(); err != nil {
			return nil, err
		}
		return bytes, nil
	}
	return ioutil.ReadFile(imagePath)
}

// download to tempfile, and convert (and resize if large) to jpeg
func (app *BotApp) downloadContentAsJpeg(messageID, imagePath string) error {
	log.Printf("get content: %s", messageID)

	res, err := app.linebot.GetMessageContent(messageID).Do()
	if err != nil {
		return err
	}
	defer res.Content.Close()
	file, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	written, err := io.Copy(file, res.Content)
	if err != nil {
		return err
	}
	if written != res.ContentLength {
		return fmt.Errorf("content lengths mismatch. (%d:%d)", written, res.ContentLength)
	}
	if err := exec.Command(
		"convert",
		file.Name(), imagePath,
	).Run(); err != nil {
		return err
	}
	return nil
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
