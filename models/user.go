package models

type User struct {
	ID         uint         `gorm:"primaryKey"`
	TelegramID int          `gorm:"unique"`
	RefreshKD  int          `json:"refresh_kd"`
	AllLots    []AllLots    `gorm:"foreignKey:user_id" json:"-"`
	ActiveLots []ActiveLots `gorm:"foreignKey:user_id" json:"-"`
	Lang       string       `json:"lang"`
	Premium    bool
}
