package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"time"

	"hxcoupon/config"
	"hxcoupon/internal/dto/response"
	redisutil "hxcoupon/internal/pkg/redis"

	"github.com/gin-gonic/gin"
)

func StoreAuth(cfg config.StoreAuthConfig, credentialGetter func(appKey string) (storeID uint64, appSecret string, err error)) gin.HandlerFunc {
	tolerance := time.Duration(cfg.TimestampTolerance) * time.Second

	return func(c *gin.Context) {
		appKey := c.GetHeader("X-App-Key")
		timestampStr := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")
		signature := c.GetHeader("X-Signature")

		if appKey == "" || timestampStr == "" || nonce == "" || signature == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Error(40100, "missing auth headers"))
			return
		}

		// Verify timestamp within tolerance
		ts, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Error(40100, "invalid timestamp"))
			return
		}
		now := time.Now().Unix()
		if abs(now-ts) > int64(tolerance.Seconds()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Error(40100, "timestamp out of range"))
			return
		}

		// Verify nonce not replayed (Redis)
		nonceKey := "nonce:" + appKey + ":" + nonce
		ctx := c.Request.Context()
		ok, err := redisutil.Client.SetNX(ctx, nonceKey, "1", 5*time.Minute).Result()
		if err == nil && !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Error(40100, "duplicate nonce"))
			return
		}
		// If Redis is down (err != nil), continue without nonce check (graceful degradation)

		// Read body for signature verification
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Error(50000, "read body failed"))
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Get credentials
		storeID, appSecret, err := credentialGetter(appKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Error(40101, "invalid app key"))
			return
		}

		// Verify signature
		signingStr := c.Request.Method + "\n" + c.Request.URL.Path + "\n" + timestampStr + "\n" + nonce + "\n" + string(bodyBytes)
		expectedSig := computeHMAC(appSecret, signingStr)

		if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSig)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Error(40100, "invalid signature"))
			return
		}

		c.Set("store_id", storeID)
		c.Next()
	}
}

func computeHMAC(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
