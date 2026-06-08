package model

import "time"

type AdminUser struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username        string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	PasswordHash    string     `gorm:"type:varchar(256);not null" json:"-"`
	Role            string     `gorm:"type:enum('super_admin','admin','member');not null;default:'member'" json:"role"`
	MemberType      *string    `gorm:"type:enum('issuer','consumer','both');default:null" json:"member_type"`
	Status          int8       `gorm:"not null;default:1" json:"status"`
	ApprovalStatus  int8       `gorm:"not null;default:0" json:"approval_status"`
	CompanyName     *string    `gorm:"type:varchar(128);default:null" json:"company_name"`
	ContactName     *string    `gorm:"type:varchar(64);default:null" json:"contact_name"`
	ContactPhone    *string    `gorm:"type:varchar(20);default:null" json:"contact_phone"`
	Email           *string    `gorm:"type:varchar(128);default:null" json:"email"`
	RejectReason    *string    `gorm:"type:varchar(512);default:null" json:"reject_reason"`
	RegisteredAt    *time.Time `gorm:"default:null" json:"registered_at"`
	ApprovedAt      *time.Time `gorm:"default:null" json:"approved_at"`
	ApprovedBy      *uint64    `gorm:"default:null" json:"approved_by"`
	BusinessLicense *string    `gorm:"type:varchar(256);default:null" json:"business_license"`
	LastLoginAt     *time.Time `gorm:"default:null" json:"last_login_at"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AdminUser) TableName() string { return "admin_users" }
