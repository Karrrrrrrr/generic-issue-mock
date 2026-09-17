package service

import (
	"encoding/binary"
	"math"

	"generic-mock/channel/slash/biz"
	"generic-mock/model"

	"github.com/google/uuid"
)

func slashIDString(id model.ID) string {
	value := uuid.UUID{}
	binary.BigEndian.PutUint64(value[8:], uint64(id))
	return value.String()
}

func slashID(value string) (model.ID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value || parsed[0] != 0 || parsed[1] != 0 || parsed[2] != 0 || parsed[3] != 0 ||
		parsed[4] != 0 || parsed[5] != 0 || parsed[6] != 0 || parsed[7] != 0 {
		return 0, biz.ErrResourceNotFound
	}

	id := binary.BigEndian.Uint64(parsed[8:])
	if id == 0 || id > math.MaxInt64 {
		return 0, biz.ErrResourceNotFound
	}
	return model.ID(id), nil
}

func slashOptionalID(value *string) (*model.ID, error) {
	if value == nil {
		return nil, nil
	}
	id, err := slashID(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func slashAccountID(value string) (model.ID, error) {
	return slashID(value)
}

// slashRefundAuthorizationID permits the channel-formatted zero ID for an independent refund.
func slashRefundAuthorizationID(value *string) (*model.ID, error) {
	if value == nil {
		return nil, nil
	}
	var id model.ID
	if *value == slashIDString(0) {
		return &id, nil
	}
	id, err := slashID(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
