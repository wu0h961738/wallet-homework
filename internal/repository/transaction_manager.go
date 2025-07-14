package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// TransactionManager manages database transactions
type TransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

type TransactionFunc func(ctx context.Context, tx *sql.Tx) error

// Execute transactions
func (tm *TransactionManager) ExecuteTransaction(ctx context.Context, fn TransactionFunc) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	var committed bool
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// core
	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	// commit
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true
	return nil
}

// GetDB 获取数据库连接（仅用于repository内部使用）
func (tm *TransactionManager) GetDB() *sql.DB {
	return tm.db
}
