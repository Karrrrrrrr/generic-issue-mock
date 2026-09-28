package biz

import (
	"context"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type UICreateVirtualAccountRequest struct {
	AccountID model.ID
	Name      string
}

type CreateManagedVirtualAccountRequest struct {
	AccountID model.ID
	Name      string
	Currency  enums.Currency
}

func (u *PhotonPayUIUsecase) CreateVirtualAccount(
	ctx context.Context,
	req *UICreateVirtualAccountRequest,
) (*model.VirtualAccount, error) {
	var item *model.VirtualAccount
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.accountRepo.Exist(txCtx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check photonpay virtual account", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		if !exists {
			return photonpayerrors.ErrResourceNotFound
		}

		account, err := u.accountRepo.Find(txCtx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay virtual account owner", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: req.AccountID,
			Channel:   enums.Channel_PhotonPay,
			Type:      enums.WalletType_VirtualAccount,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepo.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create photonpay virtual account wallet", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}

		item = &model.VirtualAccount{
			Account:   account,
			AccountID: req.AccountID,
			Channel:   enums.Channel_PhotonPay,
			WalletID:  wallet.ID,
			Name:      req.Name,
			Wallet:    wallet,
		}
		if err := u.virtualAccountRepo.Create(txCtx, item); err != nil {
			zap.S().Errorw("create photonpay virtual account", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (u *PhotonPayUIUsecase) ListVirtualAccounts(ctx context.Context) ([]*model.VirtualAccount, error) {
	items, err := u.virtualAccountRepo.ListVirtualAccounts(ctx, &VirtualAccountListRequest{})
	if err != nil {
		zap.S().Errorw("list photonpay virtual accounts", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	return items, nil
}

func (u *PhotonPayUIUsecase) ListManagedVirtualAccounts(ctx context.Context, accountID *model.ID) ([]*model.VirtualAccount, error) {
	items, err := u.virtualAccountRepo.ListVirtualAccounts(ctx, &VirtualAccountListRequest{
		AccountIDs: types.PointerSlice(accountID),
	})
	if err != nil {
		zap.S().Errorw("list photonpay virtual accounts", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	return items, nil
}

func (u *PhotonPayUIUsecase) CreateManagedVirtualAccount(ctx context.Context, req *CreateManagedVirtualAccountRequest) (*model.VirtualAccount, error) {
	var item *model.VirtualAccount
	err := u.transaction.InTx(ctx, func(ctx context.Context) error {
		exists, err := u.accountRepo.Exist(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check photonpay virtual account owner", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		if !exists {
			return photonpayerrors.ErrResourceNotFound
		}
		account, err := u.accountRepo.Find(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay virtual account owner", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: req.AccountID,
			Channel:   enums.Channel_PhotonPay,
			Type:      enums.WalletType_VirtualAccount,
			Currency:  req.Currency,
		}
		if err := u.walletRepo.Create(ctx, wallet); err != nil {
			zap.S().Errorw("create photonpay virtual wallet", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		item = &model.VirtualAccount{
			Account:   account,
			AccountID: req.AccountID,
			Channel:   enums.Channel_PhotonPay,
			WalletID:  wallet.ID,
			Name:      req.Name,
		}
		if err := u.virtualAccountRepo.Create(ctx, item); err != nil {
			zap.S().Errorw("create photonpay virtual account", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		return nil
	})
	return item, err
}
