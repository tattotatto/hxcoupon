package model

import "time"

type AdminUser struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"type:varchar(256);not null" json:"-"`
	Role         string     `gorm:"type:enum('super_admin','admin','viewer');not null;default:'admin'" json:"role"`
	Status       int8       `gorm:"not null;default:1" json:"status"`
	LastLoginAt  *time.Time `gorm:"default:null" json:"last_login_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AdminUser) TableName() string { return "admin_users" }
