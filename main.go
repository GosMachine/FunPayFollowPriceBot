package main

import (
	"fmt"
	"gin_test/models"
	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func refresh(database *gorm.DB, en bool, lot string, maxPrice float64, servers []string) []models.Lot {
	url := "https://funpay.com/" + lot
	var response *http.Response
	if en {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Accept-Language", "en")
		client := &http.Client{}
		response, _ = client.Do(req)
	} else {
		response, _ = http.Get(url)
	}
	defer response.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(response.Body)
	var res []models.Lot
	doc.Find("a.tc-item").Each(func(i int, c *goquery.Selection) {
		pricee := strings.TrimSpace(strings.Join(strings.Split(c.Find("div.tc-price").Text(), " "), ""))
		price, _ := strconv.ParseFloat(pricee[:len(pricee)-3], 64)
		if price <= maxPrice {
			var newLot models.Lot
			contentDiv := doc.Find(".content-with-cd")
			newLot.Category = contentDiv.Find("h1").Text()
			header := doc.Find(".tc-header")
			header.Find("*").Each(func(i int, s *goquery.Selection) {
				class := strings.Split(s.AttrOr("class", ""), " ")
				text := strings.TrimSpace(s.Text())
				if class[0] != "" && text != "" {
					if class[0] == "tc-user" {
						value := strings.TrimSpace(c.Find("div.media-user-name").Text())
						newLot.Seller = value
					} else if class[0] == "tc-server" {
						value := strings.TrimSpace(c.Find("div.tc-server").Text())
						newLot.Server = value
					} else if class[0] == "tc-price" {
						newLot.Price = price
					} else if class[0] == "tc-amount" {
						value := strings.TrimSpace(c.Find("div.tc-amount").Text())
						newLot.Amount = value
					} else if class[0] == "tc-desc" {
						value := strings.TrimSpace(c.Find("div.tc-desc-text").Text())
						newLot.Description = value
					} else if class[0] == "tc-side" {
						value := strings.TrimSpace(c.Find("div.tc-side").Text())
						newLot.Side = value
					}
				}
			})
			if in(servers, newLot.Server) || newLot.Server == "" || newLot.Server == "Любой" || newLot.Server == "Any" || len(servers) == 0 {
				var existingLot models.Lot
				if err := database.Where(&newLot).First(&existingLot).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						database.Create(&newLot)
						res = append(res, newLot)
					}
				}
			}
			newLot = models.Lot{}
		}
	})
	return res
}

func in(ss []string, s string) bool {
	for i := 0; i < len(ss); i++ {
		if ss[i] == s {
			return true
		}
	}
	return false
}

var ticker *time.Ticker
var tickerRunning bool

