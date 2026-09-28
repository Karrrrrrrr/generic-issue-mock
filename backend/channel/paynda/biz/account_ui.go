package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type PayndaUICreateAccountRequest struct {
	Name string
}

type PayndaUIUpdateAccountRequest struct {
	ID   model.ID
	Name string
}

func (u *PayndaUIUsecase) CreateAccount(
	ctx context.Context,
	req *PayndaUICreateAccountRequest,
) (*PayndaAccountWallet, error) {
	var result *PayndaAccountWallet
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		account := &model.Account{
			Channel: enums.Channel_Paynda,
			Name:    req.Name,
		}
		if err := u.accountRepository.Create(txCtx, account); err != nil {
			zap.S().Errorw("create paynda UI account", "error", err)
			return ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: account.ID,
			Channel:   enums.Channel_Paynda,
			Available: decimal.Zero,
			Type:      enums.WalletType_Account,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create paynda UI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		account.WalletID = wallet.ID
		if err := u.accountRepository.Save(txCtx, account); err != nil {
			zap.S().Errorw("attach paynda UI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		result = &PayndaAccountWallet{
			Account: account,
			Wallet:  wallet,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *PayndaUIUsecase) ListAccounts(
	ctx context.Context,
	req *PayndaListRequest,
) ([]*model.Account, int64, error) {
	items, err := u.accountRepository.List(ctx, &AccountListRequest{
		IDs:    types.PointerSlice(req.AccountID),
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list paynda UI accounts", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	total, err := u.accountRepository.Count(ctx, &AccountCountRequest{
		IDs: types.PointerSlice(req.AccountID),
	})
	if err != nil {
		zap.S().Errorw("count paynda UI accounts", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *PayndaUIUsecase) UpdateAccount(
	ctx context.Context,
	req *PayndaUIUpdateAccountRequest,
) (*model.Account, error) {
	exists, err := u.accountRepository.ExistByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check paynda UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item, err := u.accountRepository.FindByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find paynda UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.Name = req.Name
	if err := u.accountRepository.Save(ctx, item); err != nil {
		zap.S().Errorw("update paynda UI account", "error", err)
		return nil, ErrDatabaseOperation
	}

	return item, nil
}
