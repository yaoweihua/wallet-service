package service

import (
    "github.com/yaoweihua/wallet-service/repository"
    "github.com/yaoweihua/wallet-service/model"
    "context"
    "fmt"
    "github.com/jmoiron/sqlx"
)

type TransactionService struct {
    transactionRepo *repository.TransactionRepository
    dbConn          *sqlx.DB
}

func NewTransactionService(dbConn *sqlx.DB) *TransactionService {
    transactionRepo := repository.NewTransactionRepository(dbConn)

    return &TransactionService{
        transactionRepo: transactionRepo,
        dbConn:          dbConn,
    }
}

// GetTransactions retrieves the transaction records of the specified user
func (s *TransactionService) GetTransactions(ctx context.Context, userID int) ([]model.Transaction, error) {
    transactions, err := s.transactionRepo.GetTransactions(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch transactions from repository: %w", err)
    }

    return transactions, nil
}
