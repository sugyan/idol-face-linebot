package app

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/disintegration/gift"
)

func rotateAndCropImage(src image.Image, rect image.Rectangle, angle float64) image.Image {
	g := gift.New(gift.Rotate(float32(angle), color.Black, gift.CubicInterpolation))
	dstBounds := g.Bounds(src.Bounds())
	center := rotatePoint(
		image.Pt(
			(rect.Min.X+rect.Max.X+dstBounds.Dx()-src.Bounds().Dx()+1)/2,
			(rect.Min.Y+rect.Max.Y+dstBounds.Dy()-src.Bounds().Dy()+1)/2,
		), // target center point + offsets
		image.Pt(
			(dstBounds.Dx()+1)/2,
			(dstBounds.Dy()+1)/2,
		), // center of rotated bounds
		-angle,
	)
	// crop x1.2 size rectangle
	crop := image.Rect(
		center.X-int(float32(rect.Dx())*0.6+.5),
		center.Y-int(float32(rect.Dy())*0.6+.5),
		center.X+int(float32(rect.Dx())*0.6+.5),
		center.Y+int(float32(rect.Dy())*0.6+.5),
	)
	g.Add(
		gift.Crop(crop),
		gift.ResizeToFit(302, 200, gift.CubicResampling),
	)
	dst := image.NewRGBA(g.Bounds(src.Bounds()))
	g.Draw(dst, src)
	return dst
}

func padForThumbnailImage(src image.Image) image.Image {
	w := int(float32(src.Bounds().Dy())*1.51 + .5)
	dst := image.NewRGBA(image.Rect(0, 0, w, src.Bounds().Dy()))
	draw.Draw(dst, src.Bounds().Add(image.Pt((w-src.Bounds().Dx()+1)/2, 0)), src, image.ZP, draw.Src)
	return dst
}

func rotatePoint(p, c image.Point, deg float64) image.Point {
	rad := math.Pi / 180 * deg
	x := math.Cos(rad)*float64(p.X-c.X) - math.Sin(rad)*float64(p.Y-c.Y) + float64(c.X)
	y := math.Sin(rad)*float64(p.X-c.X) + math.Cos(rad)*float64(p.Y-c.Y) + float64(c.Y)
	return image.Pt(int(x+.5), int(y+.5))
}
