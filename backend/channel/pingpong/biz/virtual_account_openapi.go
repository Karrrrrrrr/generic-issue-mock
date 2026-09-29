package biz

import (
	"context"
	"strings"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharedbiz "generic-mock/shared/biz"

	"go.uber.org/zap"
)

type CreateVirtualAccountRequest struct {
	AccountID model.ID
	Name      string
}

type VirtualAccountBalancesRequest struct {
	AccountID model.ID
	ID        *model.ID
}

type openAPIVirtualAccountReference struct {
	AccountID model.ID
	ID        model.ID
}

func (uc *PingPongOpenAPIUsecase) CreateVirtualAccount(ctx context.Context, req *CreateVirtualAccountRequest) (*model.VirtualAccount, error) {
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
		if err := uc.sharedWalletRepo.Create(ctx, &sharedbiz.WalletCreateRequest{Wallet: wallet}); err != nil {
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

func (uc *PingPongOpenAPIUsecase) VirtualAccountBalances(ctx context.Context, req *VirtualAccountBalancesRequest) ([]*model.VirtualAccount, error) {
	if req.AccountID <= 0 || (req.ID != nil && *req.ID <= 0) {
		return nil, pingerrors.ErrInvalid
	}
	items, err := uc.virtualAccountRepo.List(ctx, &VirtualAccountListRequest{
		AccountIDs: []model.ID{req.AccountID},
		IDs:        types.PointerSlice(req.ID),
	})
	if err != nil {
		zap.S().Errorw("list pingpong virtual account balances", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return items, nil
}

func (uc *PingPongOpenAPIUsecase) getVirtualAccount(ctx context.Context, req *openAPIVirtualAccountReference) (*model.VirtualAccount, error) {
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
