package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func (a *app) imageHandler(w http.ResponseWriter, r *http.Request) {
	file, err := a.getImageFile(r)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer os.Remove(file.Name())
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	written, err := w.Write(bytes)
	if err != nil {
		log.Print(err.Error())
	}
	log.Printf("sent %d bytes.", written)
}

func (a *app) getImageFile(r *http.Request) (*os.File, error) {
	query := r.URL.Query()
	key := query.Get("key")
	messageID, err := a.decrypt(key)
	if err != nil {
		return nil, err
	}
	log.Printf("get content: %s", messageID)
	res, err := a.linebot.GetMessageContent(messageID).Do()
	if err != nil {
		return nil, err
	}
	// download to tempfile
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())

	written, err := io.Copy(tmp, res.Content)
	if err != nil {
		return nil, err
	}
	if written != res.ContentLength {
		return nil, fmt.Errorf("content lengths mismatch. (%d:%d)", written, res.ContentLength)
	}
	// convert to jpeg file
	file, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	srt := query.Get("srt")
	w := query.Get("w")
	h := query.Get("h")
	var command *exec.Cmd
	if len(srt) > 0 && len(w) > 0 && len(h) > 0 {
		xSize, _ := strconv.Atoi(w)
		ySize, _ := strconv.Atoi(h)
		command = exec.Command(
			"convert",
			"-background", "black",
			"-virtual-pixel", "background",
			"-distort", "SRT", srt,
			"-crop", fmt.Sprintf("%dx%d+0+0", xSize, ySize),
			"-extent", fmt.Sprintf("%dx%d-%d+0", int(float64(xSize)*1.51+0.5), ySize, int(float64(xSize)*0.51*0.5+0.5)),
			tmp.Name(), file.Name(),
		)
	} else {
		command = exec.Command(
			"convert",
			"-resize", "1600x1600>",
			tmp.Name(), file.Name(),
		)
	}
	if err := command.Run(); err != nil {
		return nil, err
	}
	return file, nil
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
