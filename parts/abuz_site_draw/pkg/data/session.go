package data

import "abuz_site_draw/shared/axcrudobject"

type Session struct {
	axcrudobject.Model
	Hash      string `gorm:"unique" json:"hash"`
	UserRefer uint   `json:"-"`
}
