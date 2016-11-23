package app

import (
	"image"
	"image/color"
	"math"

	"github.com/disintegration/gift"
	"github.com/sugyan/idol-face-linebot/recognizer"
)

func extractFaceImage(src image.Image, face recognizer.RecognizedFace) image.Image {
	xMin := math.MaxInt32
	xMax := math.MinInt32
	yMin := math.MaxInt32
	yMax := math.MinInt32
	for _, b := range face.Bounding {
		if b.X < xMin {
			xMin = b.X
		}
		if b.X > xMax {
			xMax = b.X
		}
		if b.Y < yMin {
			yMin = b.Y
		}
		if b.Y > yMax {
			yMax = b.Y
		}
	}
	g := gift.New(gift.Rotate(float32(face.Angle.Roll), color.Black, gift.CubicInterpolation))
	dstBounds := g.Bounds(src.Bounds())
	center := translate(
		image.Pt(
			(xMin+xMax+dstBounds.Dx()-src.Bounds().Dx()+1)/2,
			(yMin+yMax+dstBounds.Dy()-src.Bounds().Dy()+1)/2,
		), // target center point + offsets
		image.Pt(
			(dstBounds.Dx()+1)/2,
			(dstBounds.Dy()+1)/2,
		), // center of rotated bounds
		-face.Angle.Roll,
	)
	crop := image.Rect(
		center.X-(xMax-xMin+1)/2,
		center.Y-(yMax-yMin+1)/2,
		center.X+(xMax-xMin+1)/2,
		center.Y+(yMax-yMin+1)/2,
	)
	g.Add(
		gift.Crop(crop),
		gift.ResizeToFit(302, 200, gift.CubicResampling),
	)
	dst := image.NewRGBA(g.Bounds(src.Bounds()))
	g.Draw(dst, src)
	// TODO: fill to 1.51:1
	return dst
}

func translate(p, c image.Point, deg float64) image.Point {
	rad := math.Pi / 180 * deg
	x := math.Cos(rad)*float64(p.X-c.X) - math.Sin(rad)*float64(p.Y-c.Y) + float64(c.X)
	y := math.Sin(rad)*float64(p.X-c.X) + math.Cos(rad)*float64(p.Y-c.Y) + float64(c.Y)
	return image.Pt(int(x+.5), int(y+.5))
}
