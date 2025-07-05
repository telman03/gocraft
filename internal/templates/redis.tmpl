package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

// Config holds Redis configuration
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// Connect initializes Redis connection
func Connect() {
	config := getConfig()
	
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// Test connection
	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	log.Println("Connected to Redis successfully")
}

// getConfig returns Redis configuration from environment
func getConfig() Config {
	return Config{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
		PoolSize: 10,
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Set stores a key-value pair with optional expiration
func Set(key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return Client.Set(Ctx, key, jsonValue, expiration).Err()
}

// Get retrieves a value by key
func Get(key string, dest interface{}) error {
	val, err := Client.Get(Ctx, key).Result()
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key
func Delete(key string) error {
	return Client.Del(Ctx, key).Err()
}

// Exists checks if a key exists
func Exists(key string) (bool, error) {
	result, err := Client.Exists(Ctx, key).Result()
	return result > 0, err
}

// SetNX sets a key only if it doesn't exist
func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	
	return Client.SetNX(Ctx, key, jsonValue, expiration).Result()
}

// Incr increments a counter
func Incr(key string) (int64, error) {
	return Client.Incr(Ctx, key).Result()
}

// IncrBy increments a counter by a specific amount
func IncrBy(key string, value int64) (int64, error) {
	return Client.IncrBy(Ctx, key, value).Result()
}

// Expire sets expiration for a key
func Expire(key string, expiration time.Duration) error {
	return Client.Expire(Ctx, key, expiration).Err()
}

// TTL gets time to live for a key
func TTL(key string) (time.Duration, error) {
	return Client.TTL(Ctx, key).Result()
}

// Close closes Redis connection
func Close() error {
	return Client.Close()
} 