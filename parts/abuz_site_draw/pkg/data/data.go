package data

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Controllers struct {
	User          UserController
	Price         PriceController
	TelegramToken string
}

func MakeControllers(db *gorm.DB, baseLog zerolog.Logger, telegramToken string) Controllers {
	return Controllers{
		User:          NewUserController(db, baseLog),
		Price:         NewPriceController(db, baseLog),
		TelegramToken: telegramToken,
	}
}
