package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis       *redis.Client
	maxRequests int
	window      time.Duration
}

func NewRateLimiter(redisClient *redis.Client, maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:       redisClient,
		maxRequests: maxRequests,
		window:      window,
	}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user identifier (IP or user ID)
		identifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		key := fmt.Sprintf("rate_limit:%s", identifier)
		ctx := context.Background()

		// Get current count
		count, err := rl.redis.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
			c.Abort()
			return
		}

		if count >= rl.maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Maximum %d requests per %v", rl.maxRequests, rl.window),
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := rl.redis.Pipeline()
		pipe.Incr(ctx, key)
		if count == 0 {
			pipe.Expire(ctx, key, rl.window)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit update failed"})
			c.Abort()
			return
		}

		c.Next()
	}
}
