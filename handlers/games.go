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

func handleMyGames(chatID int64, strChatID string, messageID int) {
	db.Redis.Del(db.Ctx, "State:"+strChatID)
	text := "Список отслеживаемых игр"
	msg := tgbotapi.NewMessage(chatID, text)
	msgE := tgbotapi.NewEditMessageText(chatID, messageID, text)
	var keyboard tgbotapi.InlineKeyboardMarkup
	user := utils.UserCache(chatID, strChatID)
	var rows []tgbotapi.InlineKeyboardButton
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Добавить игру", "Add a game")))
	circle := ": 🔴"
	for _, item := range user.AllLots {
		if item.Active {
			circle = ": 🟢"
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(item.Name+circle, "lotSettings:lotSettings:"+strconv.Itoa(int(item.ID)))
		rows = append(rows, btn)
		if len(rows)%2 == 0 {
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(rows...))
			rows = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(rows) > 0 {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(rows...))
	}
	if messageID != 0 {
		msgE.ReplyMarkup = &keyboard
		utils.SendEditMessage(msgE)
	} else {
		msg.ReplyMarkup = keyboard
		utils.SendMessage(msg)
	}
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
	game := db.Redis.Get(db.Ctx, "game:"+strChatID).Val()
	servers := db.Redis.LRange(db.Ctx, "servers:"+strChatID, 0, -1).Val()
	maxPrice := db.Redis.Get(db.Ctx, "maxPrice:"+strChatID).Val()
	name := db.Redis.Get(db.Ctx, "name:"+strChatID).Val()
	switch {
	case name == "":
		msg := tgbotapi.NewMessage(chatID, "Error: Введи название от 3 до 9 символов")
		if utils.LotName(text, strChatID) == nil {
			msg.Text = "Напиши ссылку на игру которую ты хочешь остлеживать. Пример: https://funpay.com/chips/192/"
		}
		utils.SendMessage(msg)
	case game == "":
		msg := tgbotapi.NewMessage(chatID, "Error: Попробуй снова. Пример: https://funpay.com/chips/192/")
		if utils.LotGame(text, strChatID) == nil {
			msg.Text = "Теперь напиши сервера, которые ты хочешь отслеживать. Если их нет или ты хочешь отслеживать" +
				" все сервера, напиши \"None\") Пример: Гром, Галакронд, Пиратская Бухта"
		}
		utils.SendMessage(msg)
	case len(servers) == 0:
		utils.LotServers(text, strChatID)
		utils.SendMessage(tgbotapi.NewMessage(chatID, "Напиши максимальную стоимость, которую ты "+
			"хочешь отслеживать. Пример: 3.5"))
	case maxPrice == "":
		msg := tgbotapi.NewMessage(chatID, "Error: Введи число. Пример: 3.5")
		if maxPriceFloat, err := utils.LotMaxPrice(text, strChatID); err == nil {
			msg.Text = "Игра успешно добавлена."
			go func(strChatID string) {
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
				db.Db.Create(&newLot)
				user.AllLots = append(user.AllLots, newLot)
				db.Redis.Set(db.Ctx, "UserData:"+strChatID, utils.EncodeUserData(user), time.Hour)
				db.Redis.Del(db.Ctx, "maxPrice:"+strChatID, "game:"+strChatID, "servers:"+strChatID,
					"State:"+strChatID, "name:"+strChatID)
			}(strChatID)
		}
		utils.SendMessage(msg)
	}
}
