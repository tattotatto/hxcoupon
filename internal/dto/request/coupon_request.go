package request

// AdminIssueCouponRequest for admin-side coupon issuance (includes store_id)
type AdminIssueCouponRequest struct {
	StoreID    uint64 `json:"store_id" validate:"required"`
	TemplateID uint64 `json:"template_id" validate:"required"`
	UserPhone  string `json:"user_phone" validate:"required,max=20"`
}

type IssueCouponRequest struct {
	TemplateID uint64 `json:"template_id" validate:"required"`
	UserPhone  string `json:"user_phone" validate:"required,max=20"`
}

type ConsumeCouponRequest struct {
	CouponCode  string  `json:"coupon_code" validate:"required,max=64"`
	UserPhone   string  `json:"user_phone" validate:"required,max=20"`
	StoreID     uint64  `json:"store_id" validate:"required"`
	OrderID     string  `json:"order_id" validate:"required,max=64"`
	OrderAmount float64 `json:"order_amount" validate:"required,gt=0"`
}

type RefundCouponRequest struct {
	CouponCode string `json:"coupon_code" validate:"required,max=64"`
	UserPhone  string `json:"user_phone" validate:"required,max=20"`
	StoreID    uint64 `json:"store_id" validate:"required"`
	OrderID    string `json:"order_id" validate:"required,max=64"`
}

type AvailableCouponRequest struct {
	Pagination
	UserPhone   string  `json:"user_phone" form:"user_phone" validate:"required"`
	StoreID     uint64  `json:"store_id" form:"store_id" validate:"required"`
	OrderAmount float64 `json:"order_amount" form:"order_amount" validate:"gte=0"`
}

type UserCouponRequest struct {
	Pagination
	UserPhone string `json:"user_phone" form:"user_phone" validate:"required"`
	Status    string `json:"status" form:"status"`
}
