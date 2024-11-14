package service

import (
    "context"
    "fmt"
    "github.com/yaoweihua/wallet-service/repository"
    "github.com/go-redis/redis/v8"
    "github.com/shopspring/decimal"
    "github.com/jmoiron/sqlx"
    "time"
)

// BalanceService provides services related to user balance operations.
// It interacts with the WalletRepository to manage user wallet data,
// and uses the database and Redis client for storing and retrieving balance information.
type BalanceService struct {
    walletRepo      *repository.WalletRepository
    dbConn          *sqlx.DB
    redisClient     *redis.Client
}

// NewBalanceService creates a new instance of BalanceService with the provided database connection and Redis client.
// It initializes the service with a WalletRepository to handle wallet-related database operations.
func NewBalanceService(dbConn *sqlx.DB, redisClient *redis.Client) *BalanceService {
    walletRepo := repository.NewWalletRepository(dbConn)

    return &BalanceService{
        walletRepo:      walletRepo,
        dbConn:          dbConn,
        redisClient:     redisClient,
    }
}

// GetBalance function retrieves the user's balance. It first attempts to obtain the balance from the Redis cache. If the data is not found in the Redis cache (a cache miss occurs), it then queries the database to get the balance.
func (s *BalanceService) GetBalance(ctx context.Context, userID int) (decimal.Decimal, error) {
    cacheKey := fmt.Sprintf("balance:%d", userID)

    // Attempt to retrieve the balance from the Redis cache
    cacheValue, err := s.redisClient.Get(ctx, cacheKey).Result()
    if err == redis.Nil {
        // If the Redis cache misses, query from the database
        user, err := s.walletRepo.GetUserBalance(ctx, userID)
        if err != nil {
            return decimal.Zero, fmt.Errorf("failed to get balance from DB: %w", err)
        }

        // Update the Redis cache
        if err := s.redisClient.Set(ctx, cacheKey, user.Balance.String(), time.Second*3600).Err(); err != nil {
            return decimal.Zero, fmt.Errorf("failed to cache balance: %w", err)
        }

        return user.Balance, nil
    } else if err != nil {
        // Other errors occurred while retrieving from Redis
        return decimal.Zero, fmt.Errorf("failed to get balance from Redis: %w", err)
    }

    // If the Redis cache hits, parse the balance in the cache
    cacheBalance, err := decimal.NewFromString(cacheValue)
    if err != nil {
        return decimal.Zero, fmt.Errorf("invalid balance in cache: %w", err)
    }

    return cacheBalance, nil
}
