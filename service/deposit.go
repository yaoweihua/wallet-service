// Package service provides the core business logic for handling wallet and transaction operations.
// It includes services for deposit, withdrawal, and transaction management.
// Each service interacts with the underlying repositories for data persistence and business rules enforcement.
// The services also manage Redis caching for balance data to optimize performance and reduce database load.
//
// The WithdrawService is responsible for handling user withdrawals, ensuring that the user has sufficient balance,
// recording the transaction, updating the balance, and caching the new balance in Redis.
// The service ensures thread-safety during withdrawal operations by using locks to prevent race conditions.
package service

import (
    "context"
    "fmt"
    "sync"
    "github.com/yaoweihua/wallet-service/repository"
    "github.com/yaoweihua/wallet-service/utils"
    "github.com/shopspring/decimal"
    "github.com/go-redis/redis/v8"
    "github.com/jmoiron/sqlx"
    "time"
)

// Use global locks to prevent concurrent access to the same transaction
var transferLocks sync.Map

// Use global locks to prevent concurrent access to the balance of the same user
var userLocks sync.Map

// DepositService provides methods for handling deposit operations.
// It interacts with the WalletRepository and TransactionRepository to manage user balances and transaction records.
type DepositService struct {
    walletRepo      *repository.WalletRepository
    transactionRepo *repository.TransactionRepository
    dbConn          *sqlx.DB
    redisClient     *redis.Client
}

// NewDepositService creates a new instance of DepositService.
// It initializes the service with the provided database connection and Redis client.
func NewDepositService(dbConn *sqlx.DB, redisClient *redis.Client) *DepositService {
    walletRepo := repository.NewWalletRepository(dbConn)
    transactionRepo := repository.NewTransactionRepository(dbConn)

    return &DepositService{
        walletRepo:      walletRepo,
        transactionRepo: transactionRepo,
        dbConn:          dbConn,
        redisClient:     redisClient,
    }
}

// Deposit function is responsible for handling the deposit logic
func (s *DepositService) Deposit(userID int, amount decimal.Decimal) error {
    // Acquire the user lock to prevent concurrent conflicts
    logger := utils.GetLogger()
    lock, ok := userLocks.LoadOrStore(userID, &sync.Mutex{})
    if !ok {
        logger.Warnf("Lock for user %d was newly created", userID)
    }

    userLock, ok := lock.(*sync.Mutex)
    if !ok {
        return fmt.Errorf("failed to assert lock as *sync.Mutex for user %d", userID)
    }

    userLock.Lock()
    defer userLock.Unlock()

    // Check whether the deposit amount is reasonable
    if amount.LessThanOrEqual(decimal.Zero) {
        return fmt.Errorf("Deposit amount must be greater than zero")
    }

    // Begin the database transaction
    conn := s.dbConn
    ctx := context.Background()
    tx, err := conn.Beginx()
    if err != nil {
        return fmt.Errorf("failed to start transaction: %w", err)
    }

    defer func() {
        if rErr := tx.Rollback(); rErr != nil && err == nil {
            // Only log rollback error if the original error is nil (i.e., the function hasn't failed yet)
            logger.Warnf("rollback transaction: %v", rErr)
        }
    }()

    // Query the current balance of the user, and use row-level locks to ensure the consistency of the balance
    user, err := s.walletRepo.GetUserBalance(ctx, userID)
    if err != nil {
        return err
    }

    // Update the user's balance
    newBalance := user.Balance.Add(amount)
    if err := s.walletRepo.UpdateBalance(ctx, tx, userID, newBalance); err != nil {
        return err
    }

    // Record the deposit transaction
    if err := s.transactionRepo.RecordTransaction(ctx, tx, userID, 0, amount, "deposit", "completed"); err != nil {
        return err
    }

    // Commit the transaction
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    // Update the Redis cache
    cacheKey := fmt.Sprintf("balance:%d", userID)
    if s.redisClient != nil {
        if err := s.redisClient.Set(ctx, cacheKey, newBalance.String(), time.Second*3600); err != nil {
            logger.Warnf("Warning: failed to cache balance: %v", err)
        }        
    }
    return nil
}
