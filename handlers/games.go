package handlers

import (
	"gin_test/db"
	"gin_test/models"
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func handleMyGames(chatID int64, strChatID string) {
	db.Redis.Del(db.Ctx, "State:"+strChatID)
	msg := tgbotapi.NewMessage(chatID, "Список отслеживаемых игр")
	var keyboard tgbotapi.InlineKeyboardMarkup
	user := utils.UserCache(chatID, strChatID)
	var rows []tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("Добавить игру", "Add a game"))
	for _, item := range user.AllLots {
		btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+"12")
		rows = append(rows, btn)
	}
	keyboard = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(rows...))
	msg.ReplyMarkup = keyboard
	utils.SendMessage(msg)
}

func handleAddAGame(chatID int64, strChatID string) {
	user := utils.UserCache(chatID, strChatID)
	msg := tgbotapi.NewMessage(chatID, "Максимум ты можешь отслеживать 10 игр")
	if len(user.AllLots) <= 10 {
		msg.Text = "Напиши ссылку на игру которую ты хочешь остлеживать. Пример: https://funpay.com/chips/192/"
		db.Redis.Set(db.Ctx, "State:"+strChatID, "Добавить игру", time.Hour)
	}
	utils.SendMessage(msg)
}

func handleAddAGameText(chatID int64, text, strChatID string) {
	re := regexp.MustCompile("^https://funpay.com/[a-z]+/[0-9]+/$")
	game := db.Redis.Get(db.Ctx, "game:"+strChatID).Val()
	servers := db.Redis.LRange(db.Ctx, "servers:"+strChatID, 0, -1).Val()
	maxPrice := db.Redis.Get(db.Ctx, "maxPrice:"+strChatID).Val()
	switch {
	case game == "":
		msg := tgbotapi.NewMessage(chatID, "Error: Попробуй снова. Пример: https://funpay.com/chips/192/")
		if re.MatchString(text) {
			db.Redis.Set(db.Ctx, "game:"+strChatID, text, time.Hour)
			msg.Text = "Теперь напиши сервера, которые ты хочешь отслеживать. Если их нет или ты хочешь отслеживать" +
				" все сервера, напиши \"None\") Пример: Гром, Галакронд, Пиратская Бухта"
		}
		utils.SendMessage(msg)
	case len(servers) == 0:
		chick := strings.Split(text, ", ")
		interfaceElements := make([]interface{}, len(chick))
		for i, v := range chick {
			interfaceElements[i] = v
		}
		db.Redis.RPush(db.Ctx, "servers:"+strChatID, interfaceElements...)
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Напиши максимальную стоимость, которую ты "+
			"хочешь отслеживать. Пример: 3.5"))
	case maxPrice == "":
		msg := tgbotapi.NewMessage(chatID, "Error: Введи число. Пример: 3.5")
		maxPriceFloat, err := strconv.ParseFloat(text, 64)
		if err == nil {
			msg.Text = "Игра успешно добавлена."
			utils.SendMessage(msg)
			if strings.ToLower(servers[0]) == "none" {
				servers = []string{}
			}
			user := utils.UserCache(chatID, strChatID)
			newLot := models.AllLots{
				UserID:   user.ID,
				Lot:      game,
				Servers:  servers,
				MaxPrice: maxPriceFloat,
			}
			go func(user *models.User, newLot models.AllLots, strChatID string) {
				db.Db.Create(&newLot)
				user.AllLots = append(user.AllLots, newLot)
				db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(user), time.Hour)
				db.Redis.Del(db.Ctx, "maxPrice:"+strChatID, "game:"+strChatID, "servers:"+strChatID, "State:"+strChatID)
			}(user, newLot, strChatID)
		}
	}
}

//TODO добавить имя, кружочки активности возле имени
//TODO чтоб выводились красиво а не в 1 линию(мб пагинацию)
