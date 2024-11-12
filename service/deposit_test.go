package service

import (
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/require"
    "github.com/jmoiron/sqlx"
    redismock "github.com/go-redis/redismock/v8"
    "time"
    "fmt"
)

func TestDepositService_Deposit_Success(t *testing.T) {
    // Create a mock DB
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    // Create a mock Redis client
    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close()

    // Set expectations for Redis operations
    mockRedis.ExpectSet("balance:1", "150", time.Second*3600).SetVal("")

    // Create an instance of the wallet deposit service and pass in the mock Redis client
    depositService := NewDepositService(sqlx.NewDb(db, "sqlmock"), redisClient)

    // Set database expectations
    mock.ExpectBegin()

    mock.ExpectQuery("SELECT id, balance FROM users WHERE id = \\$1").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow(1, decimal.NewFromInt(100)))

    // Expectations for transaction operations
    mock.ExpectExec("UPDATE users SET balance = \\$1, updated_at = \\$2 WHERE id = \\$3").
        WithArgs(decimal.NewFromInt(150), sqlmock.AnyArg(), 1).
        WillReturnResult(sqlmock.NewResult(1, 1))

    // Expectations for inserting transaction records
    mock.ExpectExec("INSERT INTO transactions").
        WithArgs(1, 0, decimal.NewFromInt(50), "deposit", "completed", decimal.NewFromFloat(0.0), "credit_card").
        WillReturnResult(sqlmock.NewResult(1, 1))

    mock.ExpectCommit()

    // Call the deposit method
    err = depositService.Deposit(1, decimal.NewFromInt(50))
    require.NoError(t, err)

    // Check if all the expectations are fully matched
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)

    // Check if the Redis expectations are matched
    err = mockRedis.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestDepositService_Deposit_InvalidAmount(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close()

    // Create an instance of the wallet deposit service and pass in the mock Redis client
    depositService := NewDepositService(sqlx.NewDb(db, "sqlmock"), redisClient)

    // Invalid amount test
    invalidAmounts := []decimal.Decimal{
        decimal.NewFromInt(0),    // The deposit amount is zero
        decimal.NewFromInt(-50),  // The deposit amount is a minus value
    }

    for _, amount := range invalidAmounts {
        t.Run(fmt.Sprintf("deposit amount: %s", amount.String()), func(t *testing.T) {
            // Here, directly verify whether the amount is valid. If it is not valid, return an error in advance
            err := depositService.Deposit(1, amount)

            // Verify whether an error has been returned
            require.Error(t, err)
            require.Equal(t, "deposit amount must be greater than zero", err.Error())

            // Check whether the database expectations have not been triggered
            err = mock.ExpectationsWereMet()
            require.NoError(t, err)

            // Check whether the Redis expectations have not been triggered
            err = mockRedis.ExpectationsWereMet()
            require.NoError(t, err)
        })
    }
}

