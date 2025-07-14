package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"wallet-service/internal/models"
	"wallet-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type AuthHandler struct {
	userRepo    repository.IUserRepository
	redisClient *redis.Client
	jwtSecret   string
}

func NewAuthHandler(userRepo repository.IUserRepository, redisClient *redis.Client, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		redisClient: redisClient,
		jwtSecret:   jwtSecret,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user by email
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := h.generateToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Cache token in Redis
	ctx := context.Background()
	redisKey := fmt.Sprintf("jwt:%s", token)
	err = h.redisClient.Set(ctx, redisKey, "1", 24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cache token"})
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) generateToken(userID string) (string, error) {
	// Create token with 100 years expiration for test user
	expirationTime := time.Now().Add(100 * 365 * 24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
