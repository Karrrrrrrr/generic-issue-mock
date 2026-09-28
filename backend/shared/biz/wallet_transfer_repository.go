package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
)

type WalletTransferFilters struct {
	Channel     enums.Channel
	IDs         []model.ID
	AccountIDs  []model.ID
	CardIDs     []model.ID
	Kinds       []enums.WalletTransferKind
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type WalletTransferListRequest struct {
	WalletTransferFilters
	Offset int
	Limit  int
}

type WalletTransferCountRequest struct {
	WalletTransferFilters
}

type WalletTransferCreateRequest struct {
	WalletTransfer *model.WalletTransfer
}

type WalletTransferExistByRequestIDRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	RequestID string
}

type WalletTransferFindByRequestIDRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	RequestID string
}

type WalletTransferExistRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	ID        model.ID
}

type WalletTransferFindRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	ID        model.ID
}

type WalletTransferRepo interface {
	Exist(context.Context, *WalletTransferExistRequest) (bool, error)
	Find(context.Context, *WalletTransferFindRequest) (*model.WalletTransfer, error)
	List(context.Context, *WalletTransferListRequest) ([]*model.WalletTransfer, error)
	Count(context.Context, *WalletTransferCountRequest) (int64, error)
	Create(context.Context, *WalletTransferCreateRequest) error
	ExistByRequestID(context.Context, *WalletTransferExistByRequestIDRequest) (bool, error)
	FindByRequestID(context.Context, *WalletTransferFindByRequestIDRequest) (*model.WalletTransfer, error)
}
