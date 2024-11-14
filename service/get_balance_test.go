package service

import (
    "context"
    "testing"
    "github.com/shopspring/decimal"
    "github.com/jmoiron/sqlx"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/go-redis/redis/v8"
    "github.com/go-redis/redismock/v8"
    "github.com/stretchr/testify/require"
    "time"
)

func TestBalanceService_GetBalance_CacheMiss(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close() // nolint:errcheck

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close() // nolint:errcheck

    // Set the expectation for the Redis GET operation: Return redis.Nil when the cache is not hit
    mockRedis.ExpectGet("balance:1").SetErr(redis.Nil)

    // Simulate querying the balance from the database and return a balance with decimals
    mock.ExpectQuery("SELECT id, balance FROM users WHERE id = \\$1 FOR UPDATE").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow(1, decimal.NewFromInt(100)))

    mockRedis.ExpectSet("balance:1", "100", time.Second*3600).SetVal("OK")
    
    balanceService := NewBalanceService(sqlx.NewDb(db, "sqlmock"), redisClient)
    balance, err := balanceService.GetBalance(context.Background(), 1)

    require.NoError(t, err)

    expectedBalance, err := decimal.NewFromString("100")
    require.NoError(t, err)

    // Convert the Decimal to a string when making the comparison
    require.Equal(t, expectedBalance.String(), balance.String())

    // Check whether the expectations of the SQL mock and the Redis mock have both been met
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)

    err = mockRedis.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestBalanceService_GetBalance_CacheHit(t *testing.T) {
    db, _, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close() // nolint:errcheck

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close() // nolint:errcheck

    mockRedis.ExpectGet("balance:1").SetVal("100")

    // Create an instance of BalanceService
    balanceService := NewBalanceService(sqlx.NewDb(db, "sqlmock"), redisClient)

    // Call GetBalance, expecting to retrieve the balance from the Redis cache
    balance, err := balanceService.GetBalance(context.Background(), 1)

    require.NoError(t, err)

    // The expected value is the one retrieved from Redis
    expectedBalance, err := decimal.NewFromString("100")
    require.NoError(t, err)

    // When comparing, convert the Decimal to a string for comparison
    require.Equal(t, expectedBalance.String(), balance.String())

    // Check whether the expectations of the Redis mock have been met
    err = mockRedis.ExpectationsWereMet()
    require.NoError(t, err)
}

