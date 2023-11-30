package handlers

import (
	"gin_test/db"
	"gin_test/models"
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
	"time"
)

func handleLotSettings(chatID int64, messageID int, user *models.User, data string) {
	item, _ := utils.FindAllLotsItem(user, data)
	newReply := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Включить/Выключить", "lotSettings:Change active:"+data),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Изменить Имя", "lotSettings:Change name:"+data),
			tgbotapi.NewInlineKeyboardButtonData("Изменить сервера", "lotSettings:Change servers:"+data),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Изменить Макс.Цену", "lotSettings:Change maxPrice:"+data),
			tgbotapi.NewInlineKeyboardButtonData("Изменить лот", "lotSettings:Change lot:"+data),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить", "lotSettings:Delete:"+data),
			tgbotapi.NewInlineKeyboardButtonData("Назад", "lotSettings:Back:"+data),
		),
	)
	circle := ": 🔴"
	if item.Active {
		circle = ": 🟢"
	}
	maxPriceString := strconv.FormatFloat(item.MaxPrice, 'f', -1, 64)
	msgText := item.Name + circle + "\n" + "Сервера: " + strings.Join(item.Servers, ",") +
		"\nМакс.Цена: " + maxPriceString + "\nСсылка: " + item.Lot
	utils.EditMessage(chatID, messageID, msgText, newReply)
}

func handleLotSettingsCallBack(chatID int64, messageID int, strChatID, data string) {
	user := utils.UserCache(chatID, strChatID)
	dataList := strings.Split(data, ":")
	switch dataList[0] {
	case "Delete":
		handleLotSettingsDelete(chatID, strChatID, dataList[1], messageID, user)
	case "Back":
		handleMyGames(chatID, strChatID, messageID)
	case "Change servers":
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Теперь напиши сервера, которые ты хочешь отслеживать."+
			" Если их нет или ты хочешь отслеживать все сервера, напиши \"None\") Пример: Гром, Галакронд, Пиратская Бухта"))
		db.Redis.Set(db.Ctx, "State:"+strChatID, "change lot servers", time.Hour)
		db.Redis.Set(db.Ctx, "LotID:"+strChatID, dataList[1], time.Hour)
		db.Redis.Set(db.Ctx, "MessageID:"+strChatID, messageID, time.Hour)
	case "Change name":
		utils.SendMessage(tgbotapi.NewMessage(chatID, "изменил имя"))
	case "Change maxPrice":
		utils.SendMessage(tgbotapi.NewMessage(chatID, "изменил цену"))
	case "Change active":
		handleLotSettingsActive(chatID, strChatID, dataList[1], messageID, user)
	case "Change lot":
		utils.SendMessage(tgbotapi.NewMessage(chatID, "изменил лот"))
	case "lotSettings":
		handleLotSettings(chatID, messageID, user, dataList[1])
	}
}

func handleLotSettingsDelete(chatID int64, strChatID, data string, messageID int, user *models.User) {
	_, indexToChange := utils.FindAllLotsItem(user, data)
	db.Db.Delete(&user.AllLots[indexToChange])
	user.AllLots = append(user.AllLots[:indexToChange], user.AllLots[indexToChange+1:]...)
	db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(user), time.Hour)
	handleMyGames(chatID, strChatID, messageID)
}

func handleLotSettingsServers(chatID int64, text, strChatID string) {
	utils.LotServers(text, strChatID)
	user := utils.UserCache(chatID, strChatID)
	data := db.Redis.Get(db.Ctx, "LotID:"+strChatID).Val()
	messageID, _ := strconv.Atoi(db.Redis.Get(db.Ctx, "MessageID:"+strChatID).Val())
	item := GetItem(user, data)
	item.Servers = db.Redis.LRange(db.Ctx, "servers:"+strChatID, 0, -1).Val()
	db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(user), time.Hour)
	handleLotSettings(chatID, messageID, user, data)
	db.Db.Model(&item).Update("Servers", item.Servers)
	db.Redis.Del(db.Ctx, "servers:"+strChatID, "LotID:"+strChatID, "State:"+strChatID, "MessageID:"+strChatID)
}

func handleLotSettingsActive(chatID int64, strChatID, data string, messageID int, user *models.User) {
	item := GetItem(user, data)
	item.Active = !item.Active
	db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(user), time.Hour)
	handleLotSettings(chatID, messageID, user, data)
	db.Db.Model(&item).Update("Active", item.Active)
}

func GetItem(user *models.User, data string) *models.AllLots {
	_, indexToChange := utils.FindAllLotsItem(user, data)
	return &user.AllLots[indexToChange]
}
