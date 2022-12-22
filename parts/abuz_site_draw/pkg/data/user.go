package data

import (
	"bot_tasker/shared/axcrudobject"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type User struct {
	axcrudobject.Model
	Ip string
}

type UserController struct {
	Set func(string) error
}

func NewUserController(db *gorm.DB, baseLog zerolog.Logger) UserController {
	log := baseLog.With().Str("model", "user").Logger()
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	return UserController{
		Set: func(ip string) error {

			return nil
		},
	}
}
