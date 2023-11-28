package handlers

import (
	"gin_test/db"
	"gin_test/models"
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"time"
)

func HandleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}
		handleCommand(update.Message)
	}
}

func handleCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	strChatID := strconv.Itoa(int(chatID))
	state := db.Redis.Get(db.Ctx, "State:"+strChatID).Val()
	if state != "" {
		db.Redis.Del(db.Ctx, "State:"+strChatID)
	}
	switch message.Text {
	case "/start":
		handleStart(chatID)
	case "Поддержка":
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Опишите вашу проблему"))
		db.Redis.Set(db.Ctx, "State:"+strChatID, "Поддержка", time.Hour)
	case "Настройки":
		handleSettings(chatID, strChatID)
	case "Главное меню":
		utils.SendMessage(HandleMenu(chatID, "Главное меню"))
	default:
		handleMessageText(chatID, state, message.Text)
	}
}

func handleSettings(chatID int64, strChatID string) {
	msg := tgbotapi.NewMessage(chatID, "Настройки")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		utils.AddButtonsToRow("Изменить КД обновления", "Главное меню"),
	)
	utils.SendMessage(msg)
}

func handleMessageText(chatID int64, state, text string) {
	switch state {
	case "Поддержка":
		handleSupportText(chatID, text)
	default:
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Неизвестная команда"))
	}
}

func handleSupportText(chatID int64, text string) {
	newReq := models.Support{
		TelegramID: chatID,
		Message:    text,
	}
	db.Db.Create(&newReq)
	utils.SendMessage(tgbotapi.NewMessage(chatID, "Успешно отправлено"))
}

func handleStart(chatID int64) {
	msg := HandleMenu(chatID, "Добро пожаловать")
	utils.SendMessage(msg)
	var user models.User
	result := db.Db.First(&user, "telegram_id = ?", chatID)
	if result.Error != nil {
		newUser := models.User{
			TelegramID: chatID,
			RefreshKD:  30,
		}
		db.Db.Create(&newUser)
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
