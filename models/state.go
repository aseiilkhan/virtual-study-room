// models/user.go
package models

import "gorm.io/gorm"

type State struct {
	gorm.Model
	State string `json:"state" gorm:"unique"`
	Email string `json:"email" gorm:"unique"`
}
