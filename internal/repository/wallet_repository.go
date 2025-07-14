package repository

import (
	"context"
	"database/sql"
	"fmt"

	"wallet-service/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetByID(id uuid.UUID) (*models.Wallet, error) {
	query := `SELECT id, user_id, coin_type, amount, frozen_amount, created_at FROM wallets WHERE id = $1`

	var wallet models.Wallet
	err := r.db.QueryRow(query, id).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.CoinType,
		&wallet.Amount,
		&wallet.FrozenAmount,
		&wallet.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get wallet by ID: %w", err)
	}

	return &wallet, nil
}

func (r *WalletRepository) GetByUserID(userID uuid.UUID) ([]models.Wallet, error) {
	query := `SELECT id, user_id, coin_type, amount, frozen_amount, created_at FROM wallets WHERE user_id = $1`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets by user ID: %w", err)
	}
	defer rows.Close()

	var wallets []models.Wallet
	for rows.Next() {
		var wallet models.Wallet
		err := rows.Scan(
			&wallet.ID,
			&wallet.UserID,
			&wallet.CoinType,
			&wallet.Amount,
			&wallet.FrozenAmount,
			&wallet.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet: %w", err)
		}
		wallets = append(wallets, wallet)
	}

	return wallets, nil
}

func (r *WalletRepository) Create(wallet *models.Wallet) error {
	query := `INSERT INTO wallets (id, user_id, coin_type, amount, frozen_amount) VALUES ($1, $2, $3, $4, $5) RETURNING created_at`

	return r.db.QueryRow(query, wallet.ID, wallet.UserID, wallet.CoinType, wallet.Amount, wallet.FrozenAmount).Scan(&wallet.CreatedAt)
}

func (r *WalletRepository) UpdateAmount(ctx context.Context, tx *sql.Tx, walletID uuid.UUID, amount decimal.Decimal) error {
	query := `UPDATE wallets SET amount = $1 WHERE id = $2`

	_, err := tx.ExecContext(ctx, query, amount, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet amount: %w", err)
	}

	return nil
}

func (r *WalletRepository) UpdateFrozenAmount(ctx context.Context, tx *sql.Tx, walletID uuid.UUID, frozenAmount decimal.Decimal) error {
	query := `UPDATE wallets SET frozen_amount = $1 WHERE id = $2`

	_, err := tx.ExecContext(ctx, query, frozenAmount, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet frozen amount: %w", err)
	}

	return nil
}

func (r *WalletRepository) GetByUserIDAndCoinType(userID uuid.UUID, coinType models.CoinType) (*models.Wallet, error) {
	query := `SELECT id, user_id, coin_type, amount, frozen_amount, created_at FROM wallets WHERE user_id = $1 AND coin_type = $2`

	var wallet models.Wallet
	err := r.db.QueryRow(query, userID, coinType).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.CoinType,
		&wallet.Amount,
		&wallet.FrozenAmount,
		&wallet.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get wallet by user ID and coin type: %w", err)
	}

	return &wallet, nil
}
