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
	user := utils.UserCache(chatID, strChatID)
	msg := tgbotapi.NewMessage(chatID, "Текущий KD: "+strconv.Itoa(user.RefreshKD)+" мин")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Изменить KD", "Change KD"),
		),
	)
	utils.SendMessage(msg)
	db.Redis.Del(db.Ctx, "State:"+strChatID)
}

func handleChangeKD(chatID int64, strChatID string) {
	user := utils.UserCache(chatID, strChatID)
	msg := tgbotapi.NewMessage(chatID, "Введите число от 30 до 180 минут (Купите премиум чтобы обновлять каждые 5 минут)")
	if user.Premium {
		msg.Text = "Введите число от 5 до 180 минут"
	}
	utils.SendMessage(msg)
	db.Redis.Set(db.Ctx, "State:"+strChatID, "Change KD", time.Hour)
}

func handleChangeKDText(chatID int64, text, strChatID string) {
	user := utils.UserCache(chatID, strChatID)
	minutes, err := strconv.Atoi(text)
	if err != nil {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Error: Введите число"))
		return
	}
	if user.Premium && (minutes < 5 || minutes > 180) {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Error: Введите число от 5 до 180"))
		return
	}

	if !user.Premium && (minutes < 30 || minutes > 180) {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Введите число от 30 до 180 (Купите премиум чтобы обновлять каждые 5 минут)"))
		return
	}
	utils.SendMessage(tgbotapi.NewMessage(chatID, fmt.Sprintf("Текущий KD: %s мин", text)))
	user.RefreshKD = minutes
	db.Db.Model(&user).Updates(user)
	db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(user), time.Hour)
	db.Redis.Del(db.Ctx, "State:"+strChatID)
}
