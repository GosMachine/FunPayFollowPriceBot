package models

import "github.com/lib/pq"

type Lot struct {
	ID          uint    `gorm:"primaryKey"`
	Category    string  `json:"category"`
	Server      string  `json:"server"`
	Description string  `json:"description"`
	Side        string  `json:"side"`
	Seller      string  `json:"user"`
	Amount      string  `json:"amount"`
	Price       float64 `json:"price"`
}

type AllLots struct {
	ID       uint           `gorm:"primaryKey"`
	UserID   uint           `json:"UserID"`
	Lot      string         `json:"lot"`
	Servers  pq.StringArray `gorm:"type:text[]" json:"servers"`
	MaxPrice float64        `json:"maxPrice"`
}
type ActiveLots struct {
	ID       uint           `gorm:"primaryKey"`
	UserID   uint           `json:"UserID"`
	Lot      string         `json:"lot"`
	Servers  pq.StringArray `gorm:"type:text[]" json:"servers"`
	MaxPrice float64        `json:"maxPrice"`
}
