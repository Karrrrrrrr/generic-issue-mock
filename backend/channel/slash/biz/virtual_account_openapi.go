package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type OpenAPIVirtualAccountTransferRequest struct {
	AccountID   model.ID
	Source      model.ID
	Destination model.ID
	AmountCents int64
}

type OpenAPICreateVirtualAccountRequest struct {
	AccountID model.ID
	Name      string
}

type OpenAPIUpdateVirtualAccountRequest struct {
	AccountID model.ID
	ID        model.ID
	Name      string
}

func (u *SlashOpenAPIUsecase) GetVirtualAccount(ctx context.Context, req *ResourceRequest) (*model.VirtualAccount, error) {
	exists, err := u.virtualAccountRepository.ExistByAccountID(ctx, &VirtualAccountExistByAccountIDRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check slash OpenAPI virtual account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := u.virtualAccountRepository.FindByAccountID(ctx, &VirtualAccountFindByAccountIDRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find slash OpenAPI virtual account", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *SlashOpenAPIUsecase) CreateVirtualAccount(ctx context.Context, req *OpenAPICreateVirtualAccountRequest) (*model.VirtualAccount, error) {
	if _, err := u.GetAccount(ctx, req.AccountID); err != nil {
		return nil, err
	}
	var item *model.VirtualAccount
	err := u.transaction.InTx(ctx, func(ctx context.Context) error {
		wallet := &model.Wallet{
			AccountID: req.AccountID,
			Channel:   enums.Channel_Slash,
			Type:      enums.WalletType_VirtualAccount,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepository.Create(ctx, wallet); err != nil {
			zap.S().Errorw("create slash OpenAPI virtual wallet", "error", err)
			return ErrDatabaseOperation
		}
		item = &model.VirtualAccount{
			AccountID: req.AccountID,
			Channel:   enums.Channel_Slash,
			WalletID:  wallet.ID,
			Name:      req.Name,
			Wallet:    wallet,
		}
		if err := u.virtualAccountRepository.CreateVirtualAccount(ctx, item); err != nil {
			zap.S().Errorw("create slash OpenAPI virtual account", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	return item, err
}

func (usecase *SlashOpenAPIUsecase) UpdateVirtualAccount(ctx context.Context, req *OpenAPIUpdateVirtualAccountRequest) (*model.VirtualAccount, error) {
	var item *model.VirtualAccount
	err := usecase.transaction.InTx(ctx, func(ctx context.Context) error {
		var err error
		item, err = usecase.GetVirtualAccount(ctx, &ResourceRequest{
			AccountID: &req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			return err
		}
		if err := usecase.virtualAccountRepository.Save(ctx, &SaveVirtualAccountRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
			Name:      req.Name,
		}); err != nil {
			zap.S().Errorw("rename slash OpenAPI virtual account", "error", err)
			return ErrDatabaseOperation
		}
		item.Name = req.Name
		return nil
	})
	return item, err
}

func (u *SlashOpenAPIUsecase) ListVirtualAccounts(ctx context.Context, accountID model.ID) ([]*model.VirtualAccount, error) {
	items, err := u.virtualAccountRepository.ListVirtualAccounts(ctx, &VirtualAccountListRequest{
		AccountIDs: []model.ID{accountID},
	})
	if err != nil {
		zap.S().Errorw("list slash virtual accounts", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

func (u *SlashOpenAPIUsecase) TransferVirtualAccount(ctx context.Context, req *OpenAPIVirtualAccountTransferRequest) error {
	if req.Source == req.Destination || req.AmountCents <= 0 {
		return ErrInvalidOperation
	}
	return u.transaction.InTx(ctx, func(txCtx context.Context) error {
		for _, id := range []model.ID{req.Source, req.Destination} {
			resource := &ResourceRequest{
				AccountID: &req.AccountID,
				ID:        id,
			}
			exists, err := u.virtualAccountRepository.ExistByAccountID(txCtx, (*VirtualAccountExistByAccountIDRequest)(resource))
			if err != nil {
				zap.S().Errorw("check slash virtual account", "error", err)
				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}
		}
		source, err := u.virtualAccountRepository.FindByAccountID(txCtx, &VirtualAccountFindByAccountIDRequest{
			AccountID: &req.AccountID,
			ID:        req.Source,
		})
		if err != nil {
			zap.S().Errorw("find slash source account", "error", err)
			return ErrDatabaseOperation
		}
		destination, err := u.virtualAccountRepository.FindByAccountID(txCtx, &VirtualAccountFindByAccountIDRequest{
			AccountID: &req.AccountID,
			ID:        req.Destination,
		})
		if err != nil {
			zap.S().Errorw("find slash destination account", "error", err)
			return ErrDatabaseOperation
		}
		first, second := source.WalletID, destination.WalletID
		if first > second {
			first, second = second, first
		}
		locked := make(map[model.ID]*model.Wallet, 2)
		for _, id := range []model.ID{first, second} {
			wallet, err := u.walletRepository.FindByAccountIDForUpdate(txCtx, &WalletFindByAccountIDForUpdateRequest{
				AccountID: &req.AccountID,
				ID:        id,
			})
			if err != nil {
				zap.S().Errorw("lock slash virtual account wallet", "error", err)
				return ErrDatabaseOperation
			}
			locked[id] = wallet
		}
		amount := decimal.NewFromInt(req.AmountCents).Div(decimal.NewFromInt(100))
		sourceWallet, destinationWallet := locked[source.WalletID], locked[destination.WalletID]
		if sourceWallet.Amount.LessThan(amount) {
			return ErrInvalidOperation
		}
		sourceWallet.Amount = sourceWallet.Amount.Sub(amount)
		destinationWallet.Amount = destinationWallet.Amount.Add(amount)
		if err := u.walletRepository.Save(txCtx, sourceWallet); err != nil {
			zap.S().Errorw("save slash source wallet", "error", err)
			return ErrDatabaseOperation
		}
		if err := u.walletRepository.Save(txCtx, destinationWallet); err != nil {
			zap.S().Errorw("save slash destination wallet", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
}
