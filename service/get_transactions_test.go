package service

import (
    "context"
    "fmt"
    "testing"
    "github.com/shopspring/decimal"
    "github.com/jmoiron/sqlx"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/require"
    "time"
)

func TestTransactionService_GetTransactions(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close() // nolint:errcheck

    sqlxDB := sqlx.NewDb(db, "sqlmock")

    // Create an instance of TransactionService
    transactionService := NewTransactionService(sqlxDB)

    // Simulate returning the transaction records
    createdAt1, err := time.Parse("2006-01-02", "2024-11-12")
    require.NoError(t, err)
    createdAt2, err := time.Parse("2006-01-02", "2024-11-13")
    require.NoError(t, err)

    rows := sqlmock.NewRows([]string{
        "id", "from_user_id", "to_user_id", "amount", "transaction_type", 
        "transaction_status", "created_at", "updated_at"}).
        AddRow(1, 1, 0, "100", "deposit", "completed", createdAt1, createdAt1).
        AddRow(2, 1, 0, "50", "withdraw", "completed", createdAt2, createdAt2)

    // Set the expected SQL query and ensure that the column fields are consistent
    mock.ExpectQuery("SELECT id, from_user_id, to_user_id, amount, transaction_type, transaction_status, created_at, updated_at FROM transactions WHERE from_user_id = \\$1 OR to_user_id = \\$1 ORDER BY created_at DESC").
        WithArgs(1).
        WillReturnRows(rows)

    // Call the GetTransactions method
    transactions, err := transactionService.GetTransactions(context.Background(), 1)
    require.NoError(t, err)

    // Check the returned transaction records
    require.Len(t, transactions, 2)

    // Check the first transaction record
    require.Equal(t, 1, transactions[0].ID)
    require.Equal(t, 1, transactions[0].FromUserID)
    require.Equal(t, 0, transactions[0].ToUserID)
    require.Equal(t, decimal.NewFromInt(100), transactions[0].Amount)
    require.Equal(t, "deposit", transactions[0].TransactionType)
    require.Equal(t, "completed", transactions[0].TransactionStatus)

    // Check the second transaction record
    require.Equal(t, 2, transactions[1].ID)
    require.Equal(t, 1, transactions[1].FromUserID)
    require.Equal(t, 0, transactions[1].ToUserID)
    require.Equal(t, decimal.NewFromInt(50), transactions[1].Amount)
    require.Equal(t, "withdraw", transactions[1].TransactionType)
    require.Equal(t, "completed", transactions[1].TransactionStatus)

    // Check whether all the expectations of the SQL mock have been met
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}

// TestTransactionService_GetTransactions_Error tests the scenario where an error occurs when retrieving transaction records
func TestTransactionService_GetTransactions_Error(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close() // nolint:errcheck

    sqlxDB := sqlx.NewDb(db, "sqlmock")

    // Create an instance of TransactionService
    transactionService := NewTransactionService(sqlxDB)

    // Set the expected SQL query and simulate a database query failure
    mock.ExpectQuery("SELECT id, from_user_id, to_user_id, amount, transaction_type, transaction_status, created_at, updated_at FROM transactions WHERE from_user_id = \\$1 OR to_user_id = \\$1 ORDER BY created_at DESC").
        WithArgs(1). // 用户ID为1
        WillReturnError(fmt.Errorf("database query failed"))

    // Call the GetTransactions method
    transactions, err := transactionService.GetTransactions(context.Background(), 1)

    // Expect the returned error to be not nil
    require.Error(t, err)
    require.Nil(t, transactions)
    require.Equal(t, "failed to fetch transactions from repository: database query failed", err.Error())

    // Check whether all the expectations of the SQL mock have been met
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}
