// handler/transfer_handler.go

package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/yaoweihua/wallet-service/service"
    "github.com/shopspring/decimal"
)

type TransferHandler struct {
    transferService *service.TransferService
}

func NewTransferHandler(transferService *service.TransferService) *TransferHandler {
    return &TransferHandler{transferService: transferService}
}

func (h *TransferHandler) HandleTransfer(c *gin.Context) {
    var req struct {
        FromUserID int             `json:"from_user_id"`
        ToUserID   int             `json:"to_user_id"`
        Amount     decimal.Decimal `json:"amount"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        sendResponse(c, http.StatusBadRequest, "", "Invalid request")
        return
    }

    // Call the service layer to execute the transfer logic
    err := h.transferService.Transfer(req.FromUserID, req.ToUserID, req.Amount)
    if err != nil {
        sendResponse(c, http.StatusBadRequest, "", err.Error())
        return
    }

    sendResponse(c, http.StatusOK, "", "Transfer successful")
}
