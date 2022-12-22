package aximgcolor

import (
	"bytes"
	"image"
	"image/color"
	"math"
)

func LightDarkPixelCount(img image.Image) (int, int) {
	b := img.Bounds()

	var lightPixels, darkPixels int
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if luminancePercent(img.At(x, y)) > 50 {
				lightPixels++
			} else {
				darkPixels++
			}
		}
	}
	return lightPixels, darkPixels
}
func BrightnessFromBytes(b []byte) int {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return 128
	}
	return Brightness(img)
}

func Brightness(img image.Image) int {
	bn := img.Bounds()
	var avg float64
	var count uint64
	for y := bn.Min.Y; y < bn.Max.Y; y++ {
		for x := bn.Min.X; x < bn.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r = uint32((float64(r) / 0xffff) * 0xff)
			g = uint32((float64(g) / 0xffff) * 0xff)
			b = uint32((float64(b) / 0xffff) * 0xff)

			avg += math.Floor(float64(r+g+b) / 3.0)
			count++
		}
	}
	return int(avg / float64(count))
}

func luminancePercent(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	redPercent := float64(r) / 65535 * 100
	greenPercent := float64(g) / 65535 * 100
	bluePercent := float64(b) / 65535 * 100

	return redPercent*0.2126 + greenPercent*0.7152 + bluePercent*0.0722
}
