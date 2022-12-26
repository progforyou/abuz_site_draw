package data

import (
	"abuz_site_draw/shared/axcrudobject"
	"errors"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"time"
)

type Ip struct {
	axcrudobject.Model
	Address   string `json:"address"`
	UserRefer uint   `gorm:"primaryKey" json:"-"`
}

type User struct {
	axcrudobject.Model
	Ip       []Ip      `gorm:"foreignKey:UserRefer" json:"ip"`
	Telegram string    `json:"telegram"`
	Prices   []Price   `gorm:"foreignKey:UserRefer" json:"prices"`
	Hash     string    `json:"hash"`
	Admin    bool      `json:"admin"`
	Logined  bool      `gorm:"-"`
	Timer    time.Time `json:"timer"`
	Can      bool      `gorm:"-"`
}

var Admins = []string{"nikolay35977"}

const TIME = 24

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.Admin = hasIsArrayStr(Admins, u.Telegram)
	return
}

func (u *User) AfterFind(tx *gorm.DB) (err error) {
	u.Logined = u.Telegram != ""
	u.Can = u.Timer.Before(time.Now())
	return
}

type UserController struct {
	CreateSession  func(string) error
	Get            func(string) (User, error)
	StartGame      func(string, string) error
	CheckReward    func(string) bool
	Login          func(string, string) error
	GetRewardPrice func(string, string) (Price, error)
	GetAll         func() []User
	GetAllIps      func(uint64) ([]Ip, error)
}

func NewUserController(db *gorm.DB, baseLog zerolog.Logger) UserController {
	log := baseLog.With().Str("model", "user").Logger()
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	if err := db.AutoMigrate(&Ip{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	return UserController{
		CreateSession: func(session string) error {
			var obj User
			if err := db.Preload("Ip").Where("hash = ?", session).Find(&obj).Error; err != nil {
				return err
			}
			if obj.ID == 0 {
				obj.Hash = session
				if err := db.Create(&obj).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Get: func(session string) (User, error) {
			var obj User
			if err := db.Preload("Prices").Where("hash = ?", session).Find(&obj).Error; err != nil {
				return User{}, err
			}
			return obj, nil
		},
		StartGame: func(ip, session string) error {
			var obj User
			if err := db.Preload("Ip").Where("hash = ?", session).Find(&obj).Error; err != nil {
				return err
			}
			if !obj.Logined {
				return errors.New("need to login")
			}
			if obj.Timer.Before(time.Now()) {
				obj.Timer = time.Now().Add(time.Hour * TIME)
				obj.Ip = append(obj.Ip, Ip{Address: ip})
				obj.Prices = append(obj.Prices, TestGeneratePrice())
				tx := db.Save(&obj)
				if tx.Error != nil {
					log.Error().Err(tx.Error).Msg("db update error")
					return tx.Error
				}
			} else {
				return errors.New("wait for time")
			}
			return nil
		},
		CheckReward: func(session string) bool {
			var obj User
			if err := db.Where("hash = ?", session).Find(&obj).Error; err != nil {
				return false
			}
			if obj.ID > 0 {
				return true
			}
			return false
		},
		Login: func(session, tg string) error {
			var obj User
			if err := db.Where("hash = ?", session).Find(&obj).Error; err != nil {
				return err
			}
			if obj.ID > 0 {
				obj.Telegram = tg
				if err := db.Save(&obj).Error; err != nil {
					return err
				}
			}
			return nil
		},
		GetRewardPrice: func(session, hash string) (Price, error) {
			var obj User
			//TODO for result надо засунуть интерфейс
			//Encoding, err := base64.StdEncoding.DecodeString(hash)
			/*if err != nil {
				return Price{}, err
			}*/
			if err := db.Preload("Prices").Where("hash = ?", session).Find(&obj).Error; err != nil {
				return Price{}, err
			}
			for _, price := range obj.Prices {
				if price.Hash == hash {
					return price, nil
				}
			}
			return Price{}, errors.New("none price")
		},
		GetAll: func() []User {
			var users []User
			db.Preload("Prices").Preload("Ip").Find(&users)
			return users
		},
		GetAllIps: func(uid uint64) ([]Ip, error) {
			var user User
			user.ID = uid
			if err := db.Preload("Ip").Find(&user).Error; err != nil {
				return nil, err
			}
			return user.Ip, nil
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
