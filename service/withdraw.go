package service

import (
    "context"
    "fmt"
    "sync"
    "github.com/yaoweihua/wallet-service/repository"
    "github.com/yaoweihua/wallet-service/utils"
    "github.com/shopspring/decimal"
    "github.com/jmoiron/sqlx"
    "github.com/go-redis/redis/v8"
    "time"
)

// WithdrawService provides methods for handling withdrawal operations.
// It interacts with the WalletRepository and TransactionRepository to manage user balances and transaction records.
type WithdrawService struct {
    walletRepo      *repository.WalletRepository
    transactionRepo *repository.TransactionRepository
    dbConn          *sqlx.DB
    redisClient     *redis.Client
}

// NewWithdrawService creates a new instance of WithdrawService.
// It initializes the service with the provided database connection and Redis client,
// and sets up the necessary repositories for wallet and transaction management.
func NewWithdrawService(dbConn *sqlx.DB, redisClient *redis.Client) *WithdrawService {
    walletRepo := repository.NewWalletRepository(dbConn)
    transactionRepo := repository.NewTransactionRepository(dbConn)

    return &WithdrawService{
        walletRepo:      walletRepo,
        transactionRepo: transactionRepo,
        dbConn:          dbConn,
        redisClient:     redisClient,
    }
}

// Withdraw function handles the logic of withdrawing money
func (s *WithdrawService) Withdraw(userID int, amount decimal.Decimal) error {
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
        return fmt.Errorf("Withdraw amount must be greater than zero")
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

    // Query the current balance of the user, using row-level locking to ensure balance consistency
    user, err := s.walletRepo.GetUserBalance(ctx, userID)
    if err != nil {
        return err
    }

    // Ensure that the balance is sufficient
    if user.Balance.LessThan(amount) {
        return fmt.Errorf("Insufficient balance")
    }

    // Calculate the new balance
    newBalance := user.Balance.Sub(amount)

    // Update the user's balance
    if err := s.walletRepo.UpdateBalance(ctx, tx, userID, newBalance); err != nil {
        return err
    }

    // Record the withdrawal transaction
    if err := s.transactionRepo.RecordTransaction(ctx, tx, userID, 0, amount, "withdraw", "completed"); err != nil { // ToUserID = 0
        return err
    }

    // Commit the transaction
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    // Update the Redis cache
    cacheKey := fmt.Sprintf("balance:%d", userID)
    if err := s.redisClient.Set(ctx, cacheKey, newBalance.String(), time.Second*3600); err != nil {
        logger.Warnf("Warning: failed to cache balance: %v", err)
    }

    return nil
}
