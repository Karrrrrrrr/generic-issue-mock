package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"go.uber.org/zap"
)

type ListUIWalletsRequest struct {
	UIPageRequest
	ID        *model.ID
	AccountID *model.ID
	Type      *enums.WalletType
}

func (req *ListUIWalletsRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID, req.AccountID}) ||
		(req.Type != nil && *req.Type != enums.WalletType_Account && *req.Type != enums.WalletType_Card && *req.Type != enums.WalletType_VirtualAccount) {
		return sharederrors.ErrInvalidUIRequest
	}
	return req.UIPageRequest.Validate()
}

type UIWallets interface {
	ListWallets(context.Context, *ListUIWalletsRequest) ([]*model.Wallet, int64, error)
}

func (uc *ui) ListWallets(ctx context.Context, req *ListUIWalletsRequest) ([]*model.Wallet, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := WalletFilters{
		Channel:    uc.channel,
		IDs:        types.PointerSlice(req.ID),
		AccountIDs: types.PointerSlice(req.AccountID),
		Types:      types.PointerSlice(req.Type),
	}
	items, err := uc.walletRepo.List(ctx, &WalletListRequest{
		WalletFilters: filters,
		Offset:        req.Offset,
		Limit:         req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI wallet", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.walletRepo.Count(ctx, &WalletCountRequest{WalletFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI wallet", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}
