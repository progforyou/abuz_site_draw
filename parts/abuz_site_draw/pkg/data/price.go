package data

import (
	"bot_tasker/shared/axcrudobject"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

type PriceType int

const (
	NonePrice PriceType = 0
	Promo     PriceType = 1
	Sale      PriceType = 2
)

type Price struct {
	axcrudobject.Model
	Type      PriceType `json:"type"`
	Data      string    `json:"data"`
	Date      time.Time `json:"date"`
	UserRefer uint      `gorm:"primaryKey" json:"-"`
}

func (u *Price) BeforeCreate(tx *gorm.DB) (err error) {
	u.Date = time.Now()
	return
}

type PriceController struct {
	Generate func() Price
}

func NewPriceController(db *gorm.DB, baseLog zerolog.Logger) PriceController {
	log := baseLog.With().Str("model", "price").Logger()
	if err := db.AutoMigrate(&Price{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
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
				obj.Data = "NY1"
				break
			case Sale:
				obj.Data = "20%"
				break
			}
			return obj
		}}
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func TestGeneratePrice() Price {
	var obj Price
	obj.Type = randPrice()
	switch obj.Type {
	case NonePrice:
		obj.Data = "ТЫ НЕ ВЫИГРАЛ"
		break
	case Promo:
		obj.Data = "NY1"
		break
	case Sale:
		obj.Data = "20%"
		break
	}
	return obj
}

func randPrice() PriceType {
	v := randInt(1, 100)
	if v > 0 && v < 33 {
		return NonePrice
	} else if v > 33 && v < 66 {
		return Promo
	} else {
		return Sale
	}
}
