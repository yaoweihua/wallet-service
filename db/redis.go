package db

import (
    "context"
    "github.com/go-redis/redis/v8"
    "github.com/yaoweihua/wallet-service/config"
    "log"
    "sync"
    "time"
)

var (
    redisClient  *redis.Client
    redisOnce    sync.Once // Ensure Redis connection is initialized only once
    reconnectSignal = make(chan bool, 1) // Used to handle reconnection notifications
)

// ConnectRedis initializes the Redis connection pool
func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
    redisOnce.Do(func() {
        options := &redis.Options{
            Addr:     cfg.RedisAddr,      // Redis server address
            Password: cfg.RedisPassword,  // Redis password
            DB:       0,
        }

        // Create a Redis client
        redisClient = redis.NewClient(options)

        // Test the Redis connection
        _, err := redisClient.Ping(context.Background()).Result()
        if err != nil {
            log.Fatalf("Redis connection failed: %v", err)
        }

        // Start a goroutine to monitor the health status of the Redis connection
        go monitorRedisHealth()
        go reconnectRedis()

        log.Println("Redis connection established successfully")
    })

    return redisClient, nil
}

// GetRedisClient retrieves the Redis client
func GetRedisClient() *redis.Client {
    return redisClient
}

// monitorRedisHealth periodically checks the health status of the Redis connection
func monitorRedisHealth() {
    for {
        // Check the health status of the connection every 30 seconds
        time.Sleep(30 * time.Second)
        if err := redisClient.Ping(context.Background()).Err(); err != nil {
            log.Printf("Redis connection lost: %v", err)
            reconnectSignal <- true // 发送重连信号
        }
    }
}

// reconnectRedis handles Redis reconnection
func reconnectRedis() {
    for {
        <-reconnectSignal
        log.Println("Attempting to reconnect to Redis...")
        var err error
        options := &redis.Options{
            Addr:     "your-redis-addr",
            Password: "your-redis-password",
            DB:       0,
        }

        // 尝试重新连接 Redis
        redisClient = redis.NewClient(options)
        _, err = redisClient.Ping(context.Background()).Result()
        if err != nil {
            log.Printf("Failed to reconnect to Redis: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }

        log.Println("Reconnected to Redis successfully")
    }
}
