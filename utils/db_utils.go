package utils

import (
	"gin_test/models"
	"strconv"
)

func FindAllLotsItem(user *models.User, data string) (*models.AllLots, int) {
	var item models.AllLots
	var index int
	for i, it := range user.AllLots {
		if strconv.Itoa(int(it.ID)) == data {
			item = it
			index = i
			break
		}
	}
	return &item, index
}
