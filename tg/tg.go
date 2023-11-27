package tg

import (
	"gin_test/logs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

var (
	Bot *tgbotapi.BotAPI
)

func init() {
	var err error
	Bot, err = tgbotapi.NewBotAPI("6183666895:AAFMGw8k5IwQ3U6ZWgneX8USP7khfh34SYM")
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}
}
