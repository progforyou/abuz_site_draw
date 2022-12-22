package axhttp

import (
	"bot_tasker/shared/axtools"
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/rs/zerolog/log"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var thumbRe = regexp.MustCompile(`(.*)\.(\d+)_(\d+).jpg$`)

//AutoThumb
// path - путь к медиа
// tmp - путь к temp.jpg файлам
func AutoThumb(path string, tmp string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("SMP", r.URL.Path).Bool("match", thumbRe.MatchString(r.URL.Path)).Msg("THUMB")
		if axtools.FileExist(filepath.Join(path, r.URL.Path)) {
			h.ServeHTTP(w, r)
			return
		}
		if thumbRe.MatchString(r.URL.Path) {
			sm := thumbRe.FindStringSubmatch(r.URL.Path)
			width, err := strconv.Atoi(sm[2])
			if err != nil {
				w.WriteHeader(500)
				return
			}
			height, err := strconv.Atoi(sm[3])
			if err != nil {
				w.WriteHeader(500)
				return
			}
			imagePath := filepath.Join(path, sm[1])
			if !axtools.FileExist(imagePath) {
				w.WriteHeader(404)
				w.Write([]byte("404 Image for thumbnail not found"))
				return
			}
			log.Info().Str("path", imagePath).Int("w", width).Int("h", height).Msg(sm[0])
			data, err := processImage(tmp, imagePath, width, height)
			if err != nil {
				log.Error().Err(err).Msg("create-thumb")
				w.Write([]byte(err.Error()))
				w.WriteHeader(500)
			} else {
				w.Header().Set("Content-Type", "image/jpeg")
				w.Write(data)
			}
			return
		}
		if h != nil {
			h.ServeHTTP(w, r)
		} else {
			w.WriteHeader(404)
			w.Write([]byte("404 not found"))
		}
	})
}

func processImage(tmp string, path string, w int, h int) ([]byte, error) {
	//Вначале поищем картинку
	hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s//%d//%d", path, w, h))))
	filePath := filepath.Join(tmp, hash+".jpg")
	if axtools.FileExist(filePath) { // Фаил уже есть просто вернем его
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	// Создадим фаил
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	srcImage, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	dstImageFill := imaging.Fill(srcImage, w, h, imaging.Center, imaging.Lanczos)
	out, err := os.Create(filePath)
	defer out.Close()
	if err != nil {
		return nil, err
	}
	var jpegOpt jpeg.Options
	jpegOpt.Quality = 1000
	err = jpeg.Encode(out, dstImageFill, &jpegOpt)
	log.Debug().Int("w", w).Int("h", h).Str("src", path).Str("dst", filePath).Msg("create-thumb")
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, dstImageFill, &jpegOpt)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
