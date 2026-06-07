package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type CouponTemplate struct {
	ID                 uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name               string          `gorm:"type:varchar(128);not null" json:"name"`
	Type               string          `gorm:"type:enum('full_reduction','discount','fixed_amount');not null" json:"type"`
	DiscountValue      float64         `gorm:"type:decimal(10,2);not null" json:"discount_value"`
	ThresholdAmount    float64         `gorm:"type:decimal(10,2);not null;default:0" json:"threshold_amount"`
	ApplicableScope    string          `gorm:"type:enum('all','specific');not null;default:'all'" json:"applicable_scope"`
	Stackable          bool            `gorm:"not null;default:false" json:"stackable"`
	MaxStackCount      uint8           `gorm:"default:1" json:"max_stack_count"`
	ValidityType       string          `gorm:"type:enum('fixed_date','days_after_receive');not null" json:"validity_type"`
	ValidityDays       uint            `gorm:"default:null" json:"validity_days"`
	ValidStart         *time.Time      `gorm:"default:null" json:"valid_start"`
	ValidEnd           *time.Time      `gorm:"default:null" json:"valid_end"`
	TotalQuantity      uint            `gorm:"default:0" json:"total_quantity"`
	IssuedCount        uint            `gorm:"not null;default:0" json:"issued_count"`
	PerUserLimit       uint            `gorm:"default:1" json:"per_user_limit"`
	ProductRestriction *JSON           `gorm:"type:json;default:null" json:"product_restriction"`
	Status             int8            `gorm:"not null;default:0;index" json:"status"`
	CreatedBy          string          `gorm:"type:varchar(64)" json:"created_by"`
	CreatedAt          time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (CouponTemplate) TableName() string { return "coupon_templates" }

// JSON is a generic JSON field for GORM.
type JSON json.RawMessage

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("JSON scan: type assertion to []byte failed")
	}
	*j = make(JSON, len(bytes))
	copy(*j, bytes)
	return nil
}

func (j *JSON) MarshalJSON() ([]byte, error) {
	if j == nil || len(*j) == 0 {
		return []byte("null"), nil
	}
	return *j, nil
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSON: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}
