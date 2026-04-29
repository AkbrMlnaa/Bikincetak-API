package database

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0" 
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic("Gagal parsing REDIS_URL: " + err.Error())
	}

	Rdb = redis.NewClient(opt)

	// Test koneksi
	_, err = Rdb.Ping(Ctx).Result()
	if err != nil {
		panic("Gagal terhubung ke Redis: " + err.Error())
	}

	fmt.Println("Berhasil terhubung ke Redis")
}