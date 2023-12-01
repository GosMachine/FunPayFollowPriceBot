package funpay

import (
	"errors"
	"gin_test/logs"
	"gin_test/models"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func refresh(database *gorm.DB, lot string, maxPrice float64, servers []string) []models.Lot {
	url := lot
	response, _ := http.Get(url)
	defer response.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(response.Body)
	var res []models.Lot
	tcItems := doc.Find("a.tc-item")

	tcItems.Each(func(i int, c *goquery.Selection) {
		pricee := strings.TrimSpace(strings.Join(strings.Split(c.Find("div.tc-price").Text(), " "), ""))
		price, _ := strconv.ParseFloat(pricee[:len(pricee)-3], 64)

		if price <= maxPrice {
			var newLot models.Lot
			header := doc.Find(".tc-header")
			header.Find("*").Each(func(i int, s *goquery.Selection) {
				class := strings.Split(s.AttrOr("class", ""), " ")
				text := strings.TrimSpace(s.Text())

				switch class[0] {
				case "tc-user":
					newLot.Seller = text
				case "tc-server":
					newLot.Server = text
				case "tc-price":
					newLot.Price = price
				case "tc-amount":
					newLot.Amount = text
				case "tc-desc":
					newLot.Description = text
				case "tc-side":
					newLot.Side = text
				}
			})

			if price <= maxPrice && (in(servers, newLot.Server) || newLot.Server == "" || newLot.Server == "Любой" || newLot.Server == "Any" || len(servers) == 0) {
				var existingLot models.Lot
				if err := database.Where(&newLot).First(&existingLot).Error; errors.Is(err, gorm.ErrRecordNotFound) {
					database.Create(&newLot)
					res = append(res, newLot)
				}
			}
		}
	})

	return res
}

func in(ss []string, s string) bool {
	return strings.Contains(strings.Join(ss, ","), s)
}

func refresho(database *gorm.DB, en bool, lot string, maxPrice float64, servers []string) []models.Lot {
	url := lot
	var response *http.Response
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept-Language", "ru")
	client := &http.Client{}
	response, _ = client.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Logger.Error("", zap.Error(err))
		}
	}(response.Body)
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
					if errors.Is(err, gorm.ErrRecordNotFound) {
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
