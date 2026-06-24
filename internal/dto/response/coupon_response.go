package response

import (
	"hxcoupon/internal/model"
	"time"
)

type CouponIssueResponse struct {
	CouponID        uint64    `json:"coupon_id"`
	CouponCode      string    `json:"coupon_code"`
	TemplateName    string    `json:"template_name"`
	Type            string    `json:"type"`
	DiscountValue   float64   `json:"discount_value"`
	ThresholdAmount float64   `json:"threshold_amount"`
	ValidStart      time.Time `json:"valid_start"`
	ValidEnd        time.Time `json:"valid_end"`
	Status          string    `json:"status"`
	QrCodeURL       string    `json:"qr_code_url"`
}

type CouponAvailableResponse struct {
	CouponID        uint64    `json:"coupon_id"`
	CouponCode      string    `json:"coupon_code"`
	TemplateID      uint64    `json:"template_id"`
	TemplateName    string    `json:"template_name"`
	Type            string    `json:"type"`
	DiscountValue   float64   `json:"discount_value"`
	ThresholdAmount float64   `json:"threshold_amount"`
	ValidEnd        time.Time `json:"valid_end"`
	Stackable       bool      `json:"stackable"`
	MpAppID         string    `json:"mp_appid,omitempty"`
	MpPagePath      string    `json:"mp_page_path,omitempty"`
	QrCodeURL       string    `json:"qr_code_url"`
}

type CouponConsumeResponse struct {
	CouponID      uint64    `json:"coupon_id"`
	CouponCode    string    `json:"coupon_code"`
	DiscountValue float64   `json:"discount_value"`
	ActualAmount  float64   `json:"actual_amount"`
	UsedAt        time.Time `json:"used_at"`
}

type CouponRefundResponse struct {
	CouponID   uint64 `json:"coupon_id"`
	CouponCode string `json:"coupon_code"`
	NewStatus  string `json:"new_status"`
	Restored   bool   `json:"restored"`
}

type CouponDetailResponse struct {
	CouponID        uint64                   `json:"coupon_id"`
	CouponCode      string                   `json:"coupon_code"`
	TemplateName    string                   `json:"template_name"`
	Type            string                   `json:"type"`
	DiscountValue   float64                  `json:"discount_value"`
	ThresholdAmount float64                  `json:"threshold_amount"`
	Status          string                   `json:"status"`
	UserPhone       string                   `json:"user_phone"`
	SourceStoreName string                   `json:"source_store_name"`
	ValidStart      time.Time                `json:"valid_start"`
	ValidEnd        time.Time                `json:"valid_end"`
	ReceiveTime     time.Time                `json:"receive_time"`
	UseTime         *time.Time               `json:"use_time"`
	UsedAtStoreName string                   `json:"used_at_store_name"`
	UseOrderID      string                   `json:"use_order_id"`
	MpAppID         string                   `json:"mp_appid,omitempty"`
	MpPagePath      string                   `json:"mp_page_path,omitempty"`
	QrCodeURL       string                   `json:"qr_code_url"`
	Records         []CouponUsageRecordBrief `json:"records"`
}

type CouponUsageRecordBrief struct {
	Action    string    `json:"action"`
	StoreName string    `json:"store_name"`
	CreatedAt time.Time `json:"created_at"`
}

func ToCouponDetailResponse(ci *model.CouponInstance, templateName, sourceStoreName, usedStoreName string, records []CouponUsageRecordBrief) *CouponDetailResponse {
	return &CouponDetailResponse{
		CouponID:        ci.ID,
		CouponCode:      ci.CouponCode,
		TemplateName:    templateName,
		Type:            "",
		DiscountValue:   0,
		Status:          ci.Status,
		UserPhone:       ci.UserPhone,
		SourceStoreName: sourceStoreName,
		ValidStart:      ci.ValidStart,
		ValidEnd:        ci.ValidEnd,
		ReceiveTime:     ci.ReceiveTime,
		UseTime:         ci.UseTime,
		UsedAtStoreName: usedStoreName,
		UseOrderID:      ci.UseOrderID,
		Records:         records,
	}
}
