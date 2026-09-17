package service

import (
	"strconv"

	"generic-mock/channel/paynda/biz"
	"generic-mock/model"
)

func payndaIDString(id model.ID) string {
	return strconv.FormatInt(id, 10)
}

func payndaID(value string) (model.ID, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return 0, biz.ErrResourceNotFound
	}
	return model.ID(id), nil
}

func payndaAccountID(value string) (model.ID, error) {
	return payndaID(value)
}
