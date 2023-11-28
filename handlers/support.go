package handlers

import (
	"gin_test/db"
	"gin_test/models"
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

func handleSupportText(chatID int64, text, strChatID string) {
	newReq := models.Support{
		TelegramID: chatID,
		Message:    text,
	}
	db.Db.Create(&newReq)
	utils.SendMessage(tgbotapi.NewMessage(chatID, "Успешно отправлено"))
	db.Redis.Set(db.Ctx, "SupportKD:"+strChatID, "kd", time.Minute*30)
}

func handleSupport(chatID int64, strChatID string) {
	result, err := db.Redis.Get(db.Ctx, "SupportKD:"+strChatID).Result()
	if err == nil && result != "" {
		utils.SendMessage(tgbotapi.NewMessage(chatID,
			"В последние 30 минут вы уже отправляли сообщение в поддержку"))
	} else {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Опишите вашу проблему"))
		db.Redis.Set(db.Ctx, "State:"+strChatID, "Поддержка", time.Hour)
	}
}
