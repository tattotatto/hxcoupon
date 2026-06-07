package model

import "time"

type StoreAPICredential struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID   uint64    `gorm:"not null;index" json:"store_id"`
	AppKey    string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"app_key"`
	AppSecret string    `gorm:"type:varchar(256);not null" json:"-"`
	Status    int8      `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (StoreAPICredential) TableName() string { return "store_api_credentials" }
