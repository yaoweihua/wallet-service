package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/yaoweihua/wallet-service/service"
    "github.com/shopspring/decimal"
)

// WithdrawHandler handles the withdrawal request
type WithdrawHandler struct {
    withdrawService *service.WithdrawService
}

// NewWithdrawHandler creates a new instance of WithdrawHandler
func NewWithdrawHandler(withdrawService *service.WithdrawService) *WithdrawHandler {
    return &WithdrawHandler{withdrawService: withdrawService}
}

// HandleWithdraw handles the withdrawal HTTP request
func (h *WithdrawHandler) HandleWithdraw(c *gin.Context) {
    var req struct {
        UserID int             `json:"user_id"`
        Amount decimal.Decimal `json:"amount"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        sendResponse(c, http.StatusBadRequest, "", "Invalid request")
        return
    }

    // Call the service layer to handle the withdrawal logic
    err := h.withdrawService.Withdraw(req.UserID, req.Amount)
    if err != nil {
        sendResponse(c, http.StatusBadRequest, "", err.Error())
        return
    }

    sendResponse(c, http.StatusOK, "", "Withdraw successful")
}
