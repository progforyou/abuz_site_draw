package client

import (
	"bot_tasker/parts/abuz_site_draw/pkg/data"
	"bytes"
	"embed"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"net/http"
	"text/template"
	"time"
)

var (
	//go:embed "static/*"
	htmlStatic embed.FS

	//go:embed "template/index.html"
	indexTemplate []byte
)

var xAuthSessionName = "x-auth-session"

type DataIndexPage struct {
	Data data.Reward
}

type DataIndexPost struct {
	Timer time.Time  `json:"timer"`
	Price data.Price `json:"price"`
}

const IP = "127.0.0.1"

func NewController(db *gorm.DB, r *chi.Mux, c *data.Controllers) error {
	log.Info().Msg("create page controller")
	wrap := func(f func(db *gorm.DB, w http.ResponseWriter, r *http.Request, c *data.Controllers)) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			f(db, w, r, c)
		}
	}
	r.Get("/", wrap(index))
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(xAuthSessionName)
		var session string
		if err != nil {
			if err != http.ErrNoCookie {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
		} else {
			session = cookie.Value
		}
		ip := r.Header.Get("X-Real-IP")
		ip = IP
		has := c.Reward.Check(ip, session)
		if has {
			err = c.Reward.Set(ip, session)
		} else {
			err = c.Reward.Create(ip, session)
		}
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set")
			w.Write(([]byte)(err.Error()))
			return
		}
		obj, err := c.Reward.Get(ip, session)
		jsonBytes, err := json.Marshal(DataIndexPost{Timer: obj.Timer, Price: data.TestGeneratePrice()})
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to marshal json")
			w.Write(([]byte)(err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write(jsonBytes)
	})
	r.Handle("/static/*", http.FileServer(http.FS(htmlStatic)))
	return nil
}

func index(db *gorm.DB, w http.ResponseWriter, r *http.Request, c *data.Controllers) {
	cookie, err := r.Cookie(xAuthSessionName)
	ip := r.Header.Get("X-Real-IP")
	ip = IP
	var session string
	if err != nil {
		if err != http.ErrNoCookie {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	} else {
		session = cookie.Value
	}
	if session == "" {
		sessionUuid, err := uuid.NewUUID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session = sessionUuid.String()
		sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
		http.SetCookie(w, &sessionCookie)
	}
	dataR, err := c.Reward.Get(ip, session)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to get from db")
		w.Write(([]byte)(err.Error()))
		return
	}
	dataIndex := DataIndexPage{Data: dataR}
	// Generate template
	result, err := Render(indexTemplate, dataIndex)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to render")
		w.Write(([]byte)(err.Error()))
		return
	}
	w.Write(result)
}

func mkSlice(args ...interface{}) []interface{} {
	return args
}

func Render(templateByte []byte, data interface{}) ([]byte, error) {
	funcMap := map[string]interface{}{"mkSlice": mkSlice}
	t, err := template.New("").Funcs(funcMap).Parse(string(templateByte))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create template")
		return nil, err
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to render template")
		return nil, err
	}
	return tpl.Bytes(), nil
}
