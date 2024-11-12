package repository

import (
    "context"
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/require"
    "github.com/jmoiron/sqlx"
    "time"
    "fmt"
    "github.com/sirupsen/logrus"
    "io"
)

func TestRecordTransaction(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    txRepo := &TransactionRepository{
        DB:     sqlx.NewDb(db, "sqlmock"),
        Logger: NewTestLogger(),
    }

    // Simulated transaction data
    fromUserID := 1
    toUserID := 2
    amount := decimal.NewFromFloat(100.5)
    transactionType := "transfer"
    transactionFee := decimal.NewFromFloat(0.0)
    paymentMethod := "credit_card"

    // Set the BEGIN operation of the simulated database
    mock.ExpectBegin()

    // Simulate the execution of the SQL for inserting transactions and ensure that the number of parameters matches the SQL
    mock.ExpectExec("INSERT INTO transactions").
        WithArgs(
            fromUserID, toUserID, amount, transactionType, "completed", transactionFee, paymentMethod,
        ).
        WillReturnResult(sqlmock.NewResult(1, 1))  // Simulate a successful insertion

    // Set the expectation of committing the transaction
    mock.ExpectCommit()

    // Begin the transaction
    tx, err := txRepo.DB.Beginx()
    require.NoError(t, err)

    // Call the RecordTransaction method
    err = txRepo.RecordTransaction(context.Background(), tx, fromUserID, toUserID, amount, transactionType, "completed")
    require.NoError(t, err)

    // Commit the transaction
    err = tx.Commit()
    require.NoError(t, err)

    // Verify whether the expectation has been met
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}

// TestGetTransactions tests the GetTransactions method
func TestGetTransactions(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    sqlxDB := sqlx.NewDb(db, "postgres")

    txRepo := NewTransactionRepository(sqlxDB)

    // Set the query results to be returned by the simulation
    rows := sqlmock.NewRows([]string{
        "id", "from_user_id", "to_user_id", "amount", "transaction_type", "transaction_status", "transaction_fee", "payment_method", "created_at", "updated_at",
    }).AddRow(
        1, 1, 2, decimal.NewFromFloat(100.5), "transfer", "completed", decimal.NewFromFloat(0.0), "credit_card", time.Now(), time.Now(),
    ).AddRow(
        2, 2, 3, decimal.NewFromFloat(50.5), "withdraw", "completed", decimal.NewFromFloat(1.0), "bank_transfer", time.Now(), time.Now(),
    )

    // Set the query expectations
    mock.ExpectQuery("SELECT").WithArgs(1).WillReturnRows(rows)

    // Call the method and verify
    transactions, err := txRepo.GetTransactions(context.Background(), 1)
    require.NoError(t, err)
    require.Len(t, transactions, 2)

    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestRecordTransactionError(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    logger := logrus.New()
    logger.SetOutput(io.Discard)

    txRepo := &TransactionRepository{
        DB:     sqlx.NewDb(db, "sqlmock"),
        Logger: logger,
    }

    // Simulated transaction data
    fromUserID := 1
    toUserID := 2
    amount := decimal.NewFromFloat(100.5)
    transactionType := "transfer"
    transactionFee := decimal.NewFromFloat(0.0)
    paymentMethod := "credit_card"

    // Set the BEGIN operation of the simulated database
    mock.ExpectBegin()

    // Simulate an error occurring during the execution of the SQL for inserting transactions
    mock.ExpectExec(`INSERT INTO transactions`).
        WithArgs(
            fromUserID, toUserID, amount, transactionType, "completed", transactionFee, paymentMethod,
        ).
        WillReturnError(fmt.Errorf("DB insert error"))

    // Set the expectation for the transaction to roll back
    mock.ExpectRollback()

    // Begin the transaction
    tx, err := txRepo.DB.Beginx()
    require.NoError(t, err)

    // Call the RecordTransaction method and verify the error
    err = txRepo.RecordTransaction(context.Background(), tx, fromUserID, toUserID, amount, transactionType, "completed")
    require.Error(t, err)
    require.Contains(t, err.Error(), "DB insert error")

    // Rollback the transaction
    err = tx.Rollback()
    require.NoError(t, err)

    // Verify whether the expectation has been met
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestGetTransactionsNoFound(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    sqlxDB := sqlx.NewDb(db, "postgres")

    txRepo := NewTransactionRepository(sqlxDB)

    // Simulate the database query to return empty results
    mock.ExpectQuery(`SELECT`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{
        "id", "from_user_id", "to_user_id", "amount", "transaction_type", "transaction_status", "transaction_fee", "payment_method", "created_at", "updated_at",
    }))

    // Call the GetTransactions method and verify that the returned result is empty
    transactions, err := txRepo.GetTransactions(context.Background(), 1)
    require.NoError(t, err)
    require.Empty(t, transactions)

    // Verify whether the expectation has been met
    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}
