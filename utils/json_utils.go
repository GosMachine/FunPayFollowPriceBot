package utils

import (
	"encoding/json"
	"gin_test/logs"
	"gin_test/models"
	"go.uber.org/zap"
)

func EncodeUserData(user *models.User) string {
	data, err := json.Marshal(user)
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
		return ""
	}
	return string(data)
}

func DecodeUserData(data string) *models.User {
	var user models.User
	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		logs.Logger.Error("", zap.Error(err))
		return nil
	}
	return &user
}
