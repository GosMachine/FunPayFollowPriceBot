package tg

import (
	"gin_test/logs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
	"os"
)

var (
	Bot *tgbotapi.BotAPI
)

func init() {
	var err error
	Bot, err = tgbotapi.NewBotAPI(os.Getenv("TG_TOKEN"))
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}
}
