package handlers

import (
	"gin_test/db"
	"gin_test/models"
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func HandleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}
		handleMessage(update.Message)
	}
}

func handleMessage(message *tgbotapi.Message) {
	ChatID := message.Chat.ID
	switch message.Command() {
	case "start":
		msg := HandleMenu(ChatID, "Добро пожаловать")
		utils.SendMessage(msg)
		var user models.User
		result := db.Db.First(&user, "telegram_id = ?", ChatID)
		if result.Error != nil {
			newUser := models.User{
				TelegramID: ChatID,
				Lang:       "Russian",
				RefreshKD:  30,
			}
			db.Db.Create(&newUser)
		}
	default:
		utils.SendMessage(tgbotapi.NewMessage(ChatID, "Неизвестная команда"))
	}
}

func HandleMenu(ChatID int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(ChatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		utils.AddButtonsToRow("Support", "My Games"),
		utils.AddButtonsToRow("Settings", "Premium"),
		utils.AddButtonsToRow("/go", "/stop"),
	)
	return msg
}
