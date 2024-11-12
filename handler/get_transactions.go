package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/yaoweihua/wallet-service/service"
    "github.com/yaoweihua/wallet-service/model"
)

type TransactionHandler struct {
    transactionService *service.TransactionService
}

func NewTransactionHandler(transactionService *service.TransactionService) *TransactionHandler {
    return &TransactionHandler{
        transactionService: transactionService,
    }
}

type TransactionsResponse struct {
    UserId       int     `json:"user_id"`
    Transactions []model.Transaction `json:"transactions"`
}

// GetTransactionsHandler retrieves the transaction records of a specified user
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
        UserId:  userID,
        Transactions: transactions,
    }

    sendResponse(c, http.StatusOK, data, "")
}
