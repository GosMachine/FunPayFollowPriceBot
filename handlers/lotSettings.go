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
		setRedisData(strChatID, dataList[1], "change lot servers", messageID)
	case "Change name":
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Напиши название как эта игра будет отображатсья в списке отслеживаемых игр"))
		setRedisData(strChatID, dataList[1], "change lot name", messageID)
	case "Change maxPrice":
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Напиши максимальную стоимость, которую ты "+
			"хочешь отслеживать. Пример: 3.5"))
		setRedisData(strChatID, dataList[1], "change lot maxPrice", messageID)
	case "Change active":
		handleLotSettingsActive(chatID, strChatID, dataList[1], messageID, user)
	case "Change lot":
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Напиши ссылку на игру которую ты хочешь остлеживать."+
			" Пример: https://funpay.com/chips/192/"))
		setRedisData(strChatID, dataList[1], "change lot link", messageID)
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

func updateLotData(chatID int64, strChatID, cache string) {
	user := utils.UserCache(chatID, strChatID)
	data := db.Redis.Get(db.Ctx, "LotID:"+strChatID).Val()
	messageID, _ := strconv.Atoi(db.Redis.Get(db.Ctx, "MessageID:"+strChatID).Val())
	item := GetItem(user, data)
	switch cache {
	case "servers:":
		item.Servers = db.Redis.LRange(db.Ctx, "servers:"+strChatID, 0, -1).Val()
		db.Db.Model(&item).Update("Servers", item.Servers)
	case "name:":
		item.Name = db.Redis.Get(db.Ctx, "name:"+strChatID).Val()
		db.Db.Model(&item).Update("Name", item.Name)
	case "maxPrice:":
		item.MaxPrice, _ = db.Redis.Get(db.Ctx, "maxPrice:"+strChatID).Float64()
		db.Db.Model(&item).Update("max_price", item.MaxPrice)
	case "game:":
		item.Lot = db.Redis.Get(db.Ctx, "game:"+strChatID).Val()
		db.Db.Model(&item).Update("lot", item.Lot)
	}
	db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(user), time.Hour)
	handleLotSettings(chatID, messageID, user, data)
	db.Redis.Del(db.Ctx, cache+strChatID, "LotID:"+strChatID, "State:"+strChatID, "MessageID:"+strChatID)
}

func handleLotSettingsName(chatID int64, text, strChatID string) {
	if utils.LotName(text, strChatID) == nil {
		updateLotData(chatID, strChatID, "name:")
	} else {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Error: Введи название от 3 до 9 символов"))
	}
}

func handleLotSettingsLink(chatID int64, text, strChatID string) {
	if utils.LotGame(text, strChatID) == nil {
		updateLotData(chatID, strChatID, "game:")
	} else {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Error: Попробуй снова. Пример: https://funpay.com/chips/192/"))
	}
}

func handleLotSettingsMaxPrice(chatID int64, text, strChatID string) {
	if _, err := utils.LotMaxPrice(text, strChatID); err == nil {
		updateLotData(chatID, strChatID, "maxPrice:")
	} else {
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Error: Введи число. Пример: 3.5"))
	}
}

func handleLotSettingsServers(chatID int64, text, strChatID string) {
	utils.LotServers(text, strChatID)
	updateLotData(chatID, strChatID, "servers:")
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

func setRedisData(strChatID, data, value string, messageID int) {
	db.Redis.Set(db.Ctx, "State:"+strChatID, value, time.Hour)
	db.Redis.Set(db.Ctx, "LotID:"+strChatID, data, time.Hour)
	db.Redis.Set(db.Ctx, "MessageID:"+strChatID, messageID, time.Hour)
}
