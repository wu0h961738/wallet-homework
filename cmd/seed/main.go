package main

import (
	"fmt"
	"log"
	"math/rand"

	"wallet-service/internal/config"
	"wallet-service/internal/models"
	"wallet-service/internal/persistence"
	"wallet-service/internal/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func main() {
	cfg := config.Load()
	db, err := persistence.NewPQConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize
	var userRepo repository.IUserRepository = repository.NewUserRepository(db)
	var walletRepo repository.IWalletRepository = repository.NewWalletRepository(db)

	// Seed users and wallets
	err = seedData(userRepo, walletRepo)
	if err != nil {
		log.Fatal("Failed to seed data:", err)
	}

	log.Println("Database seeded successfully!")
}

func seedData(userRepo repository.IUserRepository, walletRepo repository.IWalletRepository) error {
	// Create 20 users with wallets
	for i := 1; i <= 20; i++ {
		// Create user
		user := &models.User{
			ID:    uuid.New(),
			Name:  fmt.Sprintf("user_%03d", i),
			Email: fmt.Sprintf("user_%03d@example.com", i),
		}

		err := userRepo.Create(user)
		if err != nil {
			return fmt.Errorf("failed to create user %d: %w", i, err)
		}

		// Create wallets for each coin type
		coinTypes := []models.CoinType{models.CoinTypeBTC, models.CoinTypeETH, models.CoinTypeADA}
		for _, coinType := range coinTypes {
			// Generate random initial balance between 10 and 1000
			initialBalance := decimal.NewFromFloat(rand.Float64()*990 + 10)

			wallet := &models.Wallet{
				ID:           uuid.New(),
				UserID:       user.ID,
				CoinType:     coinType,
				Amount:       initialBalance,
				FrozenAmount: decimal.Zero,
			}

			err := walletRepo.Create(wallet)
			if err != nil {
				return fmt.Errorf("failed to create wallet for user %d, coin %s: %w", i, coinType, err)
			}
		}

		log.Printf("Created user %s with 3 wallets", user.Name)
	}

	return nil
}
