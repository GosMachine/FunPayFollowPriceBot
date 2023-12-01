package funpay

import (
	"errors"
	"gin_test/db"
	"gin_test/logs"
	"gin_test/models"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func deleteOldLots(category string, allLots []models.Lot) {
	if db.Redis.Get(db.Ctx, "DeleteKD:"+category).Val() == "true" {
		return
	}
	var lots []models.Lot
	var deleteLots []models.Lot
	if err := db.Db.Where("category = ?", category).Find(&lots).Error; err != nil {
		logs.Logger.Error("", zap.Error(err))
		return
	}
	for _, dbLot := range lots {
		var found bool
		for _, userLot := range allLots {
			if dbLot.Category == userLot.Category && dbLot.Seller == userLot.Seller && dbLot.Amount == userLot.Amount &&
				dbLot.Price == userLot.Price && dbLot.Description == userLot.Description &&
				dbLot.Server == userLot.Server && dbLot.Side == userLot.Side && !found {
				found = true
				break
			}
		}
		if !found {
			deleteLots = append(deleteLots, dbLot)
		}
	}
	lotsOperations(deleteLots, false)
	db.Redis.Set(db.Ctx, "DeleteKD:"+category, "true", time.Minute*30)
}

func Refresh(lot string, maxPrice float64, servers []string) []models.Lot {
	response, _ := http.Get(lot)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Logger.Error("", zap.Error(err))
		}
	}(response.Body)
	doc, _ := goquery.NewDocumentFromReader(response.Body)
	var res []models.Lot
	var allLots []models.Lot
	category := doc.Find(".content-with-cd").Find("h1").Text()
	doc.Find("a.tc-item").Each(func(i int, c *goquery.Selection) {
		strPrice := strings.TrimSpace(strings.Join(strings.Split(c.Find("div.tc-price").Text(), " "), ""))
		price, _ := strconv.ParseFloat(strPrice[:len(strPrice)-3], 64)

		if price <= maxPrice {
			newLot := models.Lot{Category: category}
			doc.Find(".tc-header").Find("*").Each(func(i int, s *goquery.Selection) {
				class := strings.Split(s.AttrOr("class", ""), " ")
				switch class[0] {
				case "tc-user":
					newLot.Seller = strings.TrimSpace(c.Find("div.media-user-name").Text())
				case "tc-server":
					newLot.Server = strings.TrimSpace(c.Find("div.tc-server").Text())
				case "tc-price":
					newLot.Price = price
				case "tc-amount":
					newLot.Amount = strings.TrimSpace(c.Find("div.tc-amount").Text())
				case "tc-desc":
					newLot.Description = strings.TrimSpace(c.Find("div.tc-desc-text").Text())
				case "tc-side":
					newLot.Side = strings.TrimSpace(c.Find("div.tc-side").Text())
				}
			})
			if in(servers, newLot.Server) || newLot.Server == "" || newLot.Server == "Любой" || len(servers) == 0 {
				var existingLot models.Lot
				allLots = append(allLots, newLot)
				if err := db.Db.Where(&newLot).First(&existingLot).Error; errors.Is(err, gorm.ErrRecordNotFound) {
					res = append(res, newLot)
				}
			}
		}
	})
	deleteOldLots(category, allLots)
	lotsOperations(res, true)
	return res
}

// True == Insert, False == Delete
func lotsOperations(lots []models.Lot, insert bool) {
	if len(lots) == 0 {
		return
	}
	if insert {
		if err := db.Db.Create(&lots).Error; err != nil {
			logs.Logger.Error("", zap.Error(err))
		}
	} else {
		if err := db.Db.Delete(&lots).Error; err != nil {
			logs.Logger.Error("", zap.Error(err))
		}
	}
}

func in(ss []string, s string) bool {
	return strings.Contains(strings.Join(ss, ","), s)
}
