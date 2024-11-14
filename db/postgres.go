// Package db provides database connection utilities and functions for 
// interacting with PostgreSQL and Redis. It includes connection pooling,
// configuration management, and automatic reconnection handling to 
// ensure reliable database access across the wallet service.
package db

import (
    "github.com/jmoiron/sqlx"

    // Import the pq driver for PostgreSQL initialization.
    _ "github.com/lib/pq"
    "github.com/yaoweihua/wallet-service/config"
    "time"
    "log"
    "sync"
)

var (
    db        *sqlx.DB
    dbOnce    sync.Once // Ensure the database connection is initialized only once.
    reconnect = make(chan bool, 1) // Used for handling reconnection notifications.
)

// ConnectPostgres initializes the PostgreSQL database connection pool.
func ConnectPostgres(cfg *config.Config) (*sqlx.DB, error) {
    dbOnce.Do(func() {
        var err error
        connStr := cfg.PostgreSQLURL

        // Attempt to connect to the database.
        db, err = sqlx.Connect("postgres", connStr)
        if err != nil {
            log.Fatalf("PostgreSQL connection failed: %v", err)
        }

        // Set the connection pool parameters.
        db.SetMaxIdleConns(10)                   // Maximum number of idle connections
        db.SetMaxOpenConns(100)                  // Maximum number of open connections
        db.SetConnMaxLifetime(30 * time.Minute)  // Maximum lifetime of a connection

        // Start a goroutine to perform connection health checks
        go monitorDBHealth()
        go reconnectDB(cfg)

        log.Println("PostgreSQL connection established successfully")
    })

    return db, nil
}

// GetDB retrieves the database connection pool
func GetDB() *sqlx.DB {
    return db
}

// monitorDBHealth Periodically check the health status of the database connection and attempt reconnection.
func monitorDBHealth() {
    for {
        // Check the health status of the connection pool every 30 seconds
        time.Sleep(30 * time.Second)
        if err := db.Ping(); err != nil {
            log.Printf("PostgreSQL connection lost: %v", err)
            reconnect <- true // Send a reconnection signal
        }
    }
}

// reconnectDB handles database reconnection.
func reconnectDB(cfg *config.Config) {
    for {
        <-reconnect
        log.Println("Attempting to reconnect to PostgreSQL...")
        var err error
        connStr := cfg.PostgreSQLURL

        db, err = sqlx.Connect("postgres", connStr)
        if err != nil {
            log.Printf("Failed to reconnect to PostgreSQL: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }

        db.SetMaxIdleConns(10)
        db.SetMaxOpenConns(100)
        db.SetConnMaxLifetime(30 * time.Minute)

        log.Println("Reconnected to PostgreSQL successfully")
    }
}
