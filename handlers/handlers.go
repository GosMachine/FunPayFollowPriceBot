package handlers

import (
	"gin_test/db"
	"gin_test/models"
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

func handleMessageText(chatID int64, strChatID, text string) {
	state := utils.GetState(strChatID)
	switch state {
	case "Поддержка":
		handleSupportText(chatID, text, strChatID)
	case "Change KD":
		handleChangeKDText(chatID, text, strChatID)
	case "Добавить игру":
		handleAddAGameText(chatID, text, strChatID)
	case "change lot name":
		handleLotSettingsName(chatID, text, strChatID)
	case "change lot link":
		handleLotSettingsLink(chatID, text, strChatID)
	case "change lot servers":
		handleLotSettingsServers(chatID, text, strChatID)
	case "change lot maxPrice":
		handleLotSettingsMaxPrice(chatID, text, strChatID)
	default:
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Неизвестная команда"))
	}
}

func handleStart(chatID int64, strChatID string) {
	user := utils.UserCache(chatID, strChatID)
	utils.SendMessage(HandleMenu(chatID, "Добро пожаловать"))
	if user.TelegramID == 0 {
		newUser := models.User{
			TelegramID: chatID,
			RefreshKD:  30,
		}
		db.Db.Create(&newUser)
		db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(&newUser), time.Hour)
		db.Redis.Del(db.Ctx, "State:"+strChatID)
	}
}

func HandleMenu(chatID int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		utils.AddButtonsToRow("Поддержка", "Мои игры"),
		utils.AddButtonsToRow("Настройки", "Премиум"),
		utils.AddButtonsToRow("Запустить", "Остановить"),
	)
	return msg
}
