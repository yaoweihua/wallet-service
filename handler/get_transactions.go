package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/yaoweihua/wallet-service/service"
    "github.com/yaoweihua/wallet-service/model"
)

// TransactionHandler handles HTTP requests related to transactions.
// It interacts with the TransactionService to perform operations such as retrieving, creating, and managing transactions.
type TransactionHandler struct {
    transactionService *service.TransactionService
}

// NewTransactionHandler creates a new instance of TransactionHandler with the provided TransactionService.
// This handler is responsible for handling transaction-related requests and interacting with the transaction service.
func NewTransactionHandler(transactionService *service.TransactionService) *TransactionHandler {
    return &TransactionHandler{
        transactionService: transactionService,
    }
}

// TransactionsResponse represents the structure of the response that contains a user's transaction records.
// It includes the user ID and a list of transactions associated with that user.
type TransactionsResponse struct {
    UserID       int     `json:"user_id"`
    Transactions []model.Transaction `json:"transactions"`
}

// HandleGetTransactions handles the HTTP request to retrieve a user's transaction records.
// It extracts the user ID from the URL parameters, retrieves the transactions from the service layer,
// and sends the response back to the client. If there is an error or no transactions are found,
// it returns an appropriate error message or an empty transaction list.
func (h *TransactionHandler) HandleGetTransactions(c *gin.Context) {
    // Retrieve the userID from the URL parameters
    userIDStr := c.Param("user_id")
    userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        sendResponse(c, http.StatusBadRequest, nil, "Invalid user ID")
        return
    }

    // Call the service layer's GetTransactions method to retrieve the transaction records
    transactions, err := h.transactionService.GetTransactions(c, userID)
    if err != nil {
        sendResponse(c, http.StatusInternalServerError, nil, err.Error())
        return
    }

    if len(transactions) == 0 {
        transactions = []model.Transaction{}
    }

    data := TransactionsResponse{
        UserID:  userID,
        Transactions: transactions,
    }

    sendResponse(c, http.StatusOK, data, "")
}
