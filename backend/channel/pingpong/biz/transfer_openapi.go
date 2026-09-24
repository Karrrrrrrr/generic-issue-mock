package biz

import (
	"context"
	"time"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type CardOrdersRequest struct {
	AccountID      model.ID
	CardID         *model.ID
	RequestID      *string
	From           *time.Time
	To             *time.Time
	Offset         int
	Limit          int
	SuccessfulOnly bool
}

type VirtualAccountOrderRequest struct {
	AccountID model.ID
	ID        model.ID
	Kind      common.WalletTransferKind
}

func (uc *PingPongOpenAPIUsecase) CardOrders(ctx context.Context, req *CardOrdersRequest) ([]*model.WalletTransfer, int64, error) {
	if !req.SuccessfulOnly {
		return []*model.WalletTransfer{}, 0, nil
	}
	kinds := []common.WalletTransferKind{common.WalletTransfer_CardTopUp, common.WalletTransfer_CardWithdraw}
	items, err := uc.walletTransferRepo.List(ctx, &TransferListRequest{
		AccountIDs:  []model.ID{req.AccountID},
		CardIDs:     types.PointerSlice(req.CardID),
		RequestIDs:  types.PointerSlice(req.RequestID),
		Kinds:       kinds,
		CreatedFrom: req.From,
		CreatedTo:   req.To,
		Offset:      req.Offset,
		Limit:       &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong card funding orders", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.walletTransferRepo.Count(ctx, &TransferCountRequest{
		AccountIDs:  []model.ID{req.AccountID},
		CardIDs:     types.PointerSlice(req.CardID),
		RequestIDs:  types.PointerSlice(req.RequestID),
		Kinds:       kinds,
		CreatedFrom: req.From,
		CreatedTo:   req.To,
	})
	if err != nil {
		zap.S().Errorw("count pingpong card funding orders", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}

func (uc *PingPongOpenAPIUsecase) VirtualAccountOrder(ctx context.Context, req *VirtualAccountOrderRequest) (*model.WalletTransfer, error) {
	exists, err := uc.walletTransferRepo.Exists(ctx, &TransferExistsRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong virtual account funding order", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	item, err := uc.walletTransferRepo.Find(ctx, &TransferFindRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong virtual account funding order", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if item.Kind != req.Kind {
		return nil, pingerrors.ErrNotFound
	}
	return item, nil
}
