// models/user.go
package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email                 string `json:"email" gorm:"unique"`
	Password              string `json:"password"`
	SpotifyToken          string `json:"spotifyToken"`
	SpotifyRefreshToken   string `json:"spotifyRefreshToken"`
	SpotifyTokenExpiresAt int64  `json:"spotifyTokenExpiresAt"`
}
