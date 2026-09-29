package idconv

import (
	pingerrors "generic-mock/channel/pingpong/errors"
	"strconv"
)

func ToString(id int64) string { return strconv.FormatInt(id, 10) }

func FromString(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 || ToString(id) != value {
		return 0, pingerrors.ErrInvalid
	}
	return id, nil
}

func FromOptionalString(value *string) (*int64, error) {
	if value == nil {
		return nil, nil
	}
	id, err := FromString(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func FromStringList(values []string) ([]int64, error) {
	result := make([]int64, 0, len(values))
	for _, value := range values {
		id, err := FromString(value)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}
