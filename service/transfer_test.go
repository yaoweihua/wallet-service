package service

import (
    "fmt"
    "testing"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/require"
    "github.com/jmoiron/sqlx"
    "github.com/DATA-DOG/go-sqlmock"
    redismock "github.com/go-redis/redismock/v8"
    "time"
)

func TestTransferService_Transfer_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close()

    // Set the expectations for the Redis update operation: update the balances of the two users
    mockRedis.ExpectSet("balance:1", "100", time.Second*3600).SetVal("")
    mockRedis.ExpectSet("balance:2", "250", time.Second*3600).SetVal("")

    // Create an instance of TransferService, passing in the mock DB and Redis client
    transferService := NewTransferService(sqlx.NewDb(db, "sqlmock"), redisClient)

    mock.ExpectBegin()

    // Query the balance of User 1 (the one initiating the transfer)
    mock.ExpectQuery("SELECT id, balance FROM users WHERE id = \\$1").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow(1, decimal.NewFromInt(200)))

    // Update the balance of User 1 (the one who initiates the transfer)
    mock.ExpectExec("UPDATE users SET balance = \\$1, updated_at = \\$2 WHERE id = \\$3").
        WithArgs(decimal.NewFromInt(100), sqlmock.AnyArg(), 1).
        WillReturnResult(sqlmock.NewResult(1, 1))

    // Query the balance of User 2 (the one receiving the transfer)
    mock.ExpectQuery("SELECT id, balance FROM users WHERE id = \\$1").
        WithArgs(2).
        WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow(2, decimal.NewFromInt(150)))

    // Update the balance of User 2 (the one receiving the transfer)
    mock.ExpectExec("UPDATE users SET balance = \\$1, updated_at = \\$2 WHERE id = \\$3").
        WithArgs(decimal.NewFromInt(250), sqlmock.AnyArg(), 2).
        WillReturnResult(sqlmock.NewResult(1, 1))

    // The expectation of inserting a transaction record
    mock.ExpectExec("INSERT INTO transactions").
        WithArgs(1, 2, decimal.NewFromInt(100), "transfer", "completed", decimal.NewFromFloat(0.0), "credit_card").
        WillReturnResult(sqlmock.NewResult(1, 1))

    mock.ExpectCommit()

    // Call the transfer method
    err = transferService.Transfer(1, 2, decimal.NewFromInt(100))
    require.NoError(t, err)

    // Check whether all the expectations are met
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)

    // Check whether the expectations of Redis are met
    err = mockRedis.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestTransferService_Transfer_InsufficientBalance(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close()

    // 创建 TransferService 实例，传入 mock DB 和 mock Redis 客户端
    transferService := NewTransferService(sqlx.NewDb(db, "sqlmock"), redisClient)

    // Set the database expectations
    mock.ExpectBegin()

    // Set the expectation for querying the balance of the transferring-out user (insufficient balance)
    mock.ExpectQuery("SELECT id, balance FROM users WHERE id = \\$1").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow(1, decimal.NewFromInt(30)))

    // Call the transfer method (with insufficient balance for transfer)
    err = transferService.Transfer(1, 2, decimal.NewFromInt(100))

    // Verify the returned error message
    require.Error(t, err)
    require.Equal(t, "insufficient balance for transfer", err.Error())

    // Ensure that no database operations have been carried out (no balance updates or transaction record insertions)
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)

    // Ensure that Redis operations have not been triggered since the transfer was not successful
    err = mockRedis.ExpectationsWereMet()
    require.NoError(t, err)
}


func TestTransferService_Transfer_ToSameUser(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close()

    // Create an instance of TransferService, passing in the mock DB and mock Redis client
    transferService := NewTransferService(sqlx.NewDb(db, "sqlmock"), redisClient)

    // Call the transfer method (transferring to the same user)
    err = transferService.Transfer(1, 1, decimal.NewFromInt(50))

    // Verify the returned error message
    require.Error(t, err)
    require.Equal(t, "cannot transfer to the same user", err.Error())

    // Check whether the expectations have not been triggered
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)

    // Check whether the Redis expectations have not been triggered
    err = mockRedis.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestTransferService_Transfer_InvalidAmount(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    redisClient, mockRedis := redismock.NewClientMock()
    defer redisClient.Close()

    transferService := NewTransferService(sqlx.NewDb(db, "sqlmock"), redisClient)

    // Invalid amount test
    invalidAmounts := []decimal.Decimal{
        decimal.NewFromInt(0),    // The transfer amount is zero
        decimal.NewFromInt(-50),  // The transfer amount is minus value
    }

    for _, amount := range invalidAmounts {
        t.Run(fmt.Sprintf("transfer amount: %s", amount.String()), func(t *testing.T) {
            // Call the transfer method
            err := transferService.Transfer(1, 2, amount)

            // Verify whether an error has been returned
            require.Error(t, err)
            require.Equal(t, "transfer amount must be greater than zero", err.Error())

            // Check whether the database expectations have not been triggered
            err = mock.ExpectationsWereMet()
            require.NoError(t, err)

            // Check whether the Redis expectations have not been triggered
            err = mockRedis.ExpectationsWereMet()
            require.NoError(t, err)
        })
    }
}
