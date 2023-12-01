package handlers

import (
	"fmt"
	"gin_test/funpay"
	"gin_test/utils"
)

func handleStartRefreshBot(chatID int64, strChatID string) {
	user := utils.UserCache(chatID, strChatID)
	for _, item := range user.AllLots {
		fmt.Println(funpay.Refresh(item.Lot, item.MaxPrice, item.Servers))
	}
}
