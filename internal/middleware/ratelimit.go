package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mobarok/devflow/pkg/response"
	"github.com/redis/go-redis/v9"
)

// RateLimiter returns a Gin middleware that enforces a rate limit per client (IP or User ID)
// using atomic Redis INCR operations.
func RateLimiter(client *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.Next()
			return
		}

		// Identify client by User ID if authenticated, else IP address
		clientIdentifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			if idStr, ok := userID.(string); ok && idStr != "" {
				clientIdentifier = "user:" + idStr
			}
		}

		key := fmt.Sprintf("ratelimit:%s:%s", c.FullPath(), clientIdentifier)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		pipe := client.Pipeline()
		incrCmd := pipe.Incr(ctx, key)
		pipe.ExpireNX(ctx, key, window)
		ttlCmd := pipe.TTL(ctx, key)

		if _, err := pipe.Exec(ctx); err != nil {
			// Fail open on Redis error so legitimate traffic is not blocked
			c.Next()
			return
		}

		currentCount := int(incrCmd.Val())
		ttl := ttlCmd.Val()

		remaining := limit - currentCount
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.Itoa(int(ttl.Seconds())))

		if currentCount > limit {
			c.Header("Retry-After", strconv.Itoa(int(ttl.Seconds())))
			response.Error(
				c,
				http.StatusTooManyRequests,
				"TOO_MANY_REQUESTS",
				"Rate limit exceeded. Please wait before sending more requests.",
			)
			c.Abort()
			return
		}

		c.Next()
	}
}
