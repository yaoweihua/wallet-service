// Package handler contains the HTTP request handlers for the wallet service.
// It includes functions for processing user requests, interacting with
// the service layer, and returning appropriate responses for wallet operations 
// such as deposits, withdrawals, and transfers.
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/yaoweihua/wallet-service/service"
    "github.com/shopspring/decimal"
)

// DepositHandler handles deposit requests
type DepositHandler struct {
    depositService *service.DepositService
}

// NewDepositHandler creates a new instance of DepositHandler
func NewDepositHandler(depositService *service.DepositService) *DepositHandler {
    return &DepositHandler{depositService: depositService}
}

// HandleDeposit handles the deposit HTTP request
func (h *DepositHandler) HandleDeposit(c *gin.Context) {
    var req struct {
        UserID int             `json:"user_id"`
        Amount decimal.Decimal `json:"amount"`
    }

    // Bind the request parameters
    if err := c.ShouldBindJSON(&req); err != nil {
        sendResponse(c, http.StatusBadRequest, "", "Invalid request")
        return
    }

    // Call the service layer to handle the deposit logic
    err := h.depositService.Deposit(req.UserID, req.Amount)
    if err != nil {
        // Return the specific error message
        sendResponse(c, http.StatusBadRequest, "", err.Error())
        return
    }

    // Return a response on success
    sendResponse(c, http.StatusOK, "", "Deposit successful")
}
