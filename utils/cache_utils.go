package utils

import (
	"gin_test/db"
	"gin_test/logs"
	"gin_test/models"
	"go.uber.org/zap"
	"time"
)

func UserCache(chatID int64, strChatID string) *models.User {
	userData, err := db.Redis.Get(db.Ctx, "UserData:"+strChatID).Result()
	if err == nil && userData != "" {
		if err != nil {
			logs.Logger.Error("", zap.Error(err))
		} else {
			logs.Logger.Info("userData from cache")
			return DecodeUserData(userData)
		}
	}

	var user models.User
	db.Db.First(&user, "telegram_id = ?", chatID)
	if user.TelegramID != 0 {
		db.Redis.Set(db.Ctx, "UserData:"+strChatID, EncodeUserData(&user), time.Hour)
	}
	logs.Logger.Info("userData from postgres")
	return &user
}
