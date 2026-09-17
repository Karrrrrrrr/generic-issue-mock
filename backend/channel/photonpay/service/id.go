package service

import (
	"strconv"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/model"
)

func photonPayIDString(id model.ID) string {
	return strconv.FormatInt(id, 10)
}

func photonPayID(value string) (model.ID, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return 0, biz.ErrResourceNotFound
	}
	return model.ID(id), nil
}

func photonPayAccountID(value string) (model.ID, error) {
	return photonPayID(value)
}

// photonpayRefundAuthorizationID permits the channel-formatted zero ID for an independent refund.
func photonpayRefundAuthorizationID(value *string) (*model.ID, error) {
	if value == nil {
		return nil, nil
	}
	var id model.ID
	if *value == photonPayIDString(0) {
		return &id, nil
	}
	id, err := photonPayID(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func photonPayOptionalID(value *string) (*model.ID, error) {
	if value == nil {
		return nil, nil
	}
	id, err := photonPayID(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
