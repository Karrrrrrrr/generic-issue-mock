package biz

import (
	"context"
	"strings"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type UICreateVirtualAccountRequest struct {
	AccountID model.ID
	Name      string
}

type UIListVirtualAccountsRequest struct {
	AccountID *model.ID
	Offset    int
	Limit     int
}

type uiVirtualAccountReference struct {
	AccountID model.ID
	ID        model.ID
}

func (uc *PingPongUIUsecase) CreateVirtualAccount(ctx context.Context, req *UICreateVirtualAccountRequest) (*model.VirtualAccount, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, pingerrors.ErrInvalid
	}
	var virtualAccount *model.VirtualAccount
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		account, err := uc.lockAccount(ctx, req.AccountID)
		if err != nil {
			return err
		}
		wallet := &model.Wallet{
			AccountID: account.ID,
			Channel:   common.Channel_PingPong,
			Type:      common.WalletType_VirtualAccount,
			Currency:  common.Currency_USD,
		}
		if err := uc.walletRepo.Create(ctx, wallet); err != nil {
			zap.S().Errorw("create pingpong virtual account wallet", "error", err)
			return pingerrors.ErrDatabase
		}
		virtualAccount = &model.VirtualAccount{
			AccountID: account.ID,
			Channel:   common.Channel_PingPong,
			WalletID:  wallet.ID,
			Name:      strings.TrimSpace(req.Name),
			Account:   account,
			Wallet:    wallet,
		}
		if err := uc.virtualAccountRepo.Create(ctx, virtualAccount); err != nil {
			zap.S().Errorw("create pingpong virtual account", "error", err)
			return pingerrors.ErrDatabase
		}
		return nil
	})
	return virtualAccount, err
}

func (uc *PingPongUIUsecase) ListVirtualAccounts(ctx context.Context, req *UIListVirtualAccountsRequest) ([]*model.VirtualAccount, int64, error) {
	items, err := uc.virtualAccountRepo.List(ctx, &VirtualAccountListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Offset:     req.Offset,
		Limit:      &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong UI virtual accounts", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.virtualAccountRepo.Count(ctx, &VirtualAccountCountRequest{AccountIDs: types.PointerSlice(req.AccountID)})
	if err != nil {
		zap.S().Errorw("count pingpong UI virtual accounts", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}

func (uc *PingPongUIUsecase) getVirtualAccount(ctx context.Context, req *uiVirtualAccountReference) (*model.VirtualAccount, error) {
	if req.AccountID <= 0 || req.ID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.virtualAccountRepo.Exists(ctx, &VirtualAccountExistsRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong virtual account", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	virtualAccount, err := uc.virtualAccountRepo.Find(ctx, &VirtualAccountFindRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong virtual account", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return virtualAccount, nil
}
