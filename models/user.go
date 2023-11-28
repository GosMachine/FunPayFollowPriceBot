package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID         uint         `gorm:"primaryKey"`
	TelegramID int64        `gorm:"unique"`
	RefreshKD  int          `json:"refresh_kd"`
	AllLots    []AllLots    `gorm:"foreignKey:user_id" json:"-"`
	ActiveLots []ActiveLots `gorm:"foreignKey:user_id" json:"-"`
	Premium    bool
	admin      bool
}
