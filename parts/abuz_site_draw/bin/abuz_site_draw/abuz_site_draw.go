package main

import (
	"bot_tasker/parts/abuz_site_draw/pkg/client"
	"bot_tasker/parts/abuz_site_draw/pkg/data"
	"flag"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
)

// -build-me-for: native
// -build-me-for: linux

var (
	port int
)

func init() {
	flag.IntVar(&port, "port", 8000, "set port")
	flag.Parse()
}

func main() {
	var err error
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05,000"}).Level(zerolog.DebugLevel)
	log.Debug().Msgf("Start Telegram Blog server on port %d", port)

	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Warn),
	})

	r := chi.NewRouter()
	httpLogger := log.With().Str("service", "http").Logger().Level(zerolog.InfoLevel)

	c := data.MakeControllers(db, httpLogger)

	err = client.NewController(db, r, &c)
	if err != nil {
		log.Fatal().Err(err).Msg("fail create web")
	}

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		log.Fatal().Err(err).Msg("fail start server")
	}
}
