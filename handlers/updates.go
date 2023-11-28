package handlers

import (
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
)

func HandleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.CallbackQuery != nil {
			handleCallbackQuery(update)
		} else if update.Message != nil {
			handleCommand(update.Message)
		}
	}
}

func handleCallbackQuery(update tgbotapi.Update) {
	callback := update.CallbackQuery
	chatID := callback.Message.Chat.ID
	strChatID := strconv.Itoa(int(chatID))
	//messageID := callback.Message.MessageID
	data := callback.Data
	switch data {
	case "Change KD":
		handleChangeKD(chatID, strChatID)
	}
}

func handleCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	strChatID := strconv.Itoa(int(chatID))
	state := utils.GetState(strChatID)
	switch message.Text {
	case "/start":
		handleStart(chatID, strChatID)
	case "Поддержка":
		handleSupport(chatID, strChatID)
	case "Настройки":
		handleSettings(chatID, strChatID)
	default:
		handleMessageText(chatID, state, message.Text, strChatID)
	}
}
