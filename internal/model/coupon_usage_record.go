package model

import "time"

type CouponUsageRecord struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CouponID   uint64    `gorm:"not null;index" json:"coupon_id"`
	UserPhone  string    `gorm:"type:varchar(20);not null;index" json:"user_phone"`
	StoreID    uint64    `gorm:"not null;index" json:"store_id"`
	Action     string    `gorm:"type:enum('consume','refund','expire','freeze','unfreeze');not null" json:"action"`
	OrderInfo  *JSON     `gorm:"type:json;default:null" json:"order_info"`
	Operator   string    `gorm:"type:varchar(64)" json:"operator"`
	IPAddress  string    `gorm:"type:varchar(45)" json:"ip_address"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (CouponUsageRecord) TableName() string { return "coupon_usage_records" }
