package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CoinType string
type TransactionType string
type TransactionStatus string
type Direction string

const (
	CoinTypeBTC CoinType = "BTC"
	CoinTypeETH CoinType = "ETH"
	CoinTypeADA CoinType = "ADA"

	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypeTransfer   TransactionType = "TRANSFER"

	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusDone    TransactionStatus = "DONE"
	TransactionStatusFailed  TransactionStatus = "FAILED"

	DirectionIn  Direction = "IN"
	DirectionOut Direction = "OUT"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Wallet struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	UserID       uuid.UUID       `json:"user_id" db:"user_id"`
	CoinType     CoinType        `json:"coin_type" db:"coin_type"`
	Amount       decimal.Decimal `json:"amount" db:"amount"`
	FrozenAmount decimal.Decimal `json:"frozen_amount" db:"frozen_amount"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

type Transaction struct {
	ID        uuid.UUID         `json:"id" db:"id"`
	Type      TransactionType   `json:"type" db:"type"`
	Status    TransactionStatus `json:"status" db:"status"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
}

type TransactionEntry struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	TxnID                uuid.UUID       `json:"txn_id" db:"txn_id"`
	WalletID             uuid.UUID       `json:"wallet_id" db:"wallet_id"`
	Direction            Direction       `json:"direction" db:"direction"`
	Amount               decimal.Decimal `json:"amount" db:"amount"`
	CounterpartyWalletID *uuid.UUID      `json:"counterparty_wallet_id,omitempty" db:"counterparty_wallet_id"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
}

// Request/Response models
type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type DepositRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required,gt=0"`
}

type WithdrawRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required,gt=0"`
}

type TransferRequest struct {
	ReceiverWalletID uuid.UUID       `json:"receiver_wallet_id" binding:"required"`
	Amount           decimal.Decimal `json:"amount" binding:"required,gt=0"`
}

type BalanceResponse struct {
	WalletID     uuid.UUID       `json:"wallet_id"`
	CoinType     CoinType        `json:"coin_type"`
	Amount       decimal.Decimal `json:"amount"`
	FrozenAmount decimal.Decimal `json:"frozen_amount"`
}

type TransactionHistoryRequest struct {
	StartDate            *time.Time `form:"start_date"`
	EndDate              *time.Time `form:"end_date"`
	CounterpartyWalletID *uuid.UUID `form:"counterparty_wallet_id"`
	Limit                *int       `form:"limit"`
	Offset               *int       `form:"offset"`
}

type TransactionHistoryResponse struct {
	Transactions []TransactionEntry `json:"transactions"`
	Total        int                `json:"total"`
}

type UserWalletsResponse struct {
	UserID  uuid.UUID `json:"user_id"`
	Wallets []Wallet  `json:"wallets"`
}
