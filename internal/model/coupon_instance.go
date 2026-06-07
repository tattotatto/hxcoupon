package model

import "time"

type CouponInstance struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TemplateID      uint64     `gorm:"not null;index" json:"template_id"`
	UserPhone       string     `gorm:"type:varchar(20);not null;index:idx_user_phone_status" json:"user_phone"`
	SourceStoreID   uint64     `gorm:"not null;index" json:"source_store_id"`
	CouponCode      string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"coupon_code"`
	Status          string     `gorm:"type:enum('unused','used','expired','frozen');not null;default:'unused';index:idx_user_phone_status" json:"status"`
	ReceiveTime     time.Time  `gorm:"not null" json:"receive_time"`
	UseTime         *time.Time `gorm:"default:null" json:"use_time"`
	UsedAtStoreID   *uint64    `gorm:"default:null;index" json:"used_at_store_id"`
	UseOrderID      string     `gorm:"type:varchar(64);default:null" json:"use_order_id"`
	UseOrderAmount  float64    `gorm:"type:decimal(10,2);default:null" json:"use_order_amount"`
	ValidStart      time.Time  `gorm:"not null" json:"valid_start"`
	ValidEnd        time.Time  `gorm:"not null;index:idx_valid_end" json:"valid_end"`
	IdempotencyKey  string     `gorm:"type:varchar(128);default:null;uniqueIndex:uk_idempotency" json:"idempotency_key"`
	Version         uint       `gorm:"not null;default:0" json:"version"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (CouponInstance) TableName() string { return "coupon_instances" }
