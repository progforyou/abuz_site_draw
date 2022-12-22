package axcrudobject

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

type hijack404 struct {
	http.ResponseWriter
	R         *http.Request
	notFound  bool
	Handle404 func(w http.ResponseWriter, r *http.Request) bool
}

func (h *hijack404) WriteHeader(code int) {
	if 404 == code {
		h.notFound = true
	}
	if 404 == code && h.Handle404(h.ResponseWriter, h.R) {
		panic(h)
	}
	h.ResponseWriter.WriteHeader(code)
}

func CreateThumbnail(prefix string, h http.Handler) http.Handler {
	log.Info().Msgf("Create tbnl")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nw := w
		log.Info().Msgf("FILE REQUEST: %s", r.URL.Path)
		hijack := &hijack404{ResponseWriter: w, R: r, Handle404: handle404}
		h.ServeHTTP(hijack, r)
		if hijack.notFound {
			log.Warn().Msg("Not found after server")
			r.URL.Path = "/media/1.jpg"
			h.ServeHTTP(nw, r)
		}
	})
}

func handle404(w http.ResponseWriter, r *http.Request) bool {
	log.Info().Msgf("File %s not found!", r.URL.Path)

	return false
}
