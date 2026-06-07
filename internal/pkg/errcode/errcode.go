package errcode

const (
	Success          = 0
	InvalidParams    = 40001
	AuthFailed       = 40100
	StoreNotAuth     = 40101
	Forbidden        = 40300
	NotFound         = 40400
	Conflict         = 40900
	InternalError    = 50000
	CouponExpired    = 60001
	CouponUsed       = 60002
	CouponNotApply   = 60003
	NoInventory      = 60004
	PerUserLimit     = 60005
	BelowThreshold   = 60006
	RefundMismatch   = 60007
	RateLimited      = 60008
)

var messages = map[int]string{
	Success:          "success",
	InvalidParams:    "invalid parameters",
	AuthFailed:       "authentication failed",
	StoreNotAuth:     "store not authorized",
	Forbidden:        "forbidden",
	NotFound:         "resource not found",
	Conflict:         "conflict",
	InternalError:    "internal server error",
	CouponExpired:    "coupon has expired",
	CouponUsed:       "coupon already used",
	CouponNotApply:   "coupon not applicable to this store",
	NoInventory:      "insufficient inventory",
	PerUserLimit:     "per-user limit reached",
	BelowThreshold:   "order amount below threshold",
	RefundMismatch:   "refund order mismatch",
	RateLimited:      "rate limit exceeded",
}

func Message(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "unknown error"
}
