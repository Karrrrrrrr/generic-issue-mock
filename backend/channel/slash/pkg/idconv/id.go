package idconv

import (
	"encoding/binary"
	"math"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/model"

	"github.com/google/uuid"
)

func ToUUID(id model.ID) string {
	value := uuid.UUID{}
	binary.BigEndian.PutUint64(value[8:], uint64(id))
	return value.String()
}

func ToOptionalUUID(id *model.ID) string {
	if id == nil {
		return ""
	}
	return ToUUID(*id)
}

func FromUUID(value string) (model.ID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value || parsed[0] != 0 || parsed[1] != 0 || parsed[2] != 0 || parsed[3] != 0 ||
		parsed[4] != 0 || parsed[5] != 0 || parsed[6] != 0 || parsed[7] != 0 {
		return 0, slasherrors.ErrResourceNotFound
	}

	id := binary.BigEndian.Uint64(parsed[8:])
	if id == 0 || id > math.MaxInt64 {
		return 0, slasherrors.ErrResourceNotFound
	}
	return model.ID(id), nil
}

func FromOptionalUUID(value *string) (*model.ID, error) {
	if value == nil {
		return nil, nil
	}
	id, err := FromUUID(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func FromAccountUUID(value string) (model.ID, error) {
	return FromUUID(value)
}
