package biz

import (
	"context"
	"time"

	common "generic-mock/enums"
	"generic-mock/model"
)

type TransferExistsRequest struct {
	AccountID model.ID
	ID        model.ID
}

type TransferFindRequest struct {
	AccountID model.ID
	ID        model.ID
}

type TransferListRequest struct {
	IDs         []model.ID
	AccountIDs  []model.ID
	RequestIDs  []string
	CardIDs     []model.ID
	Kinds       []common.WalletTransferKind
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Offset      int
	Limit       *int
}

type TransferCountRequest struct {
	IDs         []model.ID
	AccountIDs  []model.ID
	RequestIDs  []string
	CardIDs     []model.ID
	Kinds       []common.WalletTransferKind
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type TransferRequestExistsRequest struct {
	AccountID model.ID
	RequestID string
}

type TransferRequestFindRequest struct {
	AccountID model.ID
	RequestID string
}

type PingPongTransferRepository interface {
	ExistsRequest(context.Context, *TransferRequestExistsRequest) (bool, error)
	FindRequest(context.Context, *TransferRequestFindRequest) (*model.WalletTransfer, error)
	Exists(context.Context, *TransferExistsRequest) (bool, error)
	Find(context.Context, *TransferFindRequest) (*model.WalletTransfer, error)

	Create(context.Context, *model.WalletTransfer) error
	List(context.Context, *TransferListRequest) ([]*model.WalletTransfer, error)
	Count(context.Context, *TransferCountRequest) (int64, error)
}
