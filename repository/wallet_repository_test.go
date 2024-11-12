package repository

import (
    "context"
    "fmt"
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/require"
    "github.com/jmoiron/sqlx"
    "github.com/sirupsen/logrus"
    "database/sql"
    "github.com/yaoweihua/wallet-service/model"
    "io"
)

func NewTestLogger() *logrus.Logger {
    logger := logrus.New()
    logger.SetLevel(logrus.DebugLevel)
    return logger
}

// Test the scenario where retrieving the user's balance is successful
func TestGetUserBalance(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    logger := NewTestLogger()

    r := &WalletRepository{
        DB:     sqlx.NewDb(db, "sqlmock"),
        Logger: logger,
    }

    userID := 1
    expectedUser := model.User{
        ID:      userID,
        Balance: decimal.NewFromInt(100),
    }

    mock.ExpectQuery("SELECT id, balance FROM users WHERE id = \\$1 FOR UPDATE").
        WithArgs(userID).
        WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).
            AddRow(expectedUser.ID, expectedUser.Balance))

    user, err := r.GetUserBalance(context.Background(), userID)
    require.NoError(t, err)
    require.NotNil(t, user)
    require.Equal(t, expectedUser.ID, user.ID)
    require.Equal(t, expectedUser.Balance, user.Balance)

    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}

// Test the situation where retrieving the user's balance fails because the user does not exist
func TestGetUserBalanceNoFound(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    logger := logrus.New()
    logger.SetOutput(io.Discard)

    r := &WalletRepository{
        DB:     sqlx.NewDb(db, "sqlmock"),
        Logger: logger,
    }

    userID := 1

    mock.ExpectQuery("SELECT id, balance FROM users WHERE id = \\$1 FOR UPDATE").
        WithArgs(userID).
        WillReturnError(sql.ErrNoRows)

    user, err := r.GetUserBalance(context.Background(), userID)
    require.Error(t, err)
    require.Nil(t, user)
    require.Contains(t, err.Error(), "user 1 not found")

    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestUpdateBalance(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    r := &WalletRepository{
        DB: sqlx.NewDb(db, "sqlmock"),
        Logger: NewTestLogger(),
    }

    userID := 1
    newBalance := decimal.NewFromInt(150)

    mock.ExpectBegin()
    // Use AnyTime() to allow time matching while ignoring minor differences
    mock.ExpectExec("UPDATE users SET balance = \\$1, updated_at = \\$2 WHERE id = \\$3").
        WithArgs(newBalance, sqlmock.AnyArg(), userID).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    tx, err := r.DB.Beginx()
    require.NoError(t, err)
    err = r.UpdateBalance(context.Background(), tx, userID, newBalance)
    require.NoError(t, err)

    err = tx.Commit()
    require.NoError(t, err)

    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}

func TestUpdateBalance_Error(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    logger := logrus.New()
    logger.SetOutput(io.Discard)

    r := &WalletRepository{
        DB: sqlx.NewDb(db, "sqlmock"),
        Logger: logger,
    }

    userID := 1
    newBalance := decimal.NewFromInt(150)

    mock.ExpectBegin()
    // Use AnyTime() to allow time matching while ignoring minor differences
    mock.ExpectExec("UPDATE users SET balance = \\$1, updated_at = \\$2 WHERE id = \\$3").
        WithArgs(newBalance, sqlmock.AnyArg(), userID).
        WillReturnError(fmt.Errorf("DB error"))
    mock.ExpectRollback()

    tx, err := r.DB.Beginx()
    require.NoError(t, err)
    err = r.UpdateBalance(context.Background(), tx, userID, newBalance)
    require.Error(t, err)
    require.Contains(t, err.Error(), "failed to update balance for user 1")

    err = tx.Rollback()
    require.NoError(t, err)

    err = mock.ExpectationsWereMet()
    require.NoError(t, err)
}

