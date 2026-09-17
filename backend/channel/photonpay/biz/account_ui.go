package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type UICreateAccountRequest struct{ Name string }

type UIUpdateAccountRequest struct {
	ID   model.ID
	Name string
}

func (u *PhotonPayUIUsecase) CreateAccount(ctx context.Context, req *UICreateAccountRequest) (*model.Account, error) {
	var item *model.Account
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		item = &model.Account{
			Channel: enums.Channel_PhotonPay,
			Name:    req.Name,
		}
		if err := u.accountRepo.Create(txCtx, item); err != nil {
			zap.S().Errorw("create photonpay UI account", "error", err)
			return ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: item.ID,
			Channel:   enums.Channel_PhotonPay,
			Type:      enums.WalletType_Account,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepo.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create photonpay UI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		item.WalletID = wallet.ID
		if err := u.accountRepo.Save(txCtx, item); err != nil {
			zap.S().Errorw("attach photonpay UI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		virtualWallet := &model.Wallet{
			AccountID: item.ID,
			Channel:   enums.Channel_PhotonPay,
			Type:      enums.WalletType_VirtualAccount,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepo.Create(txCtx, virtualWallet); err != nil {
			zap.S().Errorw("create photonpay account virtual wallet", "error", err)
			return ErrDatabaseOperation
		}
		virtualAccount := &model.VirtualAccount{
			AccountID: item.ID,
			Channel:   enums.Channel_PhotonPay,
			WalletID:  virtualWallet.ID,
			Name:      item.Name,
		}
		if err := u.virtualAccountRepo.Create(txCtx, virtualAccount); err != nil {
			zap.S().Errorw("create photonpay account virtual account", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (u *PhotonPayUIUsecase) ListAccounts(ctx context.Context, req *ListRequest) ([]*model.Account, int64, error) {
	items, err := u.accountRepo.List(ctx, &AccountListRequest{
		IDs:    types.PointerSlice(req.AccountID),
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay UI accounts", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.accountRepo.Count(ctx, &AccountCountRequest{
		IDs: types.PointerSlice(req.AccountID),
	})
	if err != nil {
		zap.S().Errorw("count photonpay UI accounts", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	return items, total, nil
}

func (u *PhotonPayUIUsecase) UpdateAccount(ctx context.Context, req *UIUpdateAccountRequest) (*model.Account, error) {
	exists, err := u.accountRepo.Exist(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check photonpay UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := u.accountRepo.Find(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find photonpay UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.Name = req.Name
	if err := u.accountRepo.Save(ctx, item); err != nil {
		zap.S().Errorw("rename photonpay UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}
