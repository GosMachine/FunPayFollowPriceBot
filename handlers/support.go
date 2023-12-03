package handlers

import (
	"gin_test/db"
	"gin_test/models"
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

func handleSupportText(chatID int64, text, strChatID string) {
	utils.SendMessage(tgbotapi.NewMessage(chatID, "Успешно отправлено"))
	newReq := models.Support{
		TelegramID: chatID,
		Message:    text,
	}
	db.Db.Create(&newReq)
	db.Redis.Set(db.Ctx, "SupportKD:"+strChatID, "kd", time.Minute*30)
	db.Redis.Del(db.Ctx, "State:"+strChatID)
}

func handleSupport(chatID int64, strChatID string) {
	result, err := db.Redis.Get(db.Ctx, "SupportKD:"+strChatID).Result()
	msg := tgbotapi.NewMessage(chatID, "Опишите вашу проблему")
	if err == nil && result != "" {
		msg.Text = "В последние 30 минут вы уже отправляли сообщение в поддержку"
	} else {
		db.Redis.Set(db.Ctx, "State:"+strChatID, "Поддержка", time.Hour)
	}
	utils.SendMessage(msg)
}

//TODO сделать поддержку
