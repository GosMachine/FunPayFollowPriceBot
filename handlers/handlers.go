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

	case "change lot link":

	case "change lot servers":
		handleLotSettingsServers(chatID, text, strChatID)
	case "change lot maxPrice":

	default:
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Неизвестная команда"))
	}
}

func handleStart(chatID int64, strChatID string) {
	db.Redis.Del(db.Ctx, "State:"+strChatID)
	user := utils.UserCache(chatID, strChatID)
	utils.SendMessage(HandleMenu(chatID, "Добро пожаловать"))
	if user.TelegramID == 0 {
		newUser := models.User{
			TelegramID: chatID,
			RefreshKD:  30,
		}
		db.Db.Create(&newUser)
		db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(&newUser), time.Hour)
	}
}

func HandleMenu(chatID int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		utils.AddButtonsToRow("Поддержка", "Мои игры"),
		utils.AddButtonsToRow("Настройки", "Премиум"),
		utils.AddButtonsToRow("Старт", "Стоп"),
	)
	return msg
}

//TODO кэширование
//TODO ускорение горутинами
//TODO логирование
//TODO сделать env file
//TODO jaeger затестить
