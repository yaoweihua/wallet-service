package service

import (
    "context"
    "fmt"
    //"github.com/yaoweihua/wallet-service/db"
    "github.com/yaoweihua/wallet-service/repository"
    //"github.com/yaoweihua/wallet-service/model"
    "github.com/go-redis/redis/v8"
    "github.com/shopspring/decimal"
    "log"
    "github.com/jmoiron/sqlx"
    "time"
)

type BalanceService struct {
    walletRepo      *repository.WalletRepository
    dbConn          *sqlx.DB
    redisClient     *redis.Client
}

func NewBalanceService(dbConn *sqlx.DB, redisClient *redis.Client) *BalanceService {
    walletRepo := repository.NewWalletRepository(dbConn)

    return &BalanceService{
        walletRepo:      walletRepo,
        dbConn:          dbConn,
        redisClient:     redisClient,
    }
}

// The GetBalance function retrieves the user's balance. It first attempts to obtain the balance from the Redis cache. If the data is not found in the Redis cache (a cache miss occurs), it then queries the database to get the balance.
func (s *BalanceService) GetBalance(ctx context.Context, userID int) (decimal.Decimal, error) {
    cacheKey := fmt.Sprintf("balance:%d", userID)

    // Attempt to retrieve the balance from the Redis cache
    cacheValue, err := s.redisClient.Get(ctx, cacheKey).Result()
    if err == redis.Nil {
        // If the Redis cache misses, query from the database
        user, err := s.walletRepo.GetUserBalance(ctx, userID)
        if err != nil {
            log.Printf("Failed to get balance from DB for user %d: %v", userID, err)
            return decimal.Zero, fmt.Errorf("failed to get balance from DB: %w", err)
        }

        // Update the Redis cache
        if err := s.redisClient.Set(ctx, cacheKey, user.Balance.String(), time.Second*3600).Err(); err != nil {
            log.Printf("Failed to cache balance for user %d in Redis: %v", userID, err)
            return decimal.Zero, fmt.Errorf("failed to cache balance: %w", err)
        }

        return user.Balance, nil
    } else if err != nil {
        // Other errors occurred while retrieving from Redis
        log.Printf("Failed to get balance from Redis for user %d: %v", userID, err)
        return decimal.Zero, fmt.Errorf("failed to get balance from Redis: %w", err)
    }

    // If the Redis cache hits, parse the balance in the cache
    cacheBalance, err := decimal.NewFromString(cacheValue)
    if err != nil {
        log.Printf("Invalid balance format in Redis cache for user %d: %v", userID, err)
        return decimal.Zero, fmt.Errorf("invalid balance in cache: %w", err)
    }

    return cacheBalance, nil
}
