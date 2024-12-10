// Package api provides the HTTP handlers and routes for the wallet service,
// including the logic for interacting with users' wallets, processing transactions,
// and managing requests related to deposits, withdrawals, and transfers by William, way1910@gmail.com.
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/yaoweihua/wallet-service/handler"
    "github.com/yaoweihua/wallet-service/service"
    "github.com/jmoiron/sqlx"
    "github.com/go-redis/redis/v8"
)

// SetupRoutes sets up the Gin routes.
func SetupRoutes(r *gin.Engine, dbConn *sqlx.DB, redisClient *redis.Client) {
    // Initialize the Service layer and pass the redisClient.
    depositService := service.NewDepositService(dbConn, redisClient)
    withdrawService := service.NewWithdrawService(dbConn, redisClient)
    transferService := service.NewTransferService(dbConn, redisClient)
    balanceService := service.NewBalanceService(dbConn, redisClient)
    transactionService := service.NewTransactionService(dbConn)

    // Initialize Handlers
    depositHandler := handler.NewDepositHandler(depositService)
    withdrawHandler := handler.NewWithdrawHandler(withdrawService)
    transferHandler := handler.NewTransferHandler(transferService)
    balanceHandler := handler.NewBalanceHandler(balanceService)
    transactionHandler := handler.NewTransactionHandler(transactionService)

    // Configure the routes.
    v1 := r.Group("/v1/wallet")
    {
        v1.POST("/deposit", depositHandler.HandleDeposit)
        v1.POST("/withdraw", withdrawHandler.HandleWithdraw)
        v1.POST("/transfer", transferHandler.HandleTransfer)
        v1.GET("/:user_id/balance", balanceHandler.HandleGetBalance)
        v1.GET("/:user_id/transactions", transactionHandler.HandleGetTransactions)
    }
}
