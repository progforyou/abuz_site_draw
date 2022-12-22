package gogate

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

var timeout = time.Second * 5

func NewController(r *chi.Mux, host string) error {
	hostMatcher, err := regexp.Compile("(?P<service>\\w+)\\." + regexp.QuoteMeta(host))
	if err != nil {
		return err
	}
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		matches := hostMatcher.FindStringSubmatch(r.Host)
		if len(matches) == 0 {
			root(w, r)
		} else {
			err := service(matches[1], w, r)
			if err != nil {
				w.WriteHeader(500)
				w.Write(([]byte)("500 internal server error: " + err.Error()))
			}
		}
	})
	return nil
}

func root(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("ROOT!")
	w.Header().Set("x-gate", "root")
	w.Write(([]byte)("ROOT"))
}

func service(name string, w http.ResponseWriter, r *http.Request) (err error) {
	srv, ok := serviceChannel[name]
	if !ok {
		w.Write(([]byte)("404 service " + name + " not found"))
		w.WriteHeader(404)
		return
	}
	req, err := ToGateRequest(r)
	if err != nil {
		return
	}
	h := newRequestHolder(req)
	srv.reqChan <- h
	select {
	case res := <-h.resp:
		for _, hd := range res.Header {
			w.Header().Add(hd.Key, strings.Join(hd.Values, ","))
		}
		w.Header().Add("x-gate-ref", name)
		w.Write(res.Body)
		w.WriteHeader((int)(res.StatusCode))
		break
	case <-time.After(timeoutTTL):
		close(h.resp)
		w.Write(([]byte)("401 timeout"))
		w.WriteHeader(401)
		break
	}
	return
}

type requestHolder struct {
	req  *GateRequest
	resp chan *GateResponse
}

func newRequestHolder(r *GateRequest) *requestHolder {
	r.Id = atomic.AddUint64(&requestId, 1)
	res := &requestHolder{
		req:  r,
		resp: make(chan *GateResponse, 1),
	}

	go func() {
		time.Sleep(time.Second * 65)
		holderLock.Lock()
		defer holderLock.Unlock()
		delete(holder, res.req.Id)
	}()

	return res
}
