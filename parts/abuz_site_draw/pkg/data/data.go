package data

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Controllers struct {
	User   UserController
	Reward RewardController
}

func MakeControllers(db *gorm.DB, baseLog zerolog.Logger) Controllers {
	return Controllers{
		User:   NewUserController(db, baseLog),
		Reward: NewRewardController(db, baseLog),
	}
}
