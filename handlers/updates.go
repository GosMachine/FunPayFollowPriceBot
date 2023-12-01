package handlers

import (
	"gin_test/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
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
	messageID := callback.Message.MessageID
	data := callback.Data
	var substring string
	db.Redis.Del(db.Ctx, "State:"+strChatID)
	if strings.HasPrefix(data, "lotSettings:") {
		substring = data[len("lotSettings:"):]
		data = "lotSettings"
	}
	switch data {
	case "Change KD":
		handleChangeKD(chatID, strChatID)
	case "Add a game":
		handleAddAGame(chatID, strChatID)
	case "lotSettings":
		handleLotSettingsCallBack(chatID, messageID, strChatID, substring)
	}
}

func handleCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	strChatID := strconv.Itoa(int(chatID))
	switch message.Text {
	case "/start":
		handleStart(chatID, strChatID)
	case "Поддержка":
		handleSupport(chatID, strChatID)
	case "Настройки":
		handleSettings(chatID, strChatID)
	case "Мои игры":
		handleMyGames(chatID, strChatID, 0)
	case "Запустить":
		handleStartRefreshBot(chatID, strChatID)
	case "Остановить":
		handleStopRefreshBot(chatID)
	default:
		handleMessageText(chatID, strChatID, message.Text)
	}
}
