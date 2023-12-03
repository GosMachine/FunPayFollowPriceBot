package handlers

import (
	"gin_test/db"
	"gin_test/funpay"
	"gin_test/models"
	"gin_test/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"time"
)

var ticker *time.Ticker
var tickerRunning bool

func handleStopRefreshBot(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Успешно остановлено")
	if ticker != nil && tickerRunning {
		ticker.Stop()
		tickerRunning = false
	} else {
		msg.Text = "Error: Не запущено"
	}
	utils.SendMessage(msg)
}

func handleStartRefreshBot(chatID int64, strChatID string) {
	msg := tgbotapi.NewMessage(chatID, "Успешно запущено")
	if tickerRunning == true {
		msg.Text = "Error: Уже запущено"
	} else {
		user := utils.UserCache(chatID, strChatID)
		var activeLots []models.AllLots
		for _, lot := range user.AllLots {
			if lot.Active {
				activeLots = append(activeLots, lot)
			}
		}
		if len(activeLots) >= 1 {
			refreshKD := time.Duration(user.RefreshKD)
			if db.Redis.Get(db.Ctx, "RefreshBotKD:"+strChatID).Val() != "true" {
				sendRefreshBotMessage(chatID, activeLots)
				db.Redis.Set(db.Ctx, "RefreshBotKD:"+strChatID, "true", time.Minute*refreshKD)
			}
			tickerRunning = true
			ticker = time.NewTicker(refreshKD * time.Minute)
			go func(chatID int64, activeLots []models.AllLots, ticker *time.Ticker) {
				for range ticker.C {
					sendRefreshBotMessage(chatID, activeLots)
					db.Redis.Set(db.Ctx, "RefreshBotKD:"+strChatID, "true", time.Minute*refreshKD)
				}
			}(chatID, activeLots, ticker)
		} else {
			msg.Text = "У вас 0 активных игр"
		}
	}
	utils.SendMessage(msg)
}

func sendRefreshBotMessage(chatID int64, AllLots []models.AllLots) {
	for _, item := range AllLots {
		lots := funpay.Refresh(item.Lot, item.MaxPrice, item.Servers)
		var msgText string
		for _, v := range lots {
			strPrice := strconv.FormatFloat(v.Price, 'f', -1, 64)
			msgText += "Категория: " + v.Category + "\nПродавец: " + v.Seller + "\nЦена: " + strPrice + "₽"
			if v.Amount != "" {
				msgText += "\nНаличие: " + v.Amount
			}
			if v.Description != "" {
				msgText += "\nОписание: " + v.Description
			}
			if v.Server != "" {
				msgText += "\nСервер: " + v.Server
			}
			if v.Side != "" {
				msgText += "\nСторона: " + v.Side
			}
			msgText += "\n\n"
		}
		utils.SendMessage(tgbotapi.NewMessage(chatID, msgText))
	}
}
