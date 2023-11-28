package handlers

func handleMyGames(chatID int64, strChatID string) {
	//msg := tgbotapi.NewMessage(chatID, "Список отслеживаемых игр")
	//var keyboard tgbotapi.InlineKeyboardMarkup
	//if len(AllLots) >= 1 {
	//	var rows []tgbotapi.InlineKeyboardButton
	//	for _, item := range AllLots {
	//		btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Inactive")
	//		rows = append(rows, btn)
	//	}
	//	keyboard = tgbotapi.NewInlineKeyboardMarkup(
	//		tgbotapi.NewInlineKeyboardRow(rows...),
	//	)
	//	msg.ReplyMarkup = keyboard
	//	_, err = bot.Send(msg)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//} else {
	//	msge := tgbotapi.NewMessage(int64(telegramUserID), "You don't have any inactive games")
	//	bot.Send(msge)
	//}
	//sentMsg := utils.SendMessage(msg)
}
