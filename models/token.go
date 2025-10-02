package models

import "time"

type RefreshToken struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	Token      string    `gorm:"unique;not null" json:"token"`
	ExpiryDate time.Time `gorm:"not null" json:"expiry_date"`
}
