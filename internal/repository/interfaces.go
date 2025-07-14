package repository

import (
	"context"
	"database/sql"

	"wallet-service/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// IUserRepository defines the interface for user data operations
type IUserRepository interface {
	GetByEmail(email string) (*models.User, error)
	GetByID(id uuid.UUID) (*models.User, error)
	Create(user *models.User) error
	GetAll() ([]models.User, error)
}

// IWalletRepository defines the interface for wallet data operations
type IWalletRepository interface {
	GetByID(id uuid.UUID) (*models.Wallet, error)
	GetByUserID(userID uuid.UUID) ([]models.Wallet, error)
	Create(wallet *models.Wallet) error
	GetByUserIDAndCoinType(userID uuid.UUID, coinType models.CoinType) (*models.Wallet, error)

	// Transaction methods - 接受事务上下文
	UpdateAmount(ctx context.Context, tx *sql.Tx, walletID uuid.UUID, amount decimal.Decimal) error
	UpdateFrozenAmount(ctx context.Context, tx *sql.Tx, walletID uuid.UUID, frozenAmount decimal.Decimal) error
}

// ITransactionRepository defines the interface for transaction data operations
type ITransactionRepository interface {
	GetTransactionHistory(walletID uuid.UUID, req *models.TransactionHistoryRequest) (*models.TransactionHistoryResponse, error)
	GetTransactionByID(id uuid.UUID) (*models.Transaction, error)
	GetTransactionEntriesByTxnID(txnID uuid.UUID) ([]models.TransactionEntry, error)

	// Transaction methods - 接受事务上下文
	CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error
	CreateTransactionEntry(ctx context.Context, tx *sql.Tx, entry *models.TransactionEntry) error
}
