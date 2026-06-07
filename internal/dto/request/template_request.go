package request

type CreateTemplateRequest struct {
	Name               string    `json:"name" validate:"required,max=128"`
	Type               string    `json:"type" validate:"required,oneof=full_reduction discount fixed_amount"`
	DiscountValue      float64   `json:"discount_value" validate:"required,gt=0"`
	ThresholdAmount    float64   `json:"threshold_amount" validate:"gte=0"`
	ApplicableScope    string    `json:"applicable_scope" validate:"required,oneof=all specific"`
	StoreIDs           []uint64  `json:"store_ids"`
	Stackable          bool      `json:"stackable"`
	MaxStackCount      uint8     `json:"max_stack_count" validate:"gte=1"`
	ValidityType       string    `json:"validity_type" validate:"required,oneof=fixed_date days_after_receive"`
	ValidityDays       uint      `json:"validity_days"`
	ValidStart         string    `json:"valid_start"`
	ValidEnd           string    `json:"valid_end"`
	TotalQuantity      uint      `json:"total_quantity" validate:"gte=0"`
	PerUserLimit       uint      `json:"per_user_limit" validate:"gte=1"`
	ProductRestriction string    `json:"product_restriction"`
}

type UpdateTemplateRequest struct {
	Name               string  `json:"name" validate:"required,max=128"`
	DiscountValue      float64 `json:"discount_value" validate:"required,gt=0"`
	ThresholdAmount    float64 `json:"threshold_amount" validate:"gte=0"`
	Stackable          bool    `json:"stackable"`
	MaxStackCount      uint8   `json:"max_stack_count" validate:"gte=1"`
	ValidityDays       uint    `json:"validity_days"`
	ValidStart         string  `json:"valid_start"`
	ValidEnd           string  `json:"valid_end"`
	TotalQuantity      uint    `json:"total_quantity" validate:"gte=0"`
	PerUserLimit       uint    `json:"per_user_limit" validate:"gte=1"`
	ProductRestriction string  `json:"product_restriction"`
}

type UpdateTemplateStatusRequest struct {
	Status int8 `json:"status" validate:"oneof=0 1 2"`
}

type TemplateListRequest struct {
	Pagination
	Keyword string `json:"keyword" form:"keyword"`
	Type    string `json:"type" form:"type"`
	Status  *int8  `json:"status" form:"status"`
}
