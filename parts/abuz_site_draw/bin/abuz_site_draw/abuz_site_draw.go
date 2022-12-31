package main

import (
	"abuz_site_draw/parts/abuz_site_draw/pkg/bot"
	"abuz_site_draw/parts/abuz_site_draw/pkg/client"
	"abuz_site_draw/parts/abuz_site_draw/pkg/data"
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
	port          int
	telegramToken string
	ws            string
)

// TODO 8000
func init() {
	flag.IntVar(&port, "port", 8000, "set port")
	flag.StringVar(&telegramToken, "token", "5846375311:AAGf_hr2KCsPhY81NnTp3Z1iBMckR2CQAwk", "set telegram token")
	flag.StringVar(&ws, "ws", "https://zarabotay.info", "set website domain")
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

	c := data.MakeControllers(db, httpLogger, telegramToken)

	bot.StartTelegramBot(telegramToken, ws, &c)

	err = client.NewController(db, r, &c)
	if err != nil {
		log.Fatal().Err(err).Msg("fail create web")
	}

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		log.Fatal().Err(err).Msg("fail start server")
	}
}
