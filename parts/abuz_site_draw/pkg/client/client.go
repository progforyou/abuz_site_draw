package client

import (
	"abuz_site_draw/parts/abuz_site_draw/pkg/data"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

var (
	//go:embed "static/*"
	htmlStatic embed.FS

	//go:embed "template/index.html"
	indexTemplate []byte

	//go:embed "template/lk.html"
	lkTemplate []byte

	//go:embed "template/winning.html"
	winningTemplate []byte

	//go:embed "template/admin_winning.html"
	adminWinningTemplate []byte
)

var xAuthSessionName = "x-auth-session"

type DataIndexPage struct {
	Data data.User
}

type LoginRequest struct {
	Telegram string `json:"telegram"`
	Hash     string `json:"hash"`
	HashData string `json:"hash_data"`
}

type DataIndexPost struct {
	Timer time.Time  `json:"timer"`
	Price data.Price `json:"price"`
}

type DataAdminPage struct {
	Data []data.User `json:"data"`
}

type DataAdminGet struct {
	Data []data.Ip `json:"data"`
}

//const IP = "127.0.0.1"

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
		if session == "" {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("you hacker")
			w.Write(([]byte)(err.Error()))
			return
		}
		ip := r.Header.Get("X-Real-IP")
		dataU, err := c.User.Get(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to get")
			w.Write(([]byte)(err.Error()))
			return
		}
		if dataU.Ban {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if dataU.ID == 0 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		err = c.User.StartGame(ip, session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set")
			w.Write(([]byte)(err.Error()))
			return
		}
		obj, err := c.User.Get(session)
		jsonBytes, err := json.Marshal(DataIndexPost{Timer: obj.Timer, Price: obj.Prices[len(obj.Prices)-1]})
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to marshal json")
			w.Write(([]byte)(err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write(jsonBytes)
	})
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		var login LoginRequest
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
		if session == "" {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("you hacker")
			w.Write(([]byte)("Hacker"))
			return
		}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&login); err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("error decode")
			w.Write(([]byte)(err.Error()))
			return
		}
		if compareHash(login, c) {
			err = c.User.Login(session, login.Telegram)
			if err != nil {
				w.WriteHeader(500)
				w.Write(([]byte)(err.Error()))
				return
			}
			w.Write(([]byte)("OK"))
		} else {
			w.WriteHeader(500)
			w.Write(([]byte)("Hacker"))
			return
		}
	})
	r.Get("/lk", wrap(lk))
	r.Get("/reward", wrap(reward))
	r.Get("/promo/{hash}", wrap(rewardPromo))
	r.Get("/price/{hash}", wrap(rewardPrice))
	r.Get("/admin/reward", wrap(adminRewardPrice))
	r.Get("/admin/reward/{id}", wrap(adminRewardPrices))
	r.Get("/admin/ips/{id}", func(w http.ResponseWriter, r *http.Request) {
		ids := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(ids, 10, 64)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set user to db")
			w.Write(([]byte)(err.Error()))
			return
		}
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
		if session == "" {
			sessionUuid, err := uuid.NewUUID()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session = sessionUuid.String()
			sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
			err = c.User.CreateSession(session)
			if err != nil {
				w.WriteHeader(500)
				log.Error().Err(err).Msg("fail to set user to db")
				w.Write(([]byte)(err.Error()))
				return
			}
			http.SetCookie(w, &sessionCookie)
		}
		dataU, err := c.User.Get(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to get")
			w.Write(([]byte)(err.Error()))
			return
		}
		if dataU.ID == 0 {
			w.WriteHeader(403)
			w.Write(([]byte)(err.Error()))
			return
		}
		if !dataU.Admin {
			w.WriteHeader(403)
			w.Write(([]byte)(err.Error()))
			return
		}
		dataR, err := c.User.GetAllIps(id)

		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to get from db")
			w.Write(([]byte)(err.Error()))
			return
		}
		jsonBytes, err := json.Marshal(DataAdminGet{Data: dataR})
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to marshal json")
			w.Write(([]byte)(err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write(jsonBytes)
	})
	r.Get("/admin/ban/{tg}", func(w http.ResponseWriter, r *http.Request) {
		tg := chi.URLParam(r, "tg")
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
		if session == "" {
			sessionUuid, err := uuid.NewUUID()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session = sessionUuid.String()
			sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
			err = c.User.CreateSession(session)
			if err != nil {
				w.WriteHeader(500)
				log.Error().Err(err).Msg("fail to set user to db")
				w.Write(([]byte)(err.Error()))
				return
			}
			http.SetCookie(w, &sessionCookie)
		}
		dataU, err := c.User.Get(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to get")
			w.Write(([]byte)(err.Error()))
			return
		}
		if dataU.ID == 0 {
			w.WriteHeader(403)
			w.Write(([]byte)(err.Error()))
			return
		}
		if !dataU.Admin {
			w.WriteHeader(403)
			w.Write(([]byte)(err.Error()))
			return
		}
		err = c.User.Block(tg)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to get from db")
			w.Write(([]byte)(err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	r.Get("/admin/unban/{tg}", func(w http.ResponseWriter, r *http.Request) {
		tg := chi.URLParam(r, "tg")
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
		if session == "" {
			sessionUuid, err := uuid.NewUUID()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session = sessionUuid.String()
			sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
			err = c.User.CreateSession(session)
			if err != nil {
				w.WriteHeader(500)
				log.Error().Err(err).Msg("fail to set user to db")
				w.Write(([]byte)(err.Error()))
				return
			}
			http.SetCookie(w, &sessionCookie)
		}
		dataU, err := c.User.Get(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to get")
			w.Write(([]byte)(err.Error()))
			return
		}
		if dataU.ID == 0 {
			w.WriteHeader(403)
			w.Write(([]byte)(err.Error()))
			return
		}
		if !dataU.Admin {
			w.WriteHeader(403)
			w.Write(([]byte)(err.Error()))
			return
		}
		err = c.User.UnBlock(tg)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to get from db")
			w.Write(([]byte)(err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	r.Handle("/static/*", http.FileServer(http.FS(htmlStatic)))
	return nil
}

func index(db *gorm.DB, w http.ResponseWriter, r *http.Request, c *data.Controllers) {
	cookie, err := r.Cookie(xAuthSessionName)
	var session string
	if err != nil {
		if err != http.ErrNoCookie {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	} else {
		session = cookie.Value
		err = c.User.CreateSession(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set user to db")
			w.Write(([]byte)(err.Error()))
			return
		}
	}
	if session == "" {
		sessionUuid, err := uuid.NewUUID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session = sessionUuid.String()
		sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
		err = c.User.CreateSession(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set user to db")
			w.Write(([]byte)(err.Error()))
			return
		}
		http.SetCookie(w, &sessionCookie)
	}

	dataU, err := c.User.Get(session)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to get from db")
		w.Write(([]byte)(err.Error()))
		return
	}
	if dataU.Ban {
		dataU = data.User{}
	}
	dataIndex := DataIndexPage{Data: dataU}
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

func lk(db *gorm.DB, w http.ResponseWriter, r *http.Request, c *data.Controllers) {
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
	dataR, err := c.User.Get(session)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to get from db")
		w.Write(([]byte)(err.Error()))
		return
	}
	if dataR.Ban {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if dataR.ID == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	dataLk := DataIndexPage{Data: dataR}
	// Generate template
	result, err := Render(lkTemplate, dataLk)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to render")
		w.Write(([]byte)(err.Error()))
		return
	}
	w.Write(result)
}

func reward(db *gorm.DB, w http.ResponseWriter, r *http.Request, c *data.Controllers) {
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
	if session == "" {
		sessionUuid, err := uuid.NewUUID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session = sessionUuid.String()
		sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
		err = c.User.CreateSession(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set user to db")
			w.Write(([]byte)(err.Error()))
			return
		}
		http.SetCookie(w, &sessionCookie)
	}
	dataR, err := c.User.Get(session)

	if dataR.Ban {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if dataR.ID == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to get from db")
		w.Write(([]byte)(err.Error()))
		return
	}
	dataLk := DataIndexPage{Data: dataR}
	// Generate template
	result, err := Render(winningTemplate, dataLk)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to render")
		w.Write(([]byte)(err.Error()))
		return
	}
	w.Write(result)
}

func rewardPromo(db *gorm.DB, w http.ResponseWriter, r *http.Request, c *data.Controllers) {
	hash := chi.URLParam(r, "hash")
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
	if session == "" {
		sessionUuid, err := uuid.NewUUID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session = sessionUuid.String()
		sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
		err = c.User.CreateSession(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set user to db")
			w.Write(([]byte)(err.Error()))
			return
		}
		http.SetCookie(w, &sessionCookie)
	}
	dataR, err := c.User.GetRewardPrice(session, hash)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to get from db")
		w.Write(([]byte)(err.Error()))
		return
	}

	dataU, err := c.User.Get(session)
	if dataU.Ban {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if dataU.ID == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to render")
		w.Write(([]byte)(err.Error()))
		return
	}
	dataP, err := c.Price.GetPromo(dataR.Data)
	w.Write([]byte(dataP.Data))
}

func rewardPrice(db *gorm.DB, w http.ResponseWriter, r *http.Request, c *data.Controllers) {
	hash := chi.URLParam(r, "hash")
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
	if session == "" {
		sessionUuid, err := uuid.NewUUID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session = sessionUuid.String()
		sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
		err = c.User.CreateSession(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set user to db")
			w.Write(([]byte)(err.Error()))
			return
		}
		http.SetCookie(w, &sessionCookie)
	}
	dataR, err := c.User.GetRewardPrice(session, hash)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to get from db")
		w.Write(([]byte)(err.Error()))
		return
	}

	dataU, err := c.User.Get(session)
	if dataU.Ban {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if dataU.ID == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to render")
		w.Write(([]byte)(err.Error()))
		return
	}
	dataP, err := c.Price.GetPrice(dataR.Data)
	w.Write([]byte(dataP.Data))
}

func adminRewardPrice(db *gorm.DB, w http.ResponseWriter, r *http.Request, c *data.Controllers) {
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
	if session == "" {
		sessionUuid, err := uuid.NewUUID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session = sessionUuid.String()
		sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
		err = c.User.CreateSession(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set user to db")
			w.Write(([]byte)(err.Error()))
			return
		}
		http.SetCookie(w, &sessionCookie)
	}
	dataU, err := c.User.Get(session)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to get from db")
		w.Write(([]byte)(err.Error()))
		return
	}

	if dataU.Ban {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if dataU.ID == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !dataU.Admin {
		w.WriteHeader(403)
		log.Error().Err(err).Msg("you not admin")
		w.Write(([]byte)(err.Error()))
		return
	}

	dataR := c.User.GetAll()
	dataLk := DataAdminPage{Data: dataR}
	// Generate template
	result, err := Render(adminWinningTemplate, dataLk)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to render")
		w.Write(([]byte)(err.Error()))
		return
	}
	w.Write(result)
}

func adminRewardPrices(db *gorm.DB, w http.ResponseWriter, r *http.Request, c *data.Controllers) {
	ids := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(ids, 10, 64)
	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to set user to db")
		w.Write(([]byte)(err.Error()))
		return
	}
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
	if session == "" {
		sessionUuid, err := uuid.NewUUID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session = sessionUuid.String()
		sessionCookie := http.Cookie{Name: xAuthSessionName, Value: session, Expires: time.Now().Add(365 * 24 * time.Hour)}
		err = c.User.CreateSession(session)
		if err != nil {
			w.WriteHeader(500)
			log.Error().Err(err).Msg("fail to set user to db")
			w.Write(([]byte)(err.Error()))
			return
		}
		http.SetCookie(w, &sessionCookie)
	}
	dataR, err := c.User.GetById(id)

	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("fail to get from db")
		w.Write(([]byte)(err.Error()))
		return
	}
	dataU, err := c.User.Get(session)
	if dataU.Ban {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if dataU.ID == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !dataU.Admin {
		w.WriteHeader(403)
		log.Error().Err(err).Msg("you not admin")
		w.Write(([]byte)(err.Error()))
		return
	}
	dataLk := DataIndexPage{Data: dataR}
	// Generate template
	result, err := Render(winningTemplate, dataLk)

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

func dateFormat(date time.Time) string {
	return fmt.Sprintf("%02d.%02d.%d",
		date.Day(), date.Month(), date.Year())
}

func getPrices(dataU data.User) bool {
	for _, price := range dataU.Prices {
		if price.Win {
			return true
		}
	}
	return false
}

func Render(templateByte []byte, data interface{}) ([]byte, error) {
	funcMap := map[string]interface{}{"mkSlice": mkSlice, "dateFormat": dateFormat, "getPrices": getPrices}
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

func compareHash(login LoginRequest, c *data.Controllers) bool {
	h := sha256.New()
	h.Write([]byte(c.TelegramToken))
	secret_key := h.Sum(nil)
	mac := hmac.New(sha256.New, secret_key)
	mac.Write([]byte(login.HashData))
	hm := mac.Sum(nil)
	return fmt.Sprintf("%x", hm) == login.Hash
}
