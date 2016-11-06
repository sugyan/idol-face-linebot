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

func cropImageHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	// use tempfile for sending data
	file, err := ioutil.TempFile("", "")
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer os.Remove(file.Name())

	// convert photoURL to tempfile
	xMin, _ := strconv.Atoi(query.Get("x_min"))
	xMax, _ := strconv.Atoi(query.Get("x_max"))
	yMin, _ := strconv.Atoi(query.Get("y_min"))
	yMax, _ := strconv.Atoi(query.Get("y_max"))
	rollAngle, _ := strconv.ParseFloat(query.Get("roll_angle"), 64)
	xSize := float64(xMax-xMin) * 1.2
	ySize := float64(yMax-yMin) * 1.2
	srt := strings.Join([]string{
		fmt.Sprintf("%f,%f", float64(xMin+xMax)*0.5, float64(yMin+yMax)*0.5),
		"1.0",
		fmt.Sprintf("%f", -rollAngle),
		fmt.Sprintf("%f,%f", float64(xSize)*0.5, float64(ySize)*0.5),
	}, " ")
	cmd := exec.Command(
		"convert", query.Get("image_url"),
		"-background", "black",
		"-virtual-pixel", "background",
		"-distort", "SRT", srt,
		"-crop", fmt.Sprintf("%fx%f+0+0", xSize, ySize),
		"-extent", fmt.Sprintf("%fx%f-%f+0", ySize*1.51, ySize, ySize*0.51*0.5),
		file.Name(),
	)
	if err := cmd.Run(); err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.Copy(w, file)
}
