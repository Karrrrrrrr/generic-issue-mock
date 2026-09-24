package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"

	"go.uber.org/zap"
)

type OpenAPICreateAccountRequest struct{ Name string }
type OpenAPIUpdateAccountRequest struct {
	ID   model.ID
	Name string
}

func (usecase *PayndaOpenAPIUsecase) CreateAccount(ctx context.Context, req *OpenAPICreateAccountRequest) (*model.Account, error) {
	account := &model.Account{
		Channel: enums.Channel_Paynda,
		Name:    req.Name,
	}
	err := usecase.transaction.InTx(ctx, func(ctx context.Context) error {
		if err := usecase.accountRepository.Create(ctx, account); err != nil {
			zap.S().Errorw("create paynda OpenAPI account", "error", err)
			return ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: account.ID,
			Channel:   account.Channel,
			Currency:  enums.Currency_USD,
			Type:      enums.WalletType_Account,
		}
		if err := usecase.walletRepository.Create(ctx, wallet); err != nil {
			zap.S().Errorw("create paynda OpenAPI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		account.WalletID = wallet.ID
		if err := usecase.accountRepository.Save(ctx, account); err != nil {
			zap.S().Errorw("attach paynda OpenAPI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	return account, err
}

func (usecase *PayndaOpenAPIUsecase) UpdateAccount(ctx context.Context, req *OpenAPIUpdateAccountRequest) (*model.Account, error) {
	var account *model.Account
	err := usecase.transaction.InTx(ctx, func(ctx context.Context) error {
		item, err := usecase.GetAccountWallet(ctx, req.ID)
		if err != nil {
			return err
		}
		account = item.Account
		account.Name = req.Name
		if err := usecase.accountRepository.Save(ctx, account); err != nil {
			zap.S().Errorw("rename paynda OpenAPI account", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	return account, err
}
