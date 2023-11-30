package main

import (
	"gin_test/db"
	"gin_test/handlers"
	"gin_test/logs"
	"gin_test/models"
	"gin_test/tg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	err := db.Db.AutoMigrate(&models.Lot{}, &models.User{}, &models.AllLots{}, &models.Support{})
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := tg.Bot.GetUpdatesChan(u)
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}
	handlers.HandleUpdates(updates)
}

//TODO сделать премиум
//TODO сделать admin panel
//TODO накидать смайликов(сделать красивого бота)
