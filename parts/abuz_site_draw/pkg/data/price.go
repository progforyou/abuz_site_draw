package data

import (
	"abuz_site_draw/shared/axcrudobject"
	"encoding/base64"
	"fmt"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

type PriceType int

const (
	NonePrice PriceType = 0
	Promo     PriceType = 1
	Present   PriceType = 2
	Money5    PriceType = 3
	Money10   PriceType = 4
	Money100  PriceType = 5
)

type Price struct {
	axcrudobject.Model
	Type      PriceType `json:"type"`
	Data      string    `json:"data"`
	Hash      string    `json:"hash"`
	Date      time.Time `json:"date"`
	Win       bool      `gorm:"-" json:"win"`
	UserRefer uint      `json:"-"`
}

type PriceBase struct {
	axcrudobject.Model
	Win  bool   `json:"win"`
	Data string `json:"data"`
	Path string `json:"path"`
	Key  string `gorm:"unique" json:"key"`
}

type PromoBase struct {
	axcrudobject.Model
	Win  bool   `json:"win"`
	Data string `json:"data"`
	Key  string `gorm:"unique" json:"key"`
}

type Money5Base struct {
	axcrudobject.Model
	Win  bool   `json:"win"`
	Data string `json:"data"`
	Key  string `gorm:"unique" json:"key"`
}

type Money10Base struct {
	axcrudobject.Model
	Win  bool   `json:"win"`
	Data string `json:"data"`
	Key  string `gorm:"unique" json:"key"`
}

type Money100Base struct {
	axcrudobject.Model
	Win  bool   `json:"win"`
	Data string `json:"data"`
	Key  string `gorm:"unique" json:"key"`
}

func (u *Price) BeforeCreate(tx *gorm.DB) (err error) {
	u.Date = time.Now()
	return
}

func (u *Price) AfterFind(tx *gorm.DB) (err error) {
	u.Win = u.Type != NonePrice
	return
}

type PriceController struct {
	GetPromo    func(key string) (PromoBase, error)
	GetPrice    func(key string) (PriceBase, error)
	GetMoney5   func(key string) (Money5Base, error)
	GetMoney10  func(key string) (Money10Base, error)
	GetMoney100 func(key string) (Money100Base, error)
}

func NewPriceController(db *gorm.DB, baseLog zerolog.Logger) PriceController {
	log := baseLog.With().Str("model", "price").Logger()
	if err := db.AutoMigrate(&Price{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	if err := db.AutoMigrate(&PriceBase{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	if err := db.AutoMigrate(&PromoBase{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	if err := db.AutoMigrate(&Money5Base{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	if err := db.AutoMigrate(&Money10Base{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	if err := db.AutoMigrate(&Money100Base{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	//PriceBase
	filepath.Walk("parts/abuz_site_draw/pkg/data/prices", func(path string, info fs.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		for i, s := range strings.Split(string(content), "\n") {
			if s == "" || s == " " {
				continue
			}
			var obj PriceBase
			str := fmt.Sprintf("%s-%d", filepath.Base(path), i)
			obj.Key = strings.Replace(base64.StdEncoding.EncodeToString([]byte(str)), "/", "H", -1)
			obj.Data = s
			obj.Path = filepath.Base(path)
			db.Create(&obj)
		}
		return nil
	})

	//PromoBase
	for i := 0; i < 100; i++ {
		var obj PromoBase
		obj.Key = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d-promo-base", i)))
		obj.Data = "HNY23-15"
		db.Create(&obj)
	}
	//Money5Base
	for i := 0; i < 35; i++ {
		var obj Money5Base
		obj.Key = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d-money5-base", i)))
		obj.Data = "https://t.me/CryptoBot?start=CQQIpZaylq1x"
		db.Create(&obj)
	}
	for i := 0; i < 15; i++ {
		var obj Money5Base
		obj.Key = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d-money5-base", i)))
		obj.Data = "https://t.me/CryptoBot?start=CQexvub0Vp5r"
		db.Create(&obj)
	}
	//Money10Base
	for i := 0; i < 10; i++ {
		var obj Money10Base
		obj.Key = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d-money10-base", i)))
		obj.Data = "https://t.me/CryptoBot?start=CQcm2m1zdZN5"
		db.Create(&obj)
	}
	//Money100Base
	var obj Money100Base
	obj.Key = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d-money100-base", 0)))
	obj.Data = "https://t.me/CryptoBot?start=CQZoido1MzBn"
	db.Create(&obj)
	return PriceController{
		GetPromo: func(key string) (PromoBase, error) {
			var obj PromoBase
			if err := db.Where("key = ?", key).Find(&obj).Error; err != nil {
				return PromoBase{}, err
			}
			return obj, nil
		},
		GetPrice: func(key string) (PriceBase, error) {
			var obj PriceBase
			if err := db.Where("key = ?", key).Find(&obj).Error; err != nil {
				return PriceBase{}, err
			}
			return obj, nil
		},
		GetMoney5: func(key string) (Money5Base, error) {
			var obj Money5Base
			if err := db.Where("key = ?", key).Find(&obj).Error; err != nil {
				return Money5Base{}, err
			}
			return obj, nil
		},
		GetMoney10: func(key string) (Money10Base, error) {
			var obj Money10Base
			if err := db.Where("key = ?", key).Find(&obj).Error; err != nil {
				return Money10Base{}, err
			}
			return obj, nil
		},
		GetMoney100: func(key string) (Money100Base, error) {
			var obj Money100Base
			if err := db.Where("key = ?", key).Find(&obj).Error; err != nil {
				return Money100Base{}, err
			}
			return obj, nil
		}}
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func TestGeneratePrice(db *gorm.DB) Price {
	var obj Price
	obj.Type = randPrice()
	switch obj.Type {
	case NonePrice:
		obj.Data = "ТЫ НЕ ВЫИГРАЛ"
		break
	case Promo:
		var allObj []PromoBase
		var objP PromoBase
		db.Where("win = ?", false).Find(&allObj)
		if len(allObj) > 0 {
			rand.Seed(time.Now().Unix())
			objP = allObj[rand.Intn(len(allObj))]
		} else {
			obj.Type = NonePrice
			obj.Data = "ТЫ НЕ ВЫИГРАЛ"
			break
		}
		objP.Win = true
		db.Save(&objP)
		obj.Data = objP.Data
		obj.Hash = objP.Key
		break
	case Present:
		var allObj []PriceBase
		var objP PriceBase
		db.Where("win = ?", false).Find(&allObj)
		if len(allObj) > 0 {
			rand.Seed(time.Now().Unix())
			objP = allObj[rand.Intn(len(allObj))]
		} else {
			obj.Type = NonePrice
			obj.Data = "ТЫ НЕ ВЫИГРАЛ"
			break
		}
		objP.Win = true
		db.Save(&objP)
		obj.Data = objP.Data
		obj.Hash = objP.Key
		break
	case Money5:
		var allObj []Money5Base
		var objP Money5Base
		db.Where("win = ?", false).Find(&allObj)
		if len(allObj) > 0 {
			rand.Seed(time.Now().Unix())
			objP = allObj[rand.Intn(len(allObj))]
		} else {
			obj.Type = NonePrice
			obj.Data = "ТЫ НЕ ВЫИГРАЛ"
			break
		}
		objP.Win = true
		db.Save(&objP)
		obj.Data = objP.Data
		obj.Hash = objP.Key
		break
	case Money10:
		var allObj []Money10Base
		var objP Money10Base
		db.Where("win = ?", false).Find(&allObj)
		if len(allObj) > 0 {
			rand.Seed(time.Now().Unix())
			objP = allObj[rand.Intn(len(allObj))]
		} else {
			obj.Type = NonePrice
			obj.Data = "ТЫ НЕ ВЫИГРАЛ"
			break
		}
		objP.Win = true
		db.Save(&objP)
		obj.Data = objP.Data
		obj.Hash = objP.Key
		break
	case Money100:
		var allObj []Money100Base
		var objP Money100Base
		db.Where("win = ?", false).Find(&allObj)
		if len(allObj) > 0 {
			rand.Seed(time.Now().Unix())
			objP = allObj[rand.Intn(len(allObj))]
		} else {
			obj.Type = NonePrice
			obj.Data = "ТЫ НЕ ВЫИГРАЛ"
			break
		}
		objP.Win = true
		db.Save(&objP)
		obj.Data = objP.Data
		obj.Hash = objP.Key
		break
	}
	return obj
}

func randPrice() PriceType {
	rand.Seed(time.Now().Unix())
	v := randInt(1, 50000)
	if v == 50000 {
		return Money100
	} else if v >= 49990 && v < 50000 {
		return Money10
	} else if v >= 49940 && v < 49990 {
		return Money5
	} else if v >= 24940 && v < 49940 {
		return NonePrice
	} else if v >= 4440 && v < 24940 {
		return Present
	} else {
		return Promo
	}
}
