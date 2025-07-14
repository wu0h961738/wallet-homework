package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func IdempotencyGuard(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			c.Abort()
			return
		}

		idempotencyKey := c.GetHeader("X-Idempotency-Key")
		if idempotencyKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "X-Idempotency-Key header required"})
			c.Abort()
			return
		}

		// Create Redis key for idempotency check
		redisKey := fmt.Sprintf("txn_lock:%s:%s", userID, idempotencyKey)
		ctx := context.Background()

		// Try to set the key with 5 seconds TTL
		// If key already exists, it means this request was already processed
		success, err := redisClient.SetNX(ctx, redisKey, "1", 5*time.Second).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check idempotency"})
			c.Abort()
			return
		}

		if !success {
			c.JSON(http.StatusConflict, gin.H{"error": "Duplicate request detected"})
			c.Abort()
			return
		}

		// Store the key in context for potential cleanup
		c.Set("idempotency_key", redisKey)
		c.Next()
	}
}
