// Package config provides configuration structures and functions
// for loading and managing application settings, such as database
// connections and environment-specific variables.
package config

import (
    "os"
)

// Config is used to load the configuration file.
type Config struct {
    PostgreSQLURL string
    RedisAddr     string
    RedisPassword string
}

// LoadConfig loads the PostgreSQL configuration.
func LoadConfig() *Config {
    return &Config{
        PostgreSQLURL: getEnv("POSTGRESQL_URL", "postgresql://user:mysecretpassword@localhost/wallet?sslmode=disable"),
    }
}

// LoadRedisConfig loads the Redis configuration.
func LoadRedisConfig() *Config {
    return &Config{
        RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPassword: getEnv("REDIS_PASSWORD", "mysecretpassword"),
    }
}

// LoadRedisConfig loads the Redis configuration.
func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}
