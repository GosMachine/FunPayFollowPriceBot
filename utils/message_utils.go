package utils

import (
	"gin_test/db"
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

func SendMessage(msg tgbotapi.MessageConfig) tgbotapi.Message {
	sentMsg, err := tg.Bot.Send(msg)
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}
	return sentMsg
}

func EditInlineReplyMarkup(chatID int64, messageID int, replyMarkup tgbotapi.InlineKeyboardMarkup) {
	editMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, replyMarkup)

	_, err := tg.Bot.Send(editMsg)
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}
}

func GetState(strChatID string) string {
	state, err := db.Redis.Get(db.Ctx, "State:"+strChatID).Result()
	if err == nil && state != "" {
		db.Redis.Del(db.Ctx, "State:"+strChatID)
	}
	return state
}
