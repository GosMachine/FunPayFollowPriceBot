package handlers

import (
	"fmt"
	"gin_test/db"
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"time"
)

func handleSettings(chatID int64, strChatID string) {
	db.Redis.Del(db.Ctx, "State:"+strChatID)
	user := utils.UserCache(chatID, strChatID)
	msg := tgbotapi.NewMessage(chatID, "Текущий KD: "+strconv.Itoa(user.RefreshKD)+" мин")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Изменить KD", "Change KD"),
		),
	)
	utils.SendMessage(msg)
}

func handleChangeKD(chatID int64, strChatID string) {
	user := utils.UserCache(chatID, strChatID)
	msg := tgbotapi.NewMessage(chatID, "Enter a number from 30 to 180 minutes (Buy premium to update every 5 minutes)")
	if user.Premium {
		msg = tgbotapi.NewMessage(chatID, "Enter a number from 5 to 180 minutes")
	}
	utils.SendMessage(msg)
	db.Redis.Set(db.Ctx, "State:"+strChatID, "Change KD", time.Hour)
}

func handleChangeKDText(chatID int64, text, strChatID string) {
	user := utils.UserCache(chatID, strChatID)
	minutes, err := strconv.Atoi(text)
	if err != nil {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Error: Please enter a number"))
		return
	}
	if user.Premium && (minutes < 5 || minutes > 180) {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Error: Please enter a number from 5 to 180"))
		return
	}

	if !user.Premium && (minutes < 30 || minutes > 180) {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Enter a number from 30 to 180 minutes (Buy premium to update every 5 minutes)"))
		return
	}
	utils.SendMessage(tgbotapi.NewMessage(chatID, fmt.Sprintf("Текущий KD: %s мин", text)))
	user.RefreshKD = minutes
	db.Db.Model(&user).Updates(user)
	db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(user), time.Hour)
	db.Redis.Del(db.Ctx, "State:"+strChatID)
}
