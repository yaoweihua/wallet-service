// Package repository provides the data access layer for interacting with the
// database. It includes functions for performing CRUD operations on the
// wallet service's core entities such as users, transactions, and balances.
package repository

import (
    "context"
    "fmt"
    "github.com/jmoiron/sqlx"
    "github.com/yaoweihua/wallet-service/model"
    "github.com/shopspring/decimal"
    "github.com/yaoweihua/wallet-service/utils"
    "github.com/sirupsen/logrus"
)

// TransactionRepository provides database operations related to transactions
type TransactionRepository struct {
    DB     *sqlx.DB
    Logger *logrus.Logger
}

// NewTransactionRepository creates a new instance of TransactionRepository
func NewTransactionRepository(DB *sqlx.DB) *TransactionRepository {
    logger := utils.GetLogger()
    return &TransactionRepository{
        DB:     DB,
        Logger: logger,
    }
}

// RecordTransaction records a new transaction in the database.
// It stores the details of the transaction including the sender, receiver, amount, type, and status.
func (r *TransactionRepository) RecordTransaction(ctx context.Context, tx *sqlx.Tx, fromUserID, toUserID int, amount decimal.Decimal, txType string, txStatus string) error {
    transactionFee := decimal.NewFromFloat(0.0)  // Assume that the transaction handling fee is a fixed value, and it can also be modified according to the actual business
    paymentMethod := "credit_card"  // Assume that the payment method is a fixed value and can also be modified

    query := `
        INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, transaction_status, transaction_fee, payment_method, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
    `

    // 执行插入操作
    _, err := tx.ExecContext(ctx, query, fromUserID, toUserID, amount, txType, txStatus, transactionFee, paymentMethod)
    if err != nil {
        r.Logger.Error(fmt.Sprintf("Failed to record transaction from user %d to user %d, amount: %s, type: %s", fromUserID, toUserID, amount.String(), txType), err)
        return fmt.Errorf("failed to record transaction for user %d to user %d: %w", fromUserID, toUserID, err)
    }

    //r.Logger.Info(fmt.Sprintf("Recorded transaction from user %d to user %d, amount: %s, type: %s", fromUserID, toUserID, amount.String(), txType))
    return nil
}

// GetTransactions retrieves the transaction records of the specified user
func (r *TransactionRepository) GetTransactions(ctx context.Context, userID int) ([]model.Transaction, error) {
    var transactions []model.Transaction

    query := `
        SELECT 
            id, 
            from_user_id, 
            to_user_id, 
            amount, 
            transaction_type, 
            transaction_status, 
            created_at, 
            updated_at
        FROM transactions
        WHERE from_user_id = $1 OR to_user_id = $1
        ORDER BY created_at DESC
    `
    
    err := r.DB.SelectContext(ctx, &transactions, query, userID)
    if err != nil {
        r.Logger.Error(fmt.Sprintf("Error getting transactions for user %d", userID), err)
        return nil, err
    }

    return transactions, nil
}