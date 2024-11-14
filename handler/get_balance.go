// handler/balance_handler.go

package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/yaoweihua/wallet-service/service"
    //"github.com/shopspring/decimal"
    "strconv"
)

// APIResponse represents the structure of a response from the API.
// It includes the status code, response data (which can be of various types),
// and an optional error message.
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

// BalanceHandler handles HTTP requests related to user balance operations.
// It interacts with the BalanceService to perform actions like retrieving and updating balances.
type BalanceHandler struct {
    balanceService *service.BalanceService
}

// BalanceResponse represents the structure of the response for a user's balance.
// It includes the user ID and the current balance of the user as a string.
type BalanceResponse struct {
    UserID  int     `json:"user_id"`
    Balance string `json:"balance"`
}

// NewBalanceHandler creates a new instance of BalanceHandler with the given BalanceService.
// It initializes the handler with the provided service, which will be used to handle balance-related operations.
func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
    return &BalanceHandler{balanceService: balanceService}
}

// HandleGetBalance handles the HTTP request to get the balance of a user.
// It extracts the user ID from the context, retrieves the balance from the service,
// and sends the response back to the client. If there is an error, it sends an appropriate error message.
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
        UserID:  userID,
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