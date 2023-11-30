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

func sendMessage(config tgbotapi.Chattable) tgbotapi.Message {
	sentMsg, err := tg.Bot.Send(config)
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}
	return sentMsg
}

func SendMessage(msg tgbotapi.MessageConfig) tgbotapi.Message {
	return sendMessage(msg)
}

func SendEditMessage(msg tgbotapi.EditMessageTextConfig) tgbotapi.Message {
	return sendMessage(msg)
}

func EditMessage(chatID int64, messageID int, newText string, newButtons tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, newText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = &newButtons
	_, err := tg.Bot.Send(msg)
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
	}
}

func GetState(strChatID string) string {
	state, _ := db.Redis.Get(db.Ctx, "State:"+strChatID).Result()
	return state
}
