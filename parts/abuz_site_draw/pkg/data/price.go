package data

import "math/rand"

type PriceType int

const (
	NonePrice PriceType = 0
	Promo     PriceType = 1
	Sale      PriceType = 2
)

type Price struct {
	Type PriceType `json:"type"`
	Data string    `json:"data"`
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
