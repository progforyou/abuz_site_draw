package data

import (
	"bot_tasker/shared/axcrudobject"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type User struct {
	axcrudobject.Model
	Ip       string  `json:"ip"`
	Telegram string  `gorm:"unique" json:"telegram"`
	Prices   []Price `gorm:"foreignKey:UserRefer" json:"prices"`
	Hash     string  `json:"hash"`
	Reward   Reward  `gorm:"foreignKey:UserRefer" json:"reward"`
	Admin    bool    `json:"admin"`
}

var Admins = []string{"nikolay35977"}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.Admin = hasIsArrayStr(Admins, u.Telegram)
	return
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

func hasIsArrayStr(data []string, v string) bool {
	for _, datum := range data {
		if datum == v {
			return true
		}
	}
	return false
}
