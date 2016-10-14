package main

import (
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"net/http"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
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
	img := image.NewRGBA(image.Rect(0, 0, 193, 128))
	draw.Draw(img, image.Rect(0, 0, 112, 112).Add(image.Pt(40, 16)), face, image.Pt(0, 0), draw.Src)
	drawer := font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: inconsolata.Regular8x16,
		Dot: fixed.Point26_6{
			Y: inconsolata.Regular8x16.Metrics().Ascent,
		},
	}
	drawer.DrawString(query.Get("from"))

	jpeg.Encode(w, img, &jpeg.Options{Quality: 95})
}
