package handlers

import (
	"testing"

	"wallet-service/internal/models"
	"wallet-service/internal/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestAuthHandler_Login(t *testing.T) {
	// Create mock repositories
	mockUserRepo := repository.NewMockUserRepository()

	// Create a test user
	testUser := &models.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	}
	mockUserRepo.Create(testUser)

	// Create mock Redis client (you might want to create a mock for this too)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Create auth handler with mock dependencies
	authHandler := NewAuthHandler(mockUserRepo, redisClient, "test-secret")

	// Test that the handler was created successfully
	if authHandler == nil {
		t.Error("Expected auth handler to be created, got nil")
	}

	// Test that the handler uses the interface correctly
	if authHandler.userRepo == nil {
		t.Error("Expected userRepo to be set, got nil")
	}
}

// This test demonstrates how easy it is to test with interfaces
func TestAuthHandler_WithMockDependencies(t *testing.T) {
	// Create mock repositories
	mockUserRepo := repository.NewMockUserRepository()

	// Create a test user
	testUser := &models.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	}
	mockUserRepo.Create(testUser)

	// Verify the mock works
	retrievedUser, err := mockUserRepo.GetByEmail("test@example.com")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrievedUser == nil {
		t.Error("Expected user to be retrieved, got nil")
	}
	if retrievedUser.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", retrievedUser.Email)
	}

	// Test with non-existent user
	nonExistentUser, err := mockUserRepo.GetByEmail("nonexistent@example.com")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if nonExistentUser != nil {
		t.Error("Expected nil for non-existent user, got user")
	}
}
