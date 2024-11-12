package service

import (
    "context"
    "fmt"
    "log"
    "sync"
    "github.com/yaoweihua/wallet-service/repository"
    "github.com/shopspring/decimal"
    "github.com/jmoiron/sqlx"
    "github.com/go-redis/redis/v8"
    "time"
)

type WithdrawService struct {
    walletRepo      *repository.WalletRepository
    transactionRepo *repository.TransactionRepository
    dbConn          *sqlx.DB
    redisClient     *redis.Client
}

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

// The Withdraw function handles the logic of withdrawing money
func (s *WithdrawService) Withdraw(userID int, amount decimal.Decimal) error {
    // Acquire the user lock to prevent concurrent conflicts
    lock, _ := userLocks.LoadOrStore(userID, &sync.Mutex{})
    userLock := lock.(*sync.Mutex)
    userLock.Lock()
    defer userLock.Unlock()

    // Check whether the deposit amount is reasonable
    if amount.LessThanOrEqual(decimal.Zero) {
        return fmt.Errorf("withdraw amount must be greater than zero")
    }

    // Begin the database transaction
    conn := s.dbConn
    ctx := context.Background()
    tx, err := conn.Beginx()
    if err != nil {
        return fmt.Errorf("failed to start transaction: %w", err)
    }
    defer tx.Rollback()

    // Query the current balance of the user, using row-level locking to ensure balance consistency
    user, err := s.walletRepo.GetUserBalance(ctx, userID)
    if err != nil {
        return err
    }

    // Ensure that the balance is sufficient
    if user.Balance.LessThan(amount) {
        return fmt.Errorf("insufficient balance")
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
        log.Printf("Warning: failed to cache balance: %v", err)
    }

    return nil
}
