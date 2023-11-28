package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID         uint  `gorm:"primaryKey"`
	TelegramID int64 `gorm:"unique"`
	RefreshKD  int   `json:"refresh_kd"`
	Premium    bool
	Admin      bool
}
