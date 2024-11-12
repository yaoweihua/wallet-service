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

type TransferService struct {
    walletRepo      *repository.WalletRepository
    transactionRepo *repository.TransactionRepository
    dbConn          *sqlx.DB
    redisClient     *redis.Client
}

func NewTransferService(dbConn *sqlx.DB, redisClient *redis.Client) *TransferService {
    walletRepo := repository.NewWalletRepository(dbConn)
    transactionRepo := repository.NewTransactionRepository(dbConn)

    return &TransferService{
        walletRepo:      walletRepo,
        transactionRepo: transactionRepo,
        dbConn:          dbConn,
        redisClient:     redisClient,
    }
}

// Transfer handles the transfer logic
func (s *TransferService) Transfer(fromUserID, toUserID int, amount decimal.Decimal) error {
    if fromUserID == toUserID {
        return fmt.Errorf("cannot transfer to the same user")
    }

    // Acquire the locks for the transferring-out and receiving users. Lock them in the order of user IDs to avoid deadlocks
    lockFrom, _ := transferLocks.LoadOrStore(fromUserID, &sync.Mutex{})
    lockTo, _ := transferLocks.LoadOrStore(toUserID, &sync.Mutex{})

    // Lock in the order of user IDs to avoid deadlocks
    if fromUserID < toUserID {
        lockFrom.(*sync.Mutex).Lock()
        lockTo.(*sync.Mutex).Lock()
    } else {
        lockTo.(*sync.Mutex).Lock()
        lockFrom.(*sync.Mutex).Lock()
    }
    defer lockFrom.(*sync.Mutex).Unlock()
    defer lockTo.(*sync.Mutex).Unlock()

    // Check whether the deposit amount is reasonable
    if amount.LessThanOrEqual(decimal.Zero) {
        return fmt.Errorf("transfer amount must be greater than zero")
    }

    // Begin the database transaction
    conn := s.dbConn
    ctx := context.Background()
    tx, err := conn.Beginx()
    if err != nil {
        return fmt.Errorf("failed to start transaction: %w", err)
    }
    defer tx.Rollback()

    // Retrieve the balance of the transferring-out user
    fromUser, err := s.walletRepo.GetUserBalance(ctx, fromUserID)
    if err != nil {
        s.transactionRepo.RecordTransaction(ctx, tx, fromUserID, toUserID, amount, "transfer", "failed")
        return fmt.Errorf("failed to get balance for user %d: %w", fromUserID, err)
    }

    // Check whether the balance is sufficient
    if fromUser.Balance.LessThan(amount) {
        s.transactionRepo.RecordTransaction(ctx, tx, fromUserID, toUserID, amount, "transfer", "failed")
        return fmt.Errorf("insufficient balance for transfer")
    }

    // Deduct the balance of the transferring-out user
    newFromBalance := fromUser.Balance.Sub(amount)
    if err := s.walletRepo.UpdateBalance(ctx, tx, fromUserID, newFromBalance); err != nil {
        s.transactionRepo.RecordTransaction(ctx, tx, fromUserID, toUserID, amount, "transfer", "failed")
        return fmt.Errorf("failed to update balance for user %d: %w", fromUserID, err)
    }

    // Retrieve the balance of the receiving user
    toUser, err := s.walletRepo.GetUserBalance(ctx, toUserID)
    if err != nil {
        s.transactionRepo.RecordTransaction(ctx, tx, fromUserID, toUserID, amount, "transfer", "failed")
        return fmt.Errorf("failed to get balance for user %d: %w", toUserID, err)
    }

    // Increase the balance of the receiving user
    newToBalance := toUser.Balance.Add(amount)
    if err := s.walletRepo.UpdateBalance(ctx, tx, toUserID, newToBalance); err != nil {
        s.transactionRepo.RecordTransaction(ctx, tx, fromUserID, toUserID, amount, "transfer", "failed")
        return fmt.Errorf("failed to update balance for user %d: %w", toUserID, err)
    }

    // Record the transaction between the transferring-out user and the receiving user
    err = s.transactionRepo.RecordTransaction(ctx, tx, fromUserID, toUserID, amount, "transfer", "completed")
    if err != nil {
        s.transactionRepo.RecordTransaction(ctx, tx, fromUserID, toUserID, amount, "transfer", "failed")
        return fmt.Errorf("failed to record transaction for user %d: %w", fromUserID, err)
    }

    // Commit the transaction
    if err := tx.Commit(); err != nil {
        s.transactionRepo.RecordTransaction(ctx, tx, fromUserID, toUserID, amount, "transfer", "failed")
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    // Update the Redis cache
    cacheFrom := fmt.Sprintf("balance:%d", fromUserID)
    cacheTo := fmt.Sprintf("balance:%d", toUserID)
    if err := s.redisClient.Set(ctx, cacheFrom, newFromBalance.String(), time.Second*3600); err != nil {
        log.Printf("Warning: failed to cache balance for user %d: %v", fromUserID, err)
    }
    if err := s.redisClient.Set(ctx, cacheTo, newToBalance.String(), time.Second*3600); err != nil {
        log.Printf("Warning: failed to cache balance for user %d: %v", toUserID, err)
    }

    return nil
}
