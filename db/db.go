package db

import (
	"context"
	"fmt"
	"gin_test/logs"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var (
	Db    *gorm.DB
	Redis *redis.Client
	Ctx   = context.Background()
)

func init() {
	pgConnStr := fmt.Sprintf("user=%s password=%s dbname=%s host=127.0.0.1 sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	var err error
	Db, err = gorm.Open(postgres.Open(pgConnStr), &gorm.Config{})
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}

	Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
