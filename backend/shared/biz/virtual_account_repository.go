package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type VirtualAccountExistRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type VirtualAccountFindRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type VirtualAccountRepo interface {
	Exist(context.Context, *VirtualAccountExistRequest) (bool, error)
	Find(context.Context, *VirtualAccountFindRequest) (*model.VirtualAccount, error)
}
