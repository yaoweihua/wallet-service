// main.go by William, way1910@gmail.com.
package main

import (
    "os"
    "os/signal"
    "syscall"
    "time"
    "context"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "github.com/yaoweihua/wallet-service/config"
    "github.com/yaoweihua/wallet-service/db"
    "github.com/yaoweihua/wallet-service/api"
    "github.com/yaoweihua/wallet-service/utils"
    "net/http"
)

func main() {
    // Initialize the logger
    logger := utils.GetLogger()

    // Load the configuration
    cfg := config.LoadConfig()

    // Connect to the PostgreSQL database
    dbConn, err := db.ConnectPostgres(cfg)
    if err != nil {
        logger.Fatal("PostgreSQL connection failed:", err)
    }

    // Connect to Redis
    redisClient, err := db.ConnectRedis(cfg)
    if err != nil {
        logger.Fatal("Redis connection failed:", err)
    }

    // Initialize the Gin router
    r := gin.Default()

    // Set up the CORS middleware
    r.Use(cors.Default())

    // Set up the routes
    api.SetupRoutes(r, dbConn, redisClient)

    // Get the port configuration
    port := getPort()

    // Create an instance of http.Server
    server := &http.Server{
        Addr:    ":" + port,
        Handler: r,
    }

    // Capture the system interrupt signal to gracefully shut down the server
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        logger.Infof("Starting server on port %s", port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Server failed to start:", err)
        }
    }()

    // Wait for the interrupt signal to gracefully shut down the server
    <-quit
    logger.Info("Shutting down server...")

    // Set a timeout to ensure that it doesn't wait too long during shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        logger.Fatal("Server forced to shutdown:", err)
    }

    logger.Info("Server exited properly")
}

func getPort() string {
    port := os.Getenv("PORT")
    if port == "" {
        return "8080"  // Running on port 8080
    }
    return port
}
