package models

import "gorm.io/gorm"

type Support struct {
	gorm.Model
	ID         uint  `gorm:"primaryKey"`
	TelegramID int64 `json:"TelegramID"`
	Message    string
	Status     bool
}
