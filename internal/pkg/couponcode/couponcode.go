package couponcode

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const (
	charset       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	universalCode = "TY000"
	dateFmt       = "060102"
	randomLen     = 6
)

// Generate creates a meaningful coupon code.
// storeCode: 5-char store code for specific coupons, empty for universal.
func Generate(storeCode string) string {
	prefix := universalCode
	if storeCode != "" {
		prefix = storeCode
	}
	dateStr := time.Now().Format(dateFmt)
	randomStr := randomString(randomLen)
	return fmt.Sprintf("%s%s%s", prefix, dateStr, randomStr)
}

func randomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to deterministic but acceptable
			result[i] = charset[i%len(charset)]
			continue
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}
