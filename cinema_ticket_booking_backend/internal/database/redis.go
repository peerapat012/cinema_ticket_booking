package database

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Println("Warning: REDIS_URL is not set, Redis features disabled")
		return
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Println("Warning: Error parsing REDIS_URL:", err)
		return
	}

	client := redis.NewClient(opt)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Println("Warning: Error connecting to Redis, Redis features disabled:", err)
		return
	}

	Redis = client
	log.Println("Connected to Redis")
}

func GetRedis() *redis.Client {
	return Redis
}
