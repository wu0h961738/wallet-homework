package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"wallet-service/internal/models"
	"wallet-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletHandler struct {
	walletRepo      repository.IWalletRepository
	transactionRepo repository.ITransactionRepository
	txManager       *repository.TransactionManager
}

func NewWalletHandler(walletRepo repository.IWalletRepository, transactionRepo repository.ITransactionRepository, txManager *repository.TransactionManager) *WalletHandler {
	return &WalletHandler{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		txManager:       txManager,
	}
}

func (h *WalletHandler) Deposit(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("wallet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet ID"})
		return
	}

	var req models.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get wallet and verify ownership
	wallet, err := h.walletRepo.GetByID(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wallet"})
		return
	}

	if wallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	if wallet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Perform deposit transaction
	err = h.performDeposit(c.Request.Context(), wallet, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process deposit: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful", "amount": req.Amount})
}

func (h *WalletHandler) Withdraw(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("wallet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet ID"})
		return
	}

	var req models.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get wallet and verify ownership
	wallet, err := h.walletRepo.GetByID(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wallet"})
		return
	}

	if wallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	if wallet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check sufficient balance
	if wallet.Amount.LessThan(req.Amount) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Perform withdrawal transaction
	err = h.performWithdrawal(c.Request.Context(), wallet, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process withdrawal: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful", "amount": req.Amount})
}

func (h *WalletHandler) Transfer(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("wallet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet ID"})
		return
	}

	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get sender wallet and verify ownership
	senderWallet, err := h.walletRepo.GetByID(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sender wallet"})
		return
	}

	if senderWallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender wallet not found"})
		return
	}

	if senderWallet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get receiver wallet
	receiverWallet, err := h.walletRepo.GetByID(req.ReceiverWalletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get receiver wallet"})
		return
	}

	if receiverWallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receiver wallet not found"})
		return
	}

	// Check sufficient balance
	if senderWallet.Amount.LessThan(req.Amount) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Perform transfer transaction
	err = h.performTransfer(c.Request.Context(), senderWallet, receiverWallet, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process transfer: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful", "amount": req.Amount})
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("wallet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get wallet and verify ownership
	wallet, err := h.walletRepo.GetByID(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wallet"})
		return
	}

	if wallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	if wallet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	response := models.BalanceResponse{
		WalletID:     wallet.ID,
		CoinType:     wallet.CoinType,
		Amount:       wallet.Amount,
		FrozenAmount: wallet.FrozenAmount,
	}

	c.JSON(http.StatusOK, response)
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("wallet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get wallet and verify ownership
	wallet, err := h.walletRepo.GetByID(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wallet"})
		return
	}

	if wallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	if wallet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse query parameters
	var req models.TransactionHistoryRequest

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = &limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = &offset
		}
	}

	// Get transaction history
	response, err := h.transactionRepo.GetTransactionHistory(walletID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction history"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *WalletHandler) GetUserWallets(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get all wallets for the user
	wallets, err := h.walletRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user wallets"})
		return
	}

	response := models.UserWalletsResponse{
		UserID:  userID,
		Wallets: wallets,
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods for transaction processing
func (h *WalletHandler) performDeposit(ctx context.Context, wallet *models.Wallet, amount decimal.Decimal) error {
	return h.txManager.ExecuteTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Create transaction record
		transaction := &models.Transaction{
			ID:     uuid.New(),
			Type:   models.TransactionTypeDeposit,
			Status: models.TransactionStatusDone,
		}

		err := h.transactionRepo.CreateTransaction(ctx, tx, transaction)
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// Update wallet amount
		newAmount := wallet.Amount.Add(amount)
		err = h.walletRepo.UpdateAmount(ctx, tx, wallet.ID, newAmount)
		if err != nil {
			return fmt.Errorf("failed to update wallet amount: %w", err)
		}

		// Create transaction entry
		entry := &models.TransactionEntry{
			ID:        uuid.New(),
			TxnID:     transaction.ID,
			WalletID:  wallet.ID,
			Direction: models.DirectionIn,
			Amount:    amount,
		}

		err = h.transactionRepo.CreateTransactionEntry(ctx, tx, entry)
		if err != nil {
			return fmt.Errorf("failed to create transaction entry: %w", err)
		}

		return nil
	})
}

func (h *WalletHandler) performWithdrawal(ctx context.Context, wallet *models.Wallet, amount decimal.Decimal) error {
	return h.txManager.ExecuteTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Create transaction record
		transaction := &models.Transaction{
			ID:     uuid.New(),
			Type:   models.TransactionTypeWithdrawal,
			Status: models.TransactionStatusDone,
		}

		err := h.transactionRepo.CreateTransaction(ctx, tx, transaction)
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// Update wallet amount
		newAmount := wallet.Amount.Sub(amount)
		err = h.walletRepo.UpdateAmount(ctx, tx, wallet.ID, newAmount)
		if err != nil {
			return fmt.Errorf("failed to update wallet amount: %w", err)
		}

		// Create transaction entry
		entry := &models.TransactionEntry{
			ID:        uuid.New(),
			TxnID:     transaction.ID,
			WalletID:  wallet.ID,
			Direction: models.DirectionOut,
			Amount:    amount,
		}

		err = h.transactionRepo.CreateTransactionEntry(ctx, tx, entry)
		if err != nil {
			return fmt.Errorf("failed to create transaction entry: %w", err)
		}

		return nil
	})
}

func (h *WalletHandler) performTransfer(ctx context.Context, senderWallet, receiverWallet *models.Wallet, amount decimal.Decimal) error {
	return h.txManager.ExecuteTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Create transaction record
		transaction := &models.Transaction{
			ID:     uuid.New(),
			Type:   models.TransactionTypeTransfer,
			Status: models.TransactionStatusDone,
		}

		err := h.transactionRepo.CreateTransaction(ctx, tx, transaction)
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// Update sender wallet amount
		senderNewAmount := senderWallet.Amount.Sub(amount)
		err = h.walletRepo.UpdateAmount(ctx, tx, senderWallet.ID, senderNewAmount)
		if err != nil {
			return fmt.Errorf("failed to update sender wallet amount: %w", err)
		}

		// Update receiver wallet amount
		receiverNewAmount := receiverWallet.Amount.Add(amount)
		err = h.walletRepo.UpdateAmount(ctx, tx, receiverWallet.ID, receiverNewAmount)
		if err != nil {
			return fmt.Errorf("failed to update receiver wallet amount: %w", err)
		}

		// Create transaction entries
		senderEntry := &models.TransactionEntry{
			ID:                   uuid.New(),
			TxnID:                transaction.ID,
			WalletID:             senderWallet.ID,
			Direction:            models.DirectionOut,
			Amount:               amount,
			CounterpartyWalletID: &receiverWallet.ID,
		}

		receiverEntry := &models.TransactionEntry{
			ID:                   uuid.New(),
			TxnID:                transaction.ID,
			WalletID:             receiverWallet.ID,
			Direction:            models.DirectionIn,
			Amount:               amount,
			CounterpartyWalletID: &senderWallet.ID,
		}

		err = h.transactionRepo.CreateTransactionEntry(ctx, tx, senderEntry)
		if err != nil {
			return fmt.Errorf("failed to create sender transaction entry: %w", err)
		}

		err = h.transactionRepo.CreateTransactionEntry(ctx, tx, receiverEntry)
		if err != nil {
			return fmt.Errorf("failed to create receiver transaction entry: %w", err)
		}

		return nil
	})
}
