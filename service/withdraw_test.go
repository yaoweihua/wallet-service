package service

import (
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/require"
    "github.com/jmoiron/sqlx"
    redismock "github.com/go-redis/redismock/v8"
    "fmt"
    "time"
)

func TestWithdrawService_Withdraw_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close() // nolint:errcheck

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close() // nolint:errcheck

    // Set the Redis operation expectations
    mockRedis.ExpectSet("balance:1", "100", time.Second*3600).SetVal("")

    // Create an instance of the withdrawal service and pass in the mock Redis client
    withdrawService := NewWithdrawService(sqlx.NewDb(db, "sqlmock"), redisClient)

    mock.ExpectBegin()

    // The expectation of querying the current balance
    mock.ExpectQuery("SELECT id, balance FROM users WHERE id = \\$1").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow(1, decimal.NewFromInt(150)))

    // The expectation of updating the balance
    mock.ExpectExec("UPDATE users SET balance = \\$1, updated_at = \\$2 WHERE id = \\$3").
        WithArgs(decimal.NewFromInt(100), sqlmock.AnyArg(), 1).
        WillReturnResult(sqlmock.NewResult(1, 1))

    // The expectation of inserting the transaction record
    mock.ExpectExec("INSERT INTO transactions").
        WithArgs(1, 0, decimal.NewFromInt(50), "withdraw", "completed", decimal.NewFromFloat(0.0), "credit_card").
        WillReturnResult(sqlmock.NewResult(1, 1))

    mock.ExpectCommit()

    err = withdrawService.Withdraw(1, decimal.NewFromInt(50))
    require.NoError(t, err)

    // Check whether all the expectations are fully matched
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)

    // Check whether the Redis expectations are matched
    err = mockRedis.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestWithdrawService_Withdraw_InsufficientBalance(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close() // nolint:errcheck

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close() // nolint:errcheck

    withdrawService := NewWithdrawService(sqlx.NewDb(db, "sqlmock"), redisClient)

    mock.ExpectBegin()

    // The expectation of querying the current balance (with insufficient balance)
    mock.ExpectQuery("SELECT id, balance FROM users WHERE id = \\$1").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow(1, decimal.NewFromInt(30)))

    // The withdrawal amount is greater than the current balance
    err = withdrawService.Withdraw(1, decimal.NewFromInt(50))

    require.Error(t, err)
    require.Equal(t, "Insufficient balance", err.Error())

    // Check whether the expectations have not been triggered
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)

    // Check whether the Redis expectations have not been triggered
    err = mockRedis.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestWithdrawService_Withdraw_InvalidAmount(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close() // nolint:errcheck

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close() // nolint:errcheck

    withdrawService := NewWithdrawService(sqlx.NewDb(db, "sqlmock"), redisClient)

    // Invalid amount test
    invalidAmounts := []decimal.Decimal{
        decimal.NewFromInt(0),    // The withdrawal amount is zero
        decimal.NewFromInt(-50),  // The withdrawal amount is negative
    }

    for _, amount := range invalidAmounts {
        t.Run(fmt.Sprintf("withdraw amount: %s", amount.String()), func(t *testing.T) {
            err := withdrawService.Withdraw(1, amount)

            require.Error(t, err)
            require.Equal(t, "Withdraw amount must be greater than zero", err.Error())

            // Check whether the database expectations have not been triggered
            err = mock.ExpectationsWereMet()
            require.NoError(t, err)

            // Check whether the Redis expectations have not been triggered
            err = mockRedis.ExpectationsWereMet()
            require.NoError(t, err)
        })
    }
}
