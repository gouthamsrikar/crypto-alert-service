package models

import (
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	Coin      string  `json:"coin"`
	Price     float64 `json:"price"`
	Status    string  `json:"status"`
	FcmID     string  `json:"fcmId"`
	Direction string  `json:"direction"`
	MA        int     `json:"ma"`
	Type      string  `json:"type"`
}
