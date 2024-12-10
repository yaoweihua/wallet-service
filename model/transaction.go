// Package model contains data structures representing core domain entities,
// such as transactions and users, used throughout the wallet service by William, way1910@gmail.com..
package model

import (
    "time"

    "github.com/shopspring/decimal"
)

// Transaction represents a financial transaction between users,
// including details such as the transaction ID, type, amount, status,
// and payment method. It supports deposit, withdrawal, and transfer types.
type Transaction struct {
    ID               int             `json:"id" db:"id"`                                // Transaction ID
    FromUserID       int             `json:"from_user_id" db:"from_user_id"`            // The user ID of the transaction initiator
    ToUserID         int             `json:"to_user_id,omitempty" db:"to_user_id"`      // The user ID of the transaction recipient, 0 for deposits and withdrawals, used only for transfers
    Amount           decimal.Decimal `json:"amount" db:"amount"`                        // The transaction amount
    TransactionType  string          `json:"transaction_type" db:"transaction_type"`    // The transaction type, such as deposit, withdraw, transfer
    TransactionStatus string         `json:"transaction_status" db:"transaction_status"` // The transaction status, such as completed, failed
    TransactionFee   decimal.Decimal `json:"transaction_fee,omitempty" db:"transaction_fee"` // The transaction fee, currently set to 0.0, with potential for future expansion
    PaymentMethod    string          `json:"payment_method" db:"payment_method"`        // The payment method, such as credit_card, bank_transfer, paypal, etc.
    CreatedAt        time.Time       `json:"created_at" db:"created_at"`                // Creation time
    UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`                // Update time
}
