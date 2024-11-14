// handler/transfer_handler.go

package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/yaoweihua/wallet-service/service"
    "github.com/shopspring/decimal"
)

// TransferHandler handles HTTP requests related to fund transfers between users.
// It interacts with the TransferService to perform actions like initiating transfers and checking transfer status.
type TransferHandler struct {
    transferService *service.TransferService
}

// NewTransferHandler creates a new instance of TransferHandler with the provided TransferService.
// This handler is responsible for handling transfer requests and interacting with the transfer service.
func NewTransferHandler(transferService *service.TransferService) *TransferHandler {
    return &TransferHandler{transferService: transferService}
}

// HandleTransfer handles the HTTP request to transfer funds between two users.
// It parses the transfer request, validates the input, and calls the service layer to process the transfer.
// If the transfer is successful, it sends a success response; otherwise, it returns an error message.
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
