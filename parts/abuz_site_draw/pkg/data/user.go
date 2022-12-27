package data

import (
	"abuz_site_draw/shared/axcrudobject"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"time"
)

type Ip struct {
	axcrudobject.Model
	Address   string    `json:"address"`
	Location  string    `json:"location"`
	Date      time.Time `json:"-"`
	ParseDate string    `json:"date"`
	UserRefer uint      `json:"-"`
}

func (i *Ip) BeforeCreate(tx *gorm.DB) (err error) {
	var objIp map[string]interface{}
	i.Date = time.Now()
	href := "http://ip-api.com/json/" + i.Address
	resp, err := http.Get(href)
	if err != nil {
		log.Error().Err(err)
		return err
	}
	bodyIp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("read body")
		return err
	}
	if err := json.Unmarshal(bodyIp, &objIp); err != nil {
		log.Error().Err(err).Msg("decode json")
		return err
	}
	if objIp["status"] == "fail" {
		return
	}
	i.Location = objIp["country"].(string)
	return
}

func (i *Ip) AfterFind(tx *gorm.DB) (err error) {
	i.ParseDate = fmt.Sprintf("%02d.%02d.%d", i.Date.Day(), i.Date.Month(), i.Date.Year())
	return
}

type User struct {
	axcrudobject.Model
	Ip       []Ip      `gorm:"foreignKey:UserRefer" json:"ip"`
	Telegram string    `json:"telegram"`
	Prices   []Price   `gorm:"foreignKey:UserRefer" json:"prices"`
	Hash     []Session `gorm:"foreignKey:UserRefer" json:"hash"`
	Admin    bool      `json:"admin"`
	Logined  bool      `gorm:"-"`
	Timer    time.Time `json:"timer"`
	Can      bool      `gorm:"-"`
	Ban      bool      `json:"ban"`
}

var Admins = []string{"nikolay35977", "ddduduru", "sempai6", "fack_system", "dappless", "m1o2u3s4e5", "MakerD"}

const TIME = 24

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
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
	GetById        func(uid uint64) (User, error)
	Block          func(string) error
	UnBlock        func(string) error
}

func NewUserController(db *gorm.DB, baseLog zerolog.Logger) UserController {
	log := baseLog.With().Str("model", "user").Logger()
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	if err := db.AutoMigrate(&Ip{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	if err := db.AutoMigrate(&Session{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	return UserController{
		CreateSession: func(session string) error {
			var obj Session
			obj.Hash = session
			db.Create(&obj)
			return nil
		},
		Get: func(session string) (User, error) {
			var obj User
			var o Session
			if err := db.Where("hash = ?", session).Find(&o).Error; err != nil {
				return User{}, err
			}
			if err := db.Preload("Prices").Where("id = ?", o.UserRefer).Find(&obj).Error; err != nil {
				return User{}, err
			}
			return obj, nil
		},
		StartGame: func(ip, session string) error {
			var obj User
			var o Session
			if err := db.Where("hash = ?", session).Find(&o).Error; err != nil {
				return err
			}
			if err := db.Where("id = ?", o.UserRefer).Find(&obj).Error; err != nil {
				return err
			}
			if !obj.Logined {
				return errors.New("need to login")
			}
			if obj.Timer.Before(time.Now()) {
				obj.Timer = time.Now().Add(time.Hour * TIME)
				var ipObj Ip
				ipObj.Address = ip
				ipObj.UserRefer = uint(obj.ID)
				if err := db.Create(&ipObj).Error; err != nil {
					return err
				}
				pricesObj := TestGeneratePrice()
				pricesObj.UserRefer = uint(obj.ID)
				if err := db.Create(&pricesObj).Error; err != nil {
					return err
				}
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
			var o Session
			if err := db.Where("hash = ?", session).Find(&o).Error; err != nil {
				return false
			}
			if err := db.Where("id = ?", o.UserRefer).Find(&obj).Error; err != nil {
				return false
			}
			if obj.ID > 0 {
				return true
			}
			return false
		},
		Login: func(session, tg string) error {
			var obj User
			var o Session
			if err := db.Where("hash = ?", session).Find(&o).Error; err != nil {
				return err
			}
			if o.UserRefer != 0 {
				return nil
			}
			obj.Telegram = tg
			if err := db.Create(&obj).Error; err != nil {
				return err
			}
			o.UserRefer = uint(obj.ID)
			if err := db.Save(&o).Error; err != nil {
				return err
			}
			return nil
		},
		Block: func(tg string) error {
			var obj User
			if err := db.Where("telegram = ?", tg).Find(&obj).Error; err != nil {
				return err
			}
			obj.Ban = true
			if obj.ID > 0 {
				if err := db.Save(&obj).Error; err != nil {
					return err
				}
			}
			return nil
		},
		UnBlock: func(tg string) error {
			var obj User
			if err := db.Where("telegram = ?", tg).Find(&obj).Error; err != nil {
				return err
			}
			obj.Ban = false
			if obj.ID > 0 {
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
			var o Session
			if err := db.Where("hash = ?", session).Find(&o).Error; err != nil {
				return Price{}, err
			}
			if err := db.Preload("Prices").Where("id = ?", o.UserRefer).Find(&obj).Error; err != nil {
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
		GetById: func(uid uint64) (User, error) {
			var obj User
			obj.ID = uid
			if err := db.Preload("Prices").Find(&obj).Error; err != nil {
				return User{}, err
			}
			return obj, nil
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
