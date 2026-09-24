package biz

import (
	"context"

	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type UIListTransfersRequest struct {
	AccountID *model.ID
	Offset    int
	Limit     int
}

func (uc *PingPongUIUsecase) ListTransfers(ctx context.Context, req *UIListTransfersRequest) ([]*model.WalletTransfer, int64, error) {
	items, err := uc.walletTransferRepo.List(ctx, &TransferListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Offset:     req.Offset,
		Limit:      &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong UI transfers", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.walletTransferRepo.Count(ctx, &TransferCountRequest{AccountIDs: types.PointerSlice(req.AccountID)})
	if err != nil {
		zap.S().Errorw("count pingpong UI transfers", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}
