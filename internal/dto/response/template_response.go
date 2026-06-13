package response

import (
	"encoding/json"
	"hxcoupon/internal/model"
	"time"
)

type TemplateResponse struct {
	ID                 uint64          `json:"id"`
	Name               string          `json:"name"`
	Type               string          `json:"type"`
	DiscountValue      float64         `json:"discount_value"`
	ThresholdAmount    float64         `json:"threshold_amount"`
	ApplicableScope    string          `json:"applicable_scope"`
	StoreIDs           []uint64        `json:"store_ids,omitempty"`
	Stackable          bool            `json:"stackable"`
	MaxStackCount      uint8           `json:"max_stack_count"`
	ValidityType       string          `json:"validity_type"`
	ValidityDays       uint            `json:"validity_days"`
	ValidStart         *time.Time      `json:"valid_start"`
	ValidEnd           *time.Time      `json:"valid_end"`
	TotalQuantity      uint            `json:"total_quantity"`
	IssuedCount        uint            `json:"issued_count"`
	PerUserLimit       uint            `json:"per_user_limit"`
	ProductRestriction json.RawMessage `json:"product_restriction"`
	Status             int8            `json:"status"`
	MpAppID            string          `json:"mp_appid,omitempty"`
	MpPagePath         string          `json:"mp_page_path,omitempty"`
	StoreName          string          `json:"store_name,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
}

func ToTemplateResponse(t *model.CouponTemplate, storeIDs []uint64) *TemplateResponse {
	var pr json.RawMessage
	if t.ProductRestriction != nil {
		pr = json.RawMessage(*t.ProductRestriction)
	}
	return &TemplateResponse{
		ID:                 t.ID,
		Name:               t.Name,
		Type:               t.Type,
		DiscountValue:      t.DiscountValue,
		ThresholdAmount:    t.ThresholdAmount,
		ApplicableScope:    t.ApplicableScope,
		StoreIDs:           storeIDs,
		Stackable:          t.Stackable,
		MaxStackCount:      t.MaxStackCount,
		ValidityType:       t.ValidityType,
		ValidityDays:       t.ValidityDays,
		ValidStart:         t.ValidStart,
		ValidEnd:           t.ValidEnd,
		TotalQuantity:      t.TotalQuantity,
		IssuedCount:        t.IssuedCount,
		PerUserLimit:       t.PerUserLimit,
		ProductRestriction: pr,
		Status:             t.Status,
		CreatedAt:          t.CreatedAt,
	}
}
