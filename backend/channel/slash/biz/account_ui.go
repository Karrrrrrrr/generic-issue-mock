package biz

import (
	"context"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type ListAccountsRequest struct {
	AccountID *model.ID
	Offset    int
	Limit     int
}

type CreateAccountRequest struct {
	Name string
}

type UpdateAccountRequest struct {
	ID   model.ID
	Name string
}

func (u *SlashUIUsecase) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*model.Account, error) {
	var item *model.Account
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		item = &model.Account{
			Channel: enums.Channel_Slash,
			Name:    req.Name,
		}
		if err := u.accountRepository.Create(txCtx, item); err != nil {
			zap.S().Errorw("create slash UI account", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: item.ID,
			Channel:   enums.Channel_Slash,
			Type:      enums.WalletType_Account,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create slash UI account wallet", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		item.WalletID = wallet.ID
		if err := u.accountRepository.Save(txCtx, item); err != nil {
			zap.S().Errorw("attach slash UI account wallet", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (u *SlashUIUsecase) ListAccounts(
	ctx context.Context,
	req *ListAccountsRequest,
) ([]*model.Account, int64, error) {
	items, err := u.accountRepository.List(ctx, &AccountListRequest{
		IDs:    types.PointerSlice(req.AccountID),
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list slash UI accounts", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
	}

	total, err := u.accountRepository.Count(ctx, &AccountCountRequest{
		IDs: types.PointerSlice(req.AccountID),
	})
	if err != nil {
		zap.S().Errorw("count slash UI accounts", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *SlashUIUsecase) UpdateAccount(
	ctx context.Context,
	req *UpdateAccountRequest,
) (*model.Account, error) {
	exists, err := u.accountRepository.Exist(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check slash UI account", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, slasherrors.ErrResourceNotFound
	}

	item, err := u.accountRepository.Find(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find slash UI account", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	item.Name = req.Name
	if err := u.accountRepository.Save(ctx, item); err != nil {
		zap.S().Errorw("update slash UI account", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}

	return item, nil
}
