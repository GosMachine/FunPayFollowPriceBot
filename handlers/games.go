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
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Добавить игру", "Add a game")))
	circle := "🔴"
	for _, item := range user.AllLots {
		if item.Active {
			circle = "🟢"
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(item.Name+circle, item.Name)
		rows = append(rows, btn)
		if len(rows)%2 == 0 {
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(rows...))
			rows = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(rows) > 0 {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(rows...))
	}
	msg.ReplyMarkup = keyboard
	utils.SendMessage(msg)
}

func handleAddAGame(chatID int64, strChatID string) {
	user := utils.UserCache(chatID, strChatID)
	msg := tgbotapi.NewMessage(chatID, "Максимум ты можешь отслеживать 10 игр")
	if len(user.AllLots) <= 10 {
		msg.Text = "Напиши название как эта игра будет отображатсья в списке отслеживаемых игр"
		db.Redis.Set(db.Ctx, "State:"+strChatID, "Добавить игру", time.Hour)
	}
	utils.SendMessage(msg)
}

func handleAddAGameText(chatID int64, text, strChatID string) {
	re := regexp.MustCompile("^https://funpay.com/[a-z]+/[0-9]+/$")
	game := db.Redis.Get(db.Ctx, "game:"+strChatID).Val()
	servers := db.Redis.LRange(db.Ctx, "servers:"+strChatID, 0, -1).Val()
	maxPrice := db.Redis.Get(db.Ctx, "maxPrice:"+strChatID).Val()
	name := db.Redis.Get(db.Ctx, "name:"+strChatID).Val()
	switch {
	case name == "":
		msg := tgbotapi.NewMessage(chatID, "Error: Введи название от 3 до 9 символов")
		if len(text) <= 9 && len(text) >= 3 {
			db.Redis.Set(db.Ctx, "name:"+strChatID, text, time.Hour)
			msg.Text = "Напиши ссылку на игру которую ты хочешь остлеживать. Пример: https://funpay.com/chips/192/"
		}
		utils.SendMessage(msg)
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
				Name:     name,
				Servers:  servers,
				MaxPrice: maxPriceFloat,
			}
			go func(user *models.User, newLot models.AllLots, strChatID string) {
				db.Db.Create(&newLot)
				user.AllLots = append(user.AllLots, newLot)
				db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(user), time.Hour)
				db.Redis.Del(db.Ctx, "maxPrice:"+strChatID, "game:"+strChatID, "servers:"+strChatID,
					"State:"+strChatID, "name:"+strChatID)
			}(user, newLot, strChatID)
		}
	}
}
