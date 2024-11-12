// handler/balance_handler.go

package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/yaoweihua/wallet-service/service"
    //"github.com/shopspring/decimal"
    "strconv"
)

type APIResponse struct {
    Status  int         `json:"status"`     // Status code
    Data    interface{} `json:"data"`       // The data part can be a string, number, list, or map
    ErrMsg  string      `json:"errmsg"`     // The error message, empty if there is no error
}

func sendResponse(c *gin.Context, status int, data interface{}, errMsg string) {
    // If there is no error message, it defaults to empty
    if errMsg == "" {
        errMsg = ""
    }

    // If data is nil, set it to an empty string
    if data == nil {
        data = ""
    }

    response := APIResponse{
        Status: status,
        Data:   data,
        ErrMsg: errMsg,
    }

    c.JSON(status, response)
}

type BalanceHandler struct {
    balanceService *service.BalanceService
}

type BalanceResponse struct {
    UserId  int     `json:"user_id"`
    Balance string `json:"balance"`
}

func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
    return &BalanceHandler{balanceService: balanceService}
}

func (h *BalanceHandler) HandleGetBalance(c *gin.Context) {
    userID, err := getUserIDFromContext(c)
    if err != nil {
        sendResponse(c, http.StatusBadRequest, nil, "Invalid user ID")
        return
    }

    balance, err := h.balanceService.GetBalance(c, userID)
    if err != nil {
        sendResponse(c, http.StatusInternalServerError, nil, err.Error())
        return
    }

    data := BalanceResponse{
        UserId:  userID,
        Balance: balance.String(),
    }

    sendResponse(c, http.StatusOK, data, "")
}

func getUserIDFromContext(c *gin.Context) (int, error) {
    userIDStr := c.Param("user_id")
    userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        return 0, err
    }
    return userID, nil
}