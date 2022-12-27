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
)

type Price struct {
	axcrudobject.Model
	Type      PriceType `json:"type"`
	Data      string    `json:"data"`
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

func (u *Price) BeforeCreate(tx *gorm.DB) (err error) {
	u.Date = time.Now()
	return
}

func (u *Price) AfterFind(tx *gorm.DB) (err error) {
	u.Win = u.Type != NonePrice
	return
}

type PriceController struct {
	Generate func() Price
	GetPromo func(key string) (PromoBase, error)
	GetPrice func(key string) (PriceBase, error)
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
			obj.Key = base64.StdEncoding.EncodeToString([]byte(str))
			obj.Data = s
			obj.Path = filepath.Base(path)
			db.Create(&obj)
		}
		return nil
	})
	for i := 0; i < 100; i++ {
		var obj PromoBase
		obj.Key = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d-promo-base", i)))
		obj.Data = "HNY23-15"
		db.Create(&obj)
	}
	return PriceController{
		Generate: func() Price {
			var obj Price
			obj.Type = randPrice()
			switch obj.Type {
			case NonePrice:
				obj.Data = "ТЫ НЕ ВЫИГРАЛ"
				break
			case Promo:
				obj.Data = "HNY23-15"
				break
			case Present:
				var allObj []PriceBase
				var objP PriceBase
				db.Find(&allObj)
				if len(allObj) > 0 {
					rand.Seed(time.Now().Unix())
					objP = allObj[rand.Intn(len(allObj))]
				}
				obj.Data = objP.Key
				break
			}
			return obj
		},
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
			obj.Data = "ТЫ НЕ ВЫИГРАЛ"
			break
		}
		objP.Win = true
		db.Save(&objP)
		obj.Data = objP.Key
		break
	case Present:
		var allObj []PriceBase
		var objP PriceBase
		db.Where("win = ?", false).Find(&allObj)
		if len(allObj) > 0 {
			rand.Seed(time.Now().Unix())
			objP = allObj[rand.Intn(len(allObj))]
		} else {
			obj.Data = "ТЫ НЕ ВЫИГРАЛ"
			break
		}
		objP.Win = true
		db.Save(&objP)
		obj.Data = objP.Key
		break
	}
	return obj
}

func randPrice() PriceType {
	rand.Seed(time.Now().Unix())
	v := randInt(1, 100)
	if v > 0 && v <= 50 {
		return NonePrice
	} else if v > 50 && v <= 75 {
		return Promo
	} else {
		return Present
	}
}
