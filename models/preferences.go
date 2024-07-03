package models

import "gorm.io/gorm"

type Preferences struct {
	gorm.Model
	UserID uint   `json:"userId" gorm:"uniqueIndex"`
	Theme  string `json:"theme"`
	Layout string `json:"layout"`
}
