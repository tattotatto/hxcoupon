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
	Success:          "成功",
	InvalidParams:    "参数无效",
	AuthFailed:       "认证失败",
	StoreNotAuth:     "门店未授权",
	Forbidden:        "无权限",
	NotFound:         "资源不存在",
	Conflict:         "数据冲突",
	InternalError:    "服务器内部错误",
	CouponExpired:    "优惠券已过期",
	CouponUsed:       "优惠券已使用",
	CouponNotApply:   "优惠券不适用于该门店",
	NoInventory:      "库存不足",
	PerUserLimit:     "已达单人领取上限",
	BelowThreshold:   "订单金额未达门槛",
	RefundMismatch:   "退款订单不匹配",
	RateLimited:      "请求频率超限",
}

func Message(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "unknown error"
}
