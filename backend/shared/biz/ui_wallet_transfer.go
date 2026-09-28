package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"go.uber.org/zap"
)

type ListUIWalletTransfersRequest struct {
	UIPageRequest
	UITimeRange
	ID        *model.ID
	AccountID *model.ID
	CardID    *model.ID
	Kind      *enums.WalletTransferKind
}

func (req *ListUIWalletTransfersRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID, req.AccountID, req.CardID}) ||
		(req.Kind != nil && *req.Kind != enums.WalletTransfer_CardTopUp && *req.Kind != enums.WalletTransfer_CardWithdraw && *req.Kind != enums.WalletTransfer_VirtualAccountTopUp && *req.Kind != enums.WalletTransfer_VirtualAccountTransfer) {
		return sharederrors.ErrInvalidUIRequest
	}
	if err := req.UITimeRange.Validate(); err != nil {
		return err
	}
	return req.UIPageRequest.Validate()
}

type UIWalletTransfers interface {
	ListWalletTransfers(context.Context, *ListUIWalletTransfersRequest) ([]*model.WalletTransfer, int64, error)
}

func (uc *ui) ListWalletTransfers(ctx context.Context, req *ListUIWalletTransfersRequest) ([]*model.WalletTransfer, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := WalletTransferFilters{
		Channel:     uc.channel,
		IDs:         types.PointerSlice(req.ID),
		AccountIDs:  types.PointerSlice(req.AccountID),
		CardIDs:     types.PointerSlice(req.CardID),
		Kinds:       types.PointerSlice(req.Kind),
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
	}
	items, err := uc.walletTransferRepo.List(ctx, &WalletTransferListRequest{
		WalletTransferFilters: filters,
		Offset:                req.Offset,
		Limit:                 req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI wallet_transfer", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.walletTransferRepo.Count(ctx, &WalletTransferCountRequest{WalletTransferFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI wallet_transfer", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}
