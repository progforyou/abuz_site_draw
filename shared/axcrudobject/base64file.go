package axcrudobject

import (
	"abuz_site_draw/shared/aximgcolor"
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Base64Image struct {
	Base64File
	Brightness int
}

type Base64File struct {
	RawFile *Base64String `json:"rawFile" gorm:"-"`
	Title   string        `json:"title" gorm:"column:img_alt"`
	Src     string        `json:"src,omitempty" gorm:"column:img_src"`
	Mime    string        `json:"mime" gorm:"column:img_mime"`
}

type Base64FileOpts struct {
	Prefix            string
	SaveOriginal      bool
	AddDate           bool
	MediaAsFirstThumb bool
	Thumbnails        []ThumbnailBase64File
}

type ThumbnailBase64File struct {
	Width  int
	Height int
}

// data:image/png;base64,
type Base64String string

func (b Base64String) Bytes() ([]byte, error) {
	splitStrings := strings.SplitN(string(b), ",", 2)
	if uDec, err := b64.StdEncoding.DecodeString(splitStrings[1]); err != nil {
		return []byte{}, err
	} else {
		return uDec, nil
	}
}

var pattern = regexp.MustCompile("^data:([a-zA-Z0-9_/-]+)$")

func (b Base64String) Mime() string {
	splitStrings := strings.SplitN(string(b), ";", 2)
	if pattern.MatchString(splitStrings[0]) {
		return pattern.FindStringSubmatch(splitStrings[0])[1]
	}
	return "octet-stream"
}

func (bf Base64Image) IsDark() bool {
	return bf.Brightness < 128
}

func (bf *Base64Image) Write(mediaUrl string, mediaPath string, opts Base64FileOpts) error {

	if bf.RawFile == nil {
		return nil
	}
	fileName, ext := slug.Make(strings.TrimSuffix(bf.Title, path.Ext(bf.Title))), path.Ext(bf.Title)
	var err error
	imgBytes, err := bf.RawFile.Bytes()
	if err != nil {
		return err
	}
	bf.Mime = bf.RawFile.Mime()
	mpath, fName := getMediaPath(mediaPath, fileName, ext, opts.Prefix, opts.AddDate, 0)
	pt, err := filepath.Abs("./" + mpath)
	if err != nil {
		return err
	}
	err = createDir(filepath.Dir(pt))
	if err != nil {
		return err
	}
	bf.Brightness = aximgcolor.BrightnessFromBytes(imgBytes)
	if opts.SaveOriginal {
		err = os.WriteFile(pt, imgBytes, 0644)
		if err != nil {
			return err
		}
	}
	for _, tOpt := range opts.Thumbnails {
		err = thumb(pt, imgBytes, tOpt)
		if err != nil {
			return err
		}
	}
	src, err := getMediaUrl(mediaUrl, fName, opts.Prefix, opts.AddDate)
	if err != nil {
		return err
	}
	log.Debug().Str("path", pt).Str("media-path", mediaPath).Str("media-url", mediaUrl).Str("media", src).Msg("write base64 image")
	bf.Src = src
	return nil
}

func (bf *Base64File) Write(mediaUrl string, mediaPath string, opts Base64FileOpts) (string, error) {
	if bf.RawFile == nil {
		return "", nil
	}
	fileName, ext := slug.Make(strings.TrimSuffix(bf.Title, path.Ext(bf.Title))), path.Ext(bf.Title)
	var err error
	imgBytes, err := bf.RawFile.Bytes()
	if err != nil {
		return "", err
	}
	bf.Mime = bf.RawFile.Mime()
	mpath, fName := getMediaPath(mediaPath, fileName, ext, opts.Prefix, opts.AddDate, 0)
	pt, err := filepath.Abs("./" + mpath)
	if err != nil {
		return "", err
	}
	err = createDir(filepath.Dir(pt))
	if err != nil {
		return "", err
	}
	err = os.WriteFile(pt, imgBytes, 0644)
	if err != nil {
		return "", err
	}
	src, err := getMediaUrl(mediaUrl, fName, opts.Prefix, opts.AddDate)
	if err != nil {
		return "", err
	}
	bf.Src = src
	log.Debug().Str("path", pt).Str("media", src).Msg("write base64 file")
	return pt, nil
}

func (bf *Base64File) Read(mediaPath string, suffix string) ([]byte, error) {
	path := strings.Replace(bf.Src, "/media/", mediaPath, 1) + suffix
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

func getMediaUrl(mediaUrl string, fileName string, prefix string, addDate bool) (string, error) {
	baseUrl, err := url.Parse(mediaUrl)
	if err != nil {
		return "", err
	}
	if prefix != "" {
		baseUrl, err = baseUrl.Parse(prefix + "/")
		if err != nil {
			return "", err
		}
	}
	if addDate {
		tm := time.Now()
		baseUrl, err = baseUrl.Parse(tm.Format("2006/01/02/"))
		if err != nil {
			return "", err
		}
	}
	baseUrl, err = baseUrl.Parse(fileName)
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("get media url from url:%s, file:%s, prefix:%s, addDate:%v = %s", mediaUrl, fileName, prefix, addDate, baseUrl.String())
	return baseUrl.String(), nil
}

func getMediaPath(mediaPath string, fileName string, ext string, prefix string, addDate bool, suffix int) (string, string) {
	var res string
	fName := fmt.Sprintf("%s%s", fileName, ext)
	if suffix > 0 {
		fName = fmt.Sprintf("%s-(%d)%s", fileName, suffix, ext)
	}

	if addDate {
		tm := time.Now()
		res = path.Join(mediaPath, prefix, tm.Format("2006/01/02"), fName)
	} else {
		res = path.Join(mediaPath, prefix, fName)
	}
	absPath, _ := filepath.Abs("./" + res)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Info().Msgf("File %s not found!", res)
		return res, fName
	} else {
		if suffix > 1000 {
			panic("max os file suffix iterations!")
		}
		return getMediaPath(mediaPath, fileName, ext, prefix, addDate, suffix+1)
	}
}

func createDir(absDir string) error {
	log.Debug().Msgf("Create folder %s", absDir)
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func thumb(baseFilePath string, imageBytes []byte, opt ThumbnailBase64File) error {
	srcImage, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return err
	}
	dstImageFill := imaging.Fill(srcImage, opt.Width, opt.Height, imaging.Center, imaging.Lanczos)
	out, err := os.Create(fmt.Sprintf("%s.%d_%d.jpg.webp", baseFilePath, opt.Width, opt.Height))
	defer out.Close()
	if err != nil {
		return err
	}
	var jpegOpt jpeg.Options
	jpegOpt.Quality = 1000
	err = jpeg.Encode(out, dstImageFill, &jpegOpt)
	if err != nil {
		return err
	}
	//err = os.WriteFile(pt, bytes, 0644)
	return nil
}
