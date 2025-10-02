package models

type ApiKey struct {
	Key      string `gorm:"primaryKey;not null" json:"key"`
	UserID   uint   `gorm:"not null" json:"user_id"`
	Client   string `gorm:"not null" json:"client"`
	Scope    string
	IsActive bool `gorm:"default:true" json:"is_active"`
}
