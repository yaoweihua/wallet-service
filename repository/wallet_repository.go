package repository

import (
    "context"
    "fmt"
    "database/sql"
    "github.com/jmoiron/sqlx"
    "github.com/yaoweihua/wallet-service/model"
    "github.com/shopspring/decimal"
    "github.com/yaoweihua/wallet-service/utils"
    "time"
    "github.com/sirupsen/logrus"
)

// WalletRepository provides database operations related to wallets
type WalletRepository struct {
    DB     *sqlx.DB
    Logger *logrus.Logger
}

// NewWalletRepository creates a new instance of WalletRepository
func NewWalletRepository(db *sqlx.DB) *WalletRepository {
    logger := utils.GetLogger()
    return &WalletRepository{
        DB:     db,
        Logger: logger,
    }
}

// GetUserBalance retrieves the current balance of the user
func (r *WalletRepository) GetUserBalance(ctx context.Context, userID int) (*model.User, error) {
    var user model.User
    query := `
        SELECT id, balance
        FROM users
        WHERE id = $1
        FOR UPDATE
    `
    // Execute the query and apply a lock
    err := r.DB.GetContext(ctx, &user, query, userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user %d not found", userID)
        }
        r.Logger.Error(fmt.Sprintf("Failed to fetch user balance for user %d", userID), err)
        return nil, fmt.Errorf("failed to fetch user balance: %w", err)
    }

    return &user, nil
}

// UpdateBalance updates the user's balance
func (r *WalletRepository) UpdateBalance(ctx context.Context, tx *sqlx.Tx, userID int, newBalance decimal.Decimal) error {
    now := time.Now()
    _, err := tx.ExecContext(ctx, "UPDATE users SET balance = $1, updated_at = $2 WHERE id = $3", newBalance, now, userID)
    if err != nil {
        r.Logger.Error(fmt.Sprintf("Failed to update balance for user %d", userID), err)
        return fmt.Errorf("failed to update balance for user %d: %w", userID, err)
    }
    return nil
}
