package repository

import (
	"context"
	"database/sql"
	"fmt"

	"wallet-service/internal/models"

	"github.com/google/uuid"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error {
	query := `INSERT INTO transactions (id, type, status) VALUES ($1, $2, $3) RETURNING created_at`

	return tx.QueryRowContext(ctx, query, transaction.ID, transaction.Type, transaction.Status).Scan(&transaction.CreatedAt)
}

func (r *TransactionRepository) CreateTransactionEntry(ctx context.Context, tx *sql.Tx, entry *models.TransactionEntry) error {
	query := `INSERT INTO transaction_entries (id, txn_id, wallet_id, direction, amount, counterparty_wallet_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`

	return tx.QueryRowContext(ctx, query, entry.ID, entry.TxnID, entry.WalletID, entry.Direction, entry.Amount, entry.CounterpartyWalletID).Scan(&entry.CreatedAt)
}

func (r *TransactionRepository) GetTransactionHistory(walletID uuid.UUID, req *models.TransactionHistoryRequest) (*models.TransactionHistoryResponse, error) {
	baseQuery := `SELECT id, txn_id, wallet_id, direction, amount, counterparty_wallet_id, created_at FROM transaction_entries WHERE wallet_id = $1`
	countQuery := `SELECT COUNT(*) FROM transaction_entries WHERE wallet_id = $1`

	args := []interface{}{walletID}
	argIndex := 2

	// Add filters
	if req.StartDate != nil {
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *req.StartDate)
		argIndex++
	}

	if req.EndDate != nil {
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *req.EndDate)
		argIndex++
	}

	if req.CounterpartyWalletID != nil {
		baseQuery += fmt.Sprintf(" AND counterparty_wallet_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND counterparty_wallet_id = $%d", argIndex)
		args = append(args, *req.CounterpartyWalletID)
		argIndex++
	}

	// Add ordering
	baseQuery += " ORDER BY created_at DESC"

	// Add pagination
	if req.Limit != nil {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, *req.Limit)
		argIndex++

		if req.Offset != nil {
			baseQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, *req.Offset)
		}
	}

	// Get total count
	var total int
	err := r.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction count: %w", err)
	}

	// Get transactions
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}
	defer rows.Close()

	var transactions []models.TransactionEntry
	for rows.Next() {
		var entry models.TransactionEntry
		err := rows.Scan(
			&entry.ID,
			&entry.TxnID,
			&entry.WalletID,
			&entry.Direction,
			&entry.Amount,
			&entry.CounterpartyWalletID,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction entry: %w", err)
		}
		transactions = append(transactions, entry)
	}

	return &models.TransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
	}, nil
}

func (r *TransactionRepository) GetTransactionByID(id uuid.UUID) (*models.Transaction, error) {
	query := `SELECT id, type, status, created_at FROM transactions WHERE id = $1`

	var transaction models.Transaction
	err := r.db.QueryRow(query, id).Scan(
		&transaction.ID,
		&transaction.Type,
		&transaction.Status,
		&transaction.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}

	return &transaction, nil
}

func (r *TransactionRepository) GetTransactionEntriesByTxnID(txnID uuid.UUID) ([]models.TransactionEntry, error) {
	query := `SELECT id, txn_id, wallet_id, direction, amount, counterparty_wallet_id, created_at FROM transaction_entries WHERE txn_id = $1`

	rows, err := r.db.Query(query, txnID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction entries by txn ID: %w", err)
	}
	defer rows.Close()

	var entries []models.TransactionEntry
	for rows.Next() {
		var entry models.TransactionEntry
		err := rows.Scan(
			&entry.ID,
			&entry.TxnID,
			&entry.WalletID,
			&entry.Direction,
			&entry.Amount,
			&entry.CounterpartyWalletID,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
