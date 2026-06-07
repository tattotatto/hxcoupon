package model

import "time"

type Store struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"type:varchar(128);not null" json:"name"`
	Code         string    `gorm:"type:varchar(5);uniqueIndex;not null" json:"code"`
	AppID        string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"app_id"`
	Type         string    `gorm:"type:enum('miniprogram','h5');not null" json:"type"`
	Status       int8      `gorm:"not null;default:1" json:"status"`
	ContactName  string    `gorm:"type:varchar(64)" json:"contact_name"`
	ContactPhone string    `gorm:"type:varchar(20)" json:"contact_phone"`
	Remark       string    `gorm:"type:varchar(512)" json:"remark"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Store) TableName() string { return "stores" }
