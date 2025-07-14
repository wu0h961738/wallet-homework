package main

import (
	"log"
	"os"

	"wallet-service/internal/cache"
	"wallet-service/internal/config"
	"wallet-service/internal/handlers"
	"wallet-service/internal/middleware"
	"wallet-service/internal/persistence"
	"wallet-service/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Redis instance
	redisClient := cache.NewRedisClient(cfg.RedisHost, cfg.RedisPort)
	defer redisClient.Close()

	// Initialize database instance
	db, err := persistence.NewPQConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize transaction manager instance
	txManager := repository.NewTransactionManager(db)

	var userRepo repository.IUserRepository = repository.NewUserRepository(db)
	var walletRepo repository.IWalletRepository = repository.NewWalletRepository(db)
	var transactionRepo repository.ITransactionRepository = repository.NewTransactionRepository(db)

	authHandler := handlers.NewAuthHandler(userRepo, redisClient, cfg.JWTSecret)
	walletHandler := handlers.NewWalletHandler(walletRepo, transactionRepo, txManager)

	router := gin.Default()
	router.Use(middleware.Logger())

	// Public routes
	router.POST("/auth/login", authHandler.Login)

	walletRouter := router.Group("/")
	walletRouter.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	{
		// Wallet routes
		walletRouter.GET("/wallets", walletHandler.GetUserWallets)
		walletRouter.POST("/wallets/:wallet_id/deposit", middleware.IdempotencyGuard(redisClient), walletHandler.Deposit)
		walletRouter.POST("/wallets/:wallet_id/withdraw", middleware.IdempotencyGuard(redisClient), walletHandler.Withdraw)
		walletRouter.POST("/wallets/:wallet_id/transfer", middleware.IdempotencyGuard(redisClient), walletHandler.Transfer)
		walletRouter.GET("/wallets/:wallet_id/balance", walletHandler.GetBalance)
		walletRouter.GET("/wallets/:wallet_id/transactions", walletHandler.GetTransactions)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
