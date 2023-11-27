package db

import (
	"gin_test/logs"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	Db    *gorm.DB
	Redis *redis.Client
)

func init() {
	pgConnStr := "user=postgres password=12345tgv dbname=FunPayFollowBot host=127.0.0.1 sslmode=disable"
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
