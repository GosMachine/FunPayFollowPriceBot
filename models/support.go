package models

type Support struct {
	ID         uint `gorm:"primaryKey"`
	TelegramID int  `json:"TelegramID"`
	Message    string
}
