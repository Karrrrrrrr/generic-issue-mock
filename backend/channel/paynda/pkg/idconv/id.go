package idconv

import (
	"strconv"

	payndaerrors "generic-mock/channel/paynda/errors"
	"generic-mock/model"
)

func ToString(id model.ID) string {
	return strconv.FormatInt(id, 10)
}

func FromString(value string) (model.ID, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return 0, payndaerrors.ErrResourceNotFound
	}
	return model.ID(id), nil
}

func FromAccountString(value string) (model.ID, error) {
	return FromString(value)
}

func FromOptionalString(value *string) (*model.ID, error) {
	if value == nil {
		return nil, nil
	}
	id, err := FromString(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
