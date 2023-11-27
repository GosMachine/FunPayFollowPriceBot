package utils

import (
	"gin_test/logs"
	"gin_test/tg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

func AddButtonsToRow(buttons ...string) []tgbotapi.KeyboardButton {
	var row []tgbotapi.KeyboardButton
	for _, btn := range buttons {
		row = append(row, tgbotapi.NewKeyboardButton(btn))
	}
	return row
}

func SendMessage(msg tgbotapi.MessageConfig) {
	_, err := tg.Bot.Send(msg)
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}
}
