package utils

import (
	"fmt"
	"gin_test/db"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func LotName(text, strChatID string) error {
	runeCount := utf8.RuneCountInString(text)
	fmt.Println(runeCount)
	if runeCount >= 3 && runeCount <= 9 {
		db.Redis.Set(db.Ctx, "name:"+strChatID, text, time.Hour)
		return nil
	}
	return fmt.Errorf("bad length")
}

func LotGame(text, strChatID string) error {
	re := regexp.MustCompile("^https://funpay.com/[a-z]+/[0-9]+/$")
	if re.MatchString(text) {
		db.Redis.Set(db.Ctx, "game:"+strChatID, text, time.Hour)
		return nil
	}
	return fmt.Errorf("bad link")
}

func LotServers(text, strChatID string) {
	chick := strings.Split(text, ", ")
	interfaceElements := make([]interface{}, len(chick))
	for i, v := range chick {
		interfaceElements[i] = v
	}
	db.Redis.RPush(db.Ctx, "servers:"+strChatID, interfaceElements...)
}

func LotMaxPrice(text, strChatID string) (float64, error) {
	maxPriceFloat, err := strconv.ParseFloat(text, 64)
	if err == nil {
		db.Redis.Set(db.Ctx, "maxPrice:"+strChatID, text, time.Hour)
		return maxPriceFloat, nil
	}
	return 0, err
}
