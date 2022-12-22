package aximgcolor

import (
	"github.com/rs/zerolog/log"
	"image"
	"image/png"
	"os"
	"testing"
)

func TestLightDarkPixelCount(t *testing.T) {

	darkImage, err := getImageFromFilePath("./dark.png")
	if err != nil {
		t.Fatal(err)
	}
	l, d := LightDarkPixelCount(darkImage)
	log.Info().Int("l", l).Int("d", d).Bool("is Dart", l < d).Msg("result dark")

	lightImage, err := getImageFromFilePath("./light.png")
	if err != nil {
		t.Fatal(err)
	}
	l, d = LightDarkPixelCount(lightImage)
	log.Info().Int("l", l).Int("d", d).Bool("is Dart", l < d).Msg("result light")
}

func TestBrightness(t *testing.T) {
	darkImage, err := getImageFromFilePath("./black.png")
	if err != nil {
		t.Fatal(err)
	}
	l := Brightness(darkImage)
	log.Info().Int("b", l).Msg("result black")

	darkImage, err = getImageFromFilePath("./dark.png")
	if err != nil {
		t.Fatal(err)
	}
	l = Brightness(darkImage)
	log.Info().Int("b", l).Msg("result dark")

	lightImage, err := getImageFromFilePath("./light.png")
	if err != nil {
		t.Fatal(err)
	}
	l = Brightness(lightImage)
	log.Info().Int("l", l).Msg("result light")

	lightImage, err = getImageFromFilePath("./white.png")
	if err != nil {
		t.Fatal(err)
	}
	l = Brightness(lightImage)
	log.Info().Int("l", l).Msg("result white")
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, err := png.Decode(f)
	return image, err
}
