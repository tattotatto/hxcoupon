package model

import "time"

type CouponTemplateStore struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TemplateID uint64    `gorm:"not null;uniqueIndex:uk_template_store" json:"template_id"`
	StoreID    uint64    `gorm:"not null;uniqueIndex:uk_template_store;index" json:"store_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (CouponTemplateStore) TableName() string { return "coupon_template_stores" }
