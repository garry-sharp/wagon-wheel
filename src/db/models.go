package db

import "gorm.io/gorm"

type Price struct {
	gorm.Model
	Quote string  `json:"quote"`
	Asset string  `json:"asset"`
	Price float32 `json:"price"`
}
