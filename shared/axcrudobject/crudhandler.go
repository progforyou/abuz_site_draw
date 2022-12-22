package axcrudobject

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Pagination struct {
	Page    int                    `json:"page"`
	PerPage int                    `json:"perPage"`
	Sort    *PaginationSort        `json:"sort,omitempty"`
	Search  *string                `json:"search,omitempty"`
	Filter  map[string]interface{} `json:"filter"`
}

type PaginationOrderDirection string

const (
	DESC PaginationOrderDirection = "DESC"
	ASC                           = "ASC"
)

type PaginationSort struct {
	Field string                   `json:"field"`
	Order PaginationOrderDirection `json:"order"`
}

var ZeroPagination = Pagination{Page: 0, PerPage: 50}

func NewPagination(v url.Values) (res Pagination) {
	p := v.Get("pagination")
	if p == "" {
		return ZeroPagination
	}
	err := json.Unmarshal([]byte(p), &res)
	if err != nil {
		log.Error().Err(err).Str("p", p).Msg("unmarshal pagination")
		return ZeroPagination
	}
	res.Page -= 1
	p = v.Get("sort")
	if p != "" {
		var sort PaginationSort
		err = json.Unmarshal([]byte(p), &sort)
		if err != nil {
			log.Error().Err(err).Str("p", p).Msg("unmarshal sort")
			return ZeroPagination
		}
		res.Sort = &sort
	}
	p = v.Get("filter")
	if p != "" {
		var filter map[string]interface{}
		err = json.Unmarshal([]byte(p), &filter)
		if err != nil {
			log.Error().Err(err).Str("p", p).Msg("unmarshal filter")
			return ZeroPagination
		}
		if q, ok := filter["__q"]; ok {
			qq := q.(string)
			res.Search = &qq
			delete(filter, "__q")
		}
		res.Filter = filter
	}
	log.Info().Interface("pagination", res).Msg("p")
	return res
}

func NewPaginationOld(v url.Values) Pagination {
	p, pp := 0, 50
	if val, err := strconv.Atoi(v.Get("p")); err == nil {
		p = val
	}
	if val, err := strconv.Atoi(v.Get("pp")); err == nil {
		pp = val
	}
	return Pagination{Page: p, PerPage: pp}
}

type CrudRouterController struct {
	Name       string
	GetAll     func(p Pagination) interface{}
	GetOne     func(id uint64) interface{}
	GetMany    func(ids []uint64, p Pagination) interface{}
	Create     func(r []byte) (interface{}, error)
	Update     func(id uint64, r []byte) (interface{}, error)
	Delete     func(id uint64) error
	DeleteMany func(ids []uint64) error
	DeleteAll  func()
	Size       func(p Pagination) int64
	Log        zerolog.Logger
	Opts       CrudOptions
}

type CrudOptions struct {
	MediaPath     string
	Media         string
	ThumbnailPath string
	Static        string
	StaticPath    string
	PageTitle     string
	PrefixPath    string
	WebUrl        string
	Opts          map[string]interface{}
}

var Opts = CrudOptions{
	MediaPath:  "./media/",
	Media:      "/media/",
	PrefixPath: "/",
}

type Model struct {
	ID        uint64         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type UModel struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func NewCrudRouter(controller CrudRouterController) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if controller.GetAll == nil {
				w.WriteHeader(501)
				return
			}
			p := NewPagination(r.URL.Query())
			log.Info().Interface("p", p).Msg("Pagination")
			w.Header().Add("content-total", strconv.FormatInt(controller.Size(p), 10))
			data := controller.GetAll(p)
			render.JSON(w, r, data)
		})
		r.Get("/{id:\\d+}/", func(w http.ResponseWriter, r *http.Request) {
			if controller.GetOne == nil {
				w.WriteHeader(501)
				return
			}
			id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
			if err != nil {
				w.WriteHeader(403)
				return
			}
			obj := controller.GetOne(id)
			if obj == nil {
				w.WriteHeader(404)
				return
			}
			render.JSON(w, r, obj)
		})
		r.Get("/many/", func(w http.ResponseWriter, r *http.Request) {
			if controller.GetMany == nil {
				w.WriteHeader(501)
				return
			}
			ids, err := getIdsFromString(r.URL.Query().Get("ids"))
			if err != nil {
				w.WriteHeader(403)
				return
			}
			p := NewPagination(r.URL.Query())
			w.Header().Add("content-total", strconv.FormatInt(controller.Size(p), 10))
			obj := controller.GetMany(ids, p)
			render.JSON(w, r, obj)
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			if controller.Update == nil {
				w.WriteHeader(501)
				controller.Log.Error().Msg("not implemented")
				return
			}
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				controller.Log.Error().Err(err).Msg("read body")
				w.WriteHeader(403)
				return
			}
			obj, err := controller.Create(body)
			if err != nil {
				controller.Log.Error().Err(err).Msg("create")
				w.WriteHeader(403)
				return
			}
			render.JSON(w, r, obj)
		})
		r.Put("/{id:\\d+}/", func(w http.ResponseWriter, r *http.Request) {
			if controller.Update == nil {
				w.WriteHeader(501)
				controller.Log.Error().Msg("not implemented")
				return
			}
			id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
			if err != nil {
				w.WriteHeader(403)
				controller.Log.Error().Err(err).Msg("parse id")
				return
			}
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(403)
				controller.Log.Error().Err(err).Msg("read body")
				return
			}
			obj, err := controller.Update(id, body)
			if err != nil {
				w.WriteHeader(500)
				controller.Log.Error().Err(err).Msg("update")
				return
			}
			render.JSON(w, r, obj)
		})
		r.Delete("/{id:\\d+}/", func(w http.ResponseWriter, r *http.Request) {
			if controller.Delete == nil {
				w.WriteHeader(501)
				controller.Log.Error().Msg("not implemented")
				return
			}
			id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
			if err != nil {
				w.WriteHeader(403)
				controller.Log.Error().Err(err).Msg("parse id")
				return
			}
			err = controller.Delete(id)
			if err != nil {
				w.WriteHeader(500)
				controller.Log.Error().Err(err).Msg("delete")
				return
			}
			w.WriteHeader(201)
		})
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			if controller.DeleteMany == nil {
				w.WriteHeader(501)
				controller.Log.Error().Msg("not implemented")
				return
			}
			ids, err := getIdsFromString(r.URL.Query().Get("ids"))
			if err != nil {
				w.WriteHeader(403)
				controller.Log.Error().Err(err).Msg("parse id")
				return
			}
			err = controller.DeleteMany(ids)
			if err != nil {
				w.WriteHeader(500)
				controller.Log.Error().Err(err).Msg("delete many")
				return
			}
			w.WriteHeader(201)
		})
	}
}

func getIdsFromString(ids string) ([]uint64, error) {
	if len(ids) == 0 {
		return []uint64{}, nil
	}
	var uintIds []uint64
	for _, id := range strings.Split(ids, ",") {
		uintid, err := strconv.ParseUint(strings.TrimSpace(id), 10, 64)
		if err != nil {
			log.Error().Msgf("error parse argument: ids:%s", ids)
			return nil, err
		}
		uintIds = append(uintIds, uintid)
	}
	return uintIds, nil
}
