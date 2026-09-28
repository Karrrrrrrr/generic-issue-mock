package biz

import (
	"context"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type CreateManagedVirtualAccountRequest struct {
	AccountID model.ID
	Name      string
	Currency  enums.Currency
}

func (u *SlashUIUsecase) ListVirtualAccounts(ctx context.Context) ([]*model.VirtualAccount, error) {
	items, err := u.virtualAccountRepository.ListVirtualAccounts(ctx, &VirtualAccountListRequest{})
	if err != nil {
		zap.S().Errorw("list slash virtual accounts", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	return items, nil
}

func (u *SlashUIUsecase) ListManagedVirtualAccounts(ctx context.Context, accountID *model.ID) ([]*model.VirtualAccount, error) {
	items, err := u.virtualAccountRepository.ListVirtualAccounts(ctx, &VirtualAccountListRequest{
		AccountIDs: types.PointerSlice(accountID),
	})
	if err != nil {
		zap.S().Errorw("list slash virtual accounts", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	return items, nil
}

func (u *SlashUIUsecase) CreateManagedVirtualAccount(ctx context.Context, req *CreateManagedVirtualAccountRequest) (*model.VirtualAccount, error) {
	var item *model.VirtualAccount
	err := u.transaction.InTx(ctx, func(ctx context.Context) error {
		exists, err := u.accountRepository.Exist(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check slash virtual account owner", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		if !exists {
			return slasherrors.ErrResourceNotFound
		}
		account, err := u.accountRepository.Find(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find slash virtual account owner", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: req.AccountID,
			Channel:   enums.Channel_Slash,
			Type:      enums.WalletType_VirtualAccount,
			Currency:  req.Currency,
		}
		if err := u.walletRepository.Create(ctx, wallet); err != nil {
			zap.S().Errorw("create slash virtual wallet", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		item = &model.VirtualAccount{
			Account:   account,
			AccountID: req.AccountID,
			Channel:   enums.Channel_Slash,
			WalletID:  wallet.ID,
			Name:      req.Name,
		}
		if err := u.virtualAccountRepository.CreateVirtualAccount(ctx, item); err != nil {
			zap.S().Errorw("create slash virtual account", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		return nil
	})
	return item, err
}