func main() {
	var pred string
	connection := "user=postgres password=12345tgv dbname=FunPayTgBot host=127.0.0.1 sslmode=disable"
	database, err := gorm.Open(postgres.Open(connection), &gorm.Config{})
	if err != nil {
		panic("db connection failed")
	}
	database.AutoMigrate(&models.Lot{}, &models.User{}, &models.AllLots{}, &models.ActiveLots{}, &models.Support{})
	bot, err := tgbotapi.NewBotAPI("6183666895:AAFMGw8k5IwQ3U6ZWgneX8USP7khfh34SYM")
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	var telegramUserID int
	var user models.User
	re := regexp.MustCompile("^[a-z]+/[0-9]+/$")
	var game string
	var servers []string
	var maxPrice float64
	nowCallBack := ""
	nowLot := ""
	for update := range updates {
		if update.CallbackQuery != nil {
			var ActiveLots []models.ActiveLots
			database.Where("user_id = ?", user.ID).Find(&ActiveLots)
			var AllLots []models.AllLots
			database.Where("user_id = ?", user.ID).Find(&AllLots)
			callbackData := strings.Split(update.CallbackQuery.Data, " ")
			if len(callbackData) > 1 {
				nowCallBack = callbackData[1]
			}
			log.Println(callbackData)

			if re.MatchString(callbackData[0]) {
				nowLot = callbackData[0]
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Change game", "Change game "+nowCallBack),
						tgbotapi.NewInlineKeyboardButtonData("Change servers", "Change servers "+nowCallBack),
						tgbotapi.NewInlineKeyboardButtonData("Change Max Price", "Change Max Price "+nowCallBack),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Activate/Deactivate", "Activate/Deactivate "+nowCallBack),
						tgbotapi.NewInlineKeyboardButtonData("Delete", "Delete "+nowCallBack),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Back", "Back "+nowCallBack),
					),
				)
				editMsgText := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, callbackData[0])
				bot.Send(editMsgText)
				editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, keyboard)
				bot.Send(editMsg)
			} else {
				switch callbackData[0] {
				case "Back":
					if nowCallBack == "Active" {
						msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "")
						var rowsActive []tgbotapi.InlineKeyboardButton
						for _, item := range ActiveLots {
							btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Active")
							rowsActive = append(rowsActive, btn)
						}
						msg.Text = strconv.Itoa(len(ActiveLots)) + " Active games"
						keyboard := tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(rowsActive...),
						)
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
						editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, keyboard)
						bot.Send(editMsg)
					} else if nowCallBack == "Inactive" {
						var rows []tgbotapi.InlineKeyboardButton
						editMsgText := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "You have "+strconv.Itoa(len(AllLots))+" Inactive games")
						for _, item := range AllLots {
							btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Inactive")
							rows = append(rows, btn)
						}
						keyboard := tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(rows...),
						)
						bot.Send(editMsgText)
						editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, keyboard)
						bot.Send(editMsg)
					}
				case "Delete":
					if nowCallBack == "Active" {
						var al models.ActiveLots
						database.Where("lot = ? AND user_id = ?", nowLot, user.ID).First(&al)
						if al.Lot != "" {
							database.Delete(&al)
							msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "")
							var rowsActive []tgbotapi.InlineKeyboardButton
							database.Where("user_id = ?", user.ID).Find(&ActiveLots)
							for _, item := range ActiveLots {
								btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Active")
								rowsActive = append(rowsActive, btn)
							}
							msg.Text = strconv.Itoa(len(ActiveLots)) + " Active games"
							keyboard := tgbotapi.NewInlineKeyboardMarkup(
								tgbotapi.NewInlineKeyboardRow(rowsActive...),
							)
							_, err = bot.Send(msg)
							if err != nil {
								fmt.Println(err)
							}
							editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, keyboard)
							bot.Send(editMsg)
						}
						nowLot = ""
					} else if nowCallBack == "Inactive" {
						var al models.AllLots
						database.Where("lot = ? AND user_id = ?", nowLot, user.ID).First(&al)
						if al.Lot != "" {
							database.Delete(&al)
							var rows []tgbotapi.InlineKeyboardButton
							database.Where("user_id = ?", user.ID).Find(&AllLots)
							editMsgText := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "You have "+strconv.Itoa(len(AllLots))+" Inactive games")
							for _, item := range AllLots {
								btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Inactive")
								rows = append(rows, btn)
							}
							keyboard := tgbotapi.NewInlineKeyboardMarkup(
								tgbotapi.NewInlineKeyboardRow(rows...),
							)
							bot.Send(editMsgText)
							editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, keyboard)
							bot.Send(editMsg)
						}
						nowLot = ""
					}
				case "Activate/Deactivate":
					if nowCallBack == "Active" {
						var al models.ActiveLots
						database.Where("lot = ? AND user_id = ?", nowLot, user.ID).First(&al)
						newRecord := models.AllLots{
							ID:       al.ID,
							MaxPrice: al.MaxPrice,
							UserID:   al.UserID,
							Lot:      al.Lot,
							Servers:  al.Servers,
						}
						database.Create(&newRecord)
						database.Delete(&al)
						nowLot = ""
						database.Where("user_id = ?", user.ID).Find(&AllLots)
						database.Where("user_id = ?", user.ID).Find(&ActiveLots)
						msg := tgbotapi.NewMessage(int64(telegramUserID), "You have "+strconv.Itoa(len(AllLots))+" Inactive games")
						var keyboard tgbotapi.InlineKeyboardMarkup
						if len(AllLots) >= 1 {
							var rows []tgbotapi.InlineKeyboardButton
							for _, item := range AllLots {
								btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Inactive")
								rows = append(rows, btn)
							}
							keyboard = tgbotapi.NewInlineKeyboardMarkup(
								tgbotapi.NewInlineKeyboardRow(rows...),
							)
							msg.ReplyMarkup = keyboard
							_, err = bot.Send(msg)
							if err != nil {
								fmt.Println(err)
							}
						} else {
							msge := tgbotapi.NewMessage(int64(telegramUserID), "You don't have any inactive games")
							bot.Send(msge)
						}
						if len(ActiveLots) >= 1 {
							var rowsActive []tgbotapi.InlineKeyboardButton
							for _, item := range ActiveLots {
								btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Active")
								rowsActive = append(rowsActive, btn)
							}
							msg.Text = strconv.Itoa(len(ActiveLots)) + " Active games"
							keyboard = tgbotapi.NewInlineKeyboardMarkup(
								tgbotapi.NewInlineKeyboardRow(rowsActive...),
							)
							msg.ReplyMarkup = keyboard
							_, err = bot.Send(msg)
							if err != nil {
								fmt.Println(err)
							}
						} else {
							msge := tgbotapi.NewMessage(int64(telegramUserID), "You don't have any active games")
							bot.Send(msge)
						}
					} else if nowCallBack == "Inactive" {
						var al models.AllLots
						database.Where("lot = ? AND user_id = ?", nowLot, user.ID).First(&al)
						newRecord := models.ActiveLots{
							ID:       al.ID,
							MaxPrice: al.MaxPrice,
							UserID:   al.UserID,
							Lot:      al.Lot,
							Servers:  al.Servers,
						}
						database.Create(&newRecord)
						database.Delete(&al)
						nowLot = ""
						database.Where("user_id = ?", user.ID).Find(&AllLots)
						database.Where("user_id = ?", user.ID).Find(&ActiveLots)
						msg := tgbotapi.NewMessage(int64(telegramUserID), "You have "+strconv.Itoa(len(AllLots))+" Inactive games")
						var keyboard tgbotapi.InlineKeyboardMarkup
						if len(AllLots) >= 1 {
							var rows []tgbotapi.InlineKeyboardButton
							for _, item := range AllLots {
								btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Inactive")
								rows = append(rows, btn)
							}
							keyboard = tgbotapi.NewInlineKeyboardMarkup(
								tgbotapi.NewInlineKeyboardRow(rows...),
							)
							msg.ReplyMarkup = keyboard
							_, err = bot.Send(msg)
							if err != nil {
								fmt.Println(err)
							}
						} else {
							msge := tgbotapi.NewMessage(int64(telegramUserID), "You don't have any inactive games")
							bot.Send(msge)
						}
						if len(ActiveLots) >= 1 {
							var rowsActive []tgbotapi.InlineKeyboardButton
							for _, item := range ActiveLots {
								btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Active")
								rowsActive = append(rowsActive, btn)
							}
							msg.Text = strconv.Itoa(len(ActiveLots)) + " Active games"
							keyboard = tgbotapi.NewInlineKeyboardMarkup(
								tgbotapi.NewInlineKeyboardRow(rowsActive...),
							)
							msg.ReplyMarkup = keyboard
							_, err = bot.Send(msg)
							if err != nil {
								fmt.Println(err)
							}
						} else {
							msge := tgbotapi.NewMessage(int64(telegramUserID), "You don't have any active games")
							bot.Send(msge)
						}
					}
				}
			}
		}
		if update.Message != nil {
			var ActiveLots []models.ActiveLots
			database.Where("user_id = ?", user.ID).Find(&ActiveLots)
			var AllLots []models.AllLots
			database.Where("user_id = ?", user.ID).Find(&AllLots)
			if telegramUserID == 0 {
				telegramUserID = update.Message.From.ID
				result := database.First(&user, "telegram_id = ?", telegramUserID)
				if result.Error != nil {
					newUser := models.User{
						TelegramID: telegramUserID,
						Lang:       "Russian",
						RefreshKD:  60,
					}
					database.Create(&newUser)
				}
			}
			if update.Message.Text == "/stop" && tickerRunning {
				if ticker != nil {
					ticker.Stop()
					tickerRunning = false
					bot.Send(tgbotapi.NewMessage(int64(telegramUserID), "Successfully stopped"))
				}
			} else if update.Message.Text == "/go" && !tickerRunning {
				if len(ActiveLots) >= 1 {
					bot.Send(tgbotapi.NewMessage(int64(telegramUserID), "Successfully started"))
					tickerRunning = true
					refreshKD := time.Duration(user.RefreshKD)
					ticker = time.NewTicker(refreshKD * time.Minute)
					en := true
					if user.Lang == "Russian" {
						en = false
					}
					go func() {
						for range ticker.C {
							for _, value := range ActiveLots {
								lots := refresh(database, en, value.Lot, value.MaxPrice, value.Servers)
								var msg string
								for j := 0; j < len(lots); j++ {
									lotValue := reflect.ValueOf(lots[j])
									lotType := lotValue.Type()
									for i := 0; i < lotType.NumField(); i++ {
										field := lotType.Field(i)
										fieldValue := lotValue.Field(i).Interface()
										var valueStr string
										switch fieldValue.(type) {
										case float32, float64:
											var val string
											if en {
												val = " €"
											} else {
												val = " ₽"
											}
											valueStr = fmt.Sprintf("%.2f", fieldValue) + val
										default:
											valueStr = fmt.Sprintf("%v", fieldValue)
										}
										if field.Name != "ID" && valueStr != "" && field.Name != "Lang" {
											if en {
												msg += field.Name + ": " + valueStr + "\n"
											} else {
												if field.Name == "Price" {
													msg += "Цена" + ": " + valueStr + "\n"
												} else if field.Name == "Description" {
													msg += "Описание" + ": " + valueStr + "\n"
												} else if field.Name == "Category" {
													msg += "Категория" + ": " + valueStr + "\n"
												} else if field.Name == "Server" {
													msg += "Сервер" + ": " + valueStr + "\n"
												} else if field.Name == "Seller" {
													msg += "Продавец" + ": " + valueStr + "\n"
												} else if field.Name == "Amount" {
													msg += "Наличие" + ": " + valueStr + "\n"
												} else if field.Name == "Side" {
													msg += "Сторона" + ": " + valueStr + "\n"
												}
											}
										}
									}
									msg += "\n\n"
								}
								if msg != "" {
									msgRes := tgbotapi.NewMessage(int64(telegramUserID), msg)
									_, err := bot.Send(msgRes)
									if err != nil {
										log.Println("Ошибка отправки сообщения:", err)
									}
								}
							}
						}
					}()

				} else {
					bot.Send(tgbotapi.NewMessage(int64(telegramUserID), "You have 0 active games"))
				}
			} else if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(int64(telegramUserID), "")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Support"),
						tgbotapi.NewKeyboardButton("My Games"),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Settings"),
						tgbotapi.NewKeyboardButton("Premium"),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("/go"),
						tgbotapi.NewKeyboardButton("/stop"),
					),
				)
				msg.Text = "Choose a language, install a Refresh KD and add the games I need to track"
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
			} else if update.Message.Text == "Support" {
				pred = update.Message.Text
				msg := tgbotapi.NewMessage(int64(telegramUserID), "Describe your problem: ")
				bot.Send(msg)
			} else if update.Message.Text == "Refresh KD" {
				pred = update.Message.Text
				msg := tgbotapi.NewMessage(int64(telegramUserID), "Enter a number from 30 to 180 minutes (Buy premium to update every 5 minutes)")
				if user.Premium {
					msg = tgbotapi.NewMessage(int64(telegramUserID), "Enter a number from 5 to 180 minutes")
				}
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
			} else if update.Message.Text == "Settings" {
				msg := tgbotapi.NewMessage(int64(telegramUserID), "")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Language"),
						tgbotapi.NewKeyboardButton("Refresh KD"),
						tgbotapi.NewKeyboardButton("Back"),
					),
				)
				msg.Text = "Settings"
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
			} else if update.Message.Text == "Language" {
				pred = update.Message.Text
				msg := tgbotapi.NewMessage(int64(telegramUserID), "")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Russian"),
						tgbotapi.NewKeyboardButton("English"),
						tgbotapi.NewKeyboardButton("Back"),
					),
				)
				msg.Text = "Language"
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
			} else if update.Message.Text == "My Games" {
				msg := tgbotapi.NewMessage(int64(telegramUserID), "My Games")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Add a game"),
						tgbotapi.NewKeyboardButton("List of my games"),
						tgbotapi.NewKeyboardButton("Back"),
					),
				)
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
				pred = update.Message.Text
			} else if update.Message.Text == "Back" {
				msg := tgbotapi.NewMessage(int64(telegramUserID), "Main menu")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Support"),
						tgbotapi.NewKeyboardButton("My Games"),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Settings"),
						tgbotapi.NewKeyboardButton("Premium"),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("/go"),
						tgbotapi.NewKeyboardButton("/stop"),
					),
				)
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
				pred = ""
			} else if pred == "My Games" {
				if update.Message.Text == "List of my games" {
					msg := tgbotapi.NewMessage(int64(telegramUserID), "You have "+strconv.Itoa(len(AllLots))+" Inactive games")
					var keyboard tgbotapi.InlineKeyboardMarkup
					if len(AllLots) >= 1 {
						var rows []tgbotapi.InlineKeyboardButton
						for _, item := range AllLots {
							btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Inactive")
							rows = append(rows, btn)
						}
						keyboard = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(rows...),
						)
						msg.ReplyMarkup = keyboard
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
					} else {
						msge := tgbotapi.NewMessage(int64(telegramUserID), "You don't have any inactive games")
						bot.Send(msge)
					}
					if len(ActiveLots) >= 1 {
						var rowsActive []tgbotapi.InlineKeyboardButton
						for _, item := range ActiveLots {
							btn := tgbotapi.NewInlineKeyboardButtonData(item.Lot, item.Lot+" Active")
							rowsActive = append(rowsActive, btn)
						}
						msg.Text = strconv.Itoa(len(ActiveLots)) + " Active games"
						keyboard = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(rowsActive...),
						)
						msg.ReplyMarkup = keyboard
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
					} else {
						msge := tgbotapi.NewMessage(int64(telegramUserID), "You don't have any active games")
						bot.Send(msge)
					}
				} else if update.Message.Text == "Add a game" {
					if (len(AllLots) + len(ActiveLots)) <= 20 {
						msg := tgbotapi.NewMessage(int64(telegramUserID), "Enter the game you want to track. Example: chips/37/ You need to enter what is after https://funpay.com/")
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
						pred = "Add a game"
					} else {
						msg := tgbotapi.NewMessage(int64(telegramUserID), "The maximum you can have is only 20 games")
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
					}
				}
			} else if pred == "Add a game" {
				if game == "" {
					if re.MatchString(update.Message.Text) {
						game = update.Message.Text
						msg := tgbotapi.NewMessage(int64(telegramUserID), "Now enter the servers you want to track (if there are none or you want to track all servers, write \"None\") Example: East, West, North, South")
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
					} else {
						msg := tgbotapi.NewMessage(int64(telegramUserID), "Error: Try again. Example: chips/37/")
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
					}
				} else if len(servers) == 0 {
					if update.Message.Text == "None" {
						servers = append(servers, "")
					} else {
						chick := strings.Split(update.Message.Text, ", ")
						for i := 0; i < len(chick); i++ {
							servers = append(servers, chick[i])
						}
					}
					msg := tgbotapi.NewMessage(int64(telegramUserID), "Now enter the maximum price you want to track. Example: 3.5")
					_, err = bot.Send(msg)
					if err != nil {
						fmt.Println(err)
					}
				} else if maxPrice == 0 {
					floatPrice, err := strconv.ParseFloat(update.Message.Text, 64)
					if err != nil {
						msg := tgbotapi.NewMessage(int64(telegramUserID), "Error: Please enter a number. Example: 3.5")
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
					} else {
						maxPrice = floatPrice
						msg := tgbotapi.NewMessage(int64(telegramUserID), "Game added successfully")
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
						if servers[0] == "" {
							servers = []string{}
						}
						newLotik := models.AllLots{
							UserID:   user.ID,
							Lot:      game,
							Servers:  servers,
							MaxPrice: maxPrice,
						}
						database.Create(&newLotik)
						pred = "My Games"
						game = ""
						servers = []string{}
						maxPrice = 0
					}
				}
			} else if pred == "Language" {
				msg := tgbotapi.NewMessage(int64(telegramUserID), "Language changed successfully")
				if update.Message.Text == "English" {
					user.Lang = "English"
					database.Model(&user).Updates(user)
				} else if update.Message.Text == "Russian" {
					user.Lang = "Russian"
					database.Model(&user).Updates(user)
				} else {
					msg = tgbotapi.NewMessage(int64(telegramUserID), "Unknown language")
				}
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Support"),
						tgbotapi.NewKeyboardButton("My Games"),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Settings"),
						tgbotapi.NewKeyboardButton("Premium"),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("/go"),
						tgbotapi.NewKeyboardButton("/stop"),
					),
				)
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
				pred = ""
			} else if pred == "Refresh KD" {
				msg := tgbotapi.NewMessage(int64(telegramUserID), "Refresh KD changed successfully")
				minutes, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					msg.Text = "Error: Please enter a number"
				} else {
					if user.Premium {
						if minutes >= 5 && minutes <= 180 {
							user.RefreshKD = minutes
							database.Model(&user).Updates(user)
							pred = ""
						} else {
							msg.Text = "Error: Please enter a number from 5 to 180"
						}
					} else {
						if minutes >= 30 && minutes <= 180 {
							user.RefreshKD = minutes
							database.Model(&user).Updates(user)
							pred = ""
						} else {
							msg.Text = "Enter a number from 30 to 180 minutes (Buy premium to update every 5 minutes)"
						}
					}
				}
				_, err = bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
			} else if pred == "Support" {
				newReq := models.Support{
					TelegramID: telegramUserID,
					Message:    update.Message.Text,
				}
				database.Create(&newReq)
				msg := tgbotapi.NewMessage(int64(telegramUserID), "Your message has been sent successfully")
				bot.Send(msg)
				pred = ""
			} else {
				msg := tgbotapi.NewMessage(int64(telegramUserID), "Unknown command")
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}

}

// TODO измение цены, сервера, лота, сделать чтоб нельзя чтоб срез активированых лотов был больше 1 если ты не прем
// TODO возможность купить премку и добавить описание бота
