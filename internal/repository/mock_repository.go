package repository

import (
	"context"
	"database/sql"

	"wallet-service/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MockUserRepository implements IUserRepository for testing
type MockUserRepository struct {
	users map[string]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) Create(user *models.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetAll() ([]models.User, error) {
	users := make([]models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

// MockWalletRepository implements IWalletRepository for testing
type MockWalletRepository struct {
	wallets map[uuid.UUID]*models.Wallet
	db      *sql.DB
}

func NewMockWalletRepository() *MockWalletRepository {
	return &MockWalletRepository{
		wallets: make(map[uuid.UUID]*models.Wallet),
	}
}

func (m *MockWalletRepository) GetByID(id uuid.UUID) (*models.Wallet, error) {
	if wallet, exists := m.wallets[id]; exists {
		return wallet, nil
	}
	return nil, nil
}

func (m *MockWalletRepository) GetByUserID(userID uuid.UUID) ([]models.Wallet, error) {
	var wallets []models.Wallet
	for _, wallet := range m.wallets {
		if wallet.UserID == userID {
			wallets = append(wallets, *wallet)
		}
	}
	return wallets, nil
}

func (m *MockWalletRepository) Create(wallet *models.Wallet) error {
	m.wallets[wallet.ID] = wallet
	return nil
}

func (m *MockWalletRepository) UpdateAmount(ctx context.Context, tx *sql.Tx, walletID uuid.UUID, amount decimal.Decimal) error {
	if wallet, exists := m.wallets[walletID]; exists {
		wallet.Amount = amount
	}
	return nil
}

func (m *MockWalletRepository) UpdateFrozenAmount(ctx context.Context, tx *sql.Tx, walletID uuid.UUID, frozenAmount decimal.Decimal) error {
	if wallet, exists := m.wallets[walletID]; exists {
		wallet.FrozenAmount = frozenAmount
	}
	return nil
}

func (m *MockWalletRepository) GetByUserIDAndCoinType(userID uuid.UUID, coinType models.CoinType) (*models.Wallet, error) {
	for _, wallet := range m.wallets {
		if wallet.UserID == userID && wallet.CoinType == coinType {
			return wallet, nil
		}
	}
	return nil, nil
}

func (m *MockWalletRepository) GetDB() *sql.DB {
	return m.db
}

// MockTransactionRepository implements ITransactionRepository for testing
type MockTransactionRepository struct {
	transactions map[uuid.UUID]*models.Transaction
	entries      map[uuid.UUID]*models.TransactionEntry
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: make(map[uuid.UUID]*models.Transaction),
		entries:      make(map[uuid.UUID]*models.TransactionEntry),
	}
}

func (m *MockTransactionRepository) CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error {
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MockTransactionRepository) CreateTransactionEntry(ctx context.Context, tx *sql.Tx, entry *models.TransactionEntry) error {
	m.entries[entry.ID] = entry
	return nil
}

func (m *MockTransactionRepository) GetTransactionHistory(walletID uuid.UUID, req *models.TransactionHistoryRequest) (*models.TransactionHistoryResponse, error) {
	var entries []models.TransactionEntry
	for _, entry := range m.entries {
		if entry.WalletID == walletID {
			entries = append(entries, *entry)
		}
	}
	return &models.TransactionHistoryResponse{
		Transactions: entries,
		Total:        len(entries),
	}, nil
}

func (m *MockTransactionRepository) GetTransactionByID(id uuid.UUID) (*models.Transaction, error) {
	if transaction, exists := m.transactions[id]; exists {
		return transaction, nil
	}
	return nil, nil
}

func (m *MockTransactionRepository) GetTransactionEntriesByTxnID(txnID uuid.UUID) ([]models.TransactionEntry, error) {
	var entries []models.TransactionEntry
	for _, entry := range m.entries {
		if entry.TxnID == txnID {
			entries = append(entries, *entry)
		}
	}
	return entries, nil
}
