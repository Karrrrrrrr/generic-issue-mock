package biz

import (
	"context"

	paynda "generic-mock/channel/paynda/enums"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type PayndaAccountWallet struct {
	Account *model.Account
	Wallet  *model.Wallet
}

type PayndaAccountWalletTransferRequest struct {
	AccountID model.ID
	Amount    decimal.Decimal
}

type PayndaTransferRequest struct {
	AccountID model.ID
	CardID    model.ID
	RequestID string
	Amount    decimal.Decimal
	Type      paynda.TransferType
}

func (u *PayndaOpenAPIUsecase) GetAccountWallet(
	ctx context.Context,
	accountID model.ID,
) (*PayndaAccountWallet, error) {
	exists, err := u.accountRepository.ExistByID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("check paynda account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	account, err := u.accountRepository.FindByID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("find paynda account", "error", err)
		return nil, ErrDatabaseOperation
	}
	wallet, err := u.walletRepository.FindByID(ctx, &WalletFindByIDRequest{
		AccountID: &accountID,
		ID:        account.WalletID,
	})
	if err != nil {
		zap.S().Errorw("find paynda account wallet", "error", err)
		return nil, ErrDatabaseOperation
	}

	return &PayndaAccountWallet{
		Account: account,
		Wallet:  wallet,
	}, nil
}

func (u *PayndaOpenAPIUsecase) TransferAccountWallet(
	ctx context.Context,
	req *PayndaAccountWalletTransferRequest,
) (*PayndaAccountWallet, error) {
	accountWallet, err := u.GetAccountWallet(ctx, req.AccountID)
	if err != nil {
		return nil, err
	}

	err = u.transaction.InTx(ctx, func(txCtx context.Context) error {
		wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, &WalletFindByIDForUpdateRequest{
			AccountID: &accountWallet.Account.ID,
			ID:        accountWallet.Wallet.ID,
		})
		if err != nil {
			zap.S().Errorw("lock paynda account wallet", "error", err)
			return ErrDatabaseOperation
		}
		wallet.Available = wallet.Available.Add(req.Amount)
		wallet.In = wallet.In.Add(req.Amount)
		if err := u.walletRepository.Save(txCtx, wallet); err != nil {
			zap.S().Errorw("save paynda account wallet", "error", err)
			return ErrDatabaseOperation
		}
		accountWallet.Wallet = wallet

		return nil
	})
	if err != nil {
		return nil, err
	}

	return accountWallet, nil
}

func (u *PayndaOpenAPIUsecase) TransferCardBalance(
	ctx context.Context,
	req *PayndaTransferRequest,
) (*model.CardTransaction, error) {
	var transaction *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.GetCard(txCtx, &PayndaResourceRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			return err
		}
		if card.WalletID == 0 {
			return ErrResourceNotFound
		}

		wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, &WalletFindByIDForUpdateRequest{
			AccountID: &req.AccountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock paynda card wallet", "error", err)
			return ErrDatabaseOperation
		}
		oldAmount := wallet.Available
		transactionType := enums.CardTransactionType_FundIn
		if req.Type == paynda.TransferType_Out {
			if wallet.Available.LessThan(req.Amount) {
				return ErrInvalidOperation
			}
			wallet.Available = wallet.Available.Sub(req.Amount)
			wallet.Out = wallet.Out.Add(req.Amount)
			transactionType = enums.CardTransactionType_FundOut
		} else {
			wallet.Available = wallet.Available.Add(req.Amount)
			wallet.In = wallet.In.Add(req.Amount)
		}
		if err := u.walletRepository.Save(txCtx, wallet); err != nil {
			zap.S().Errorw("save paynda card wallet", "error", err)
			return ErrDatabaseOperation
		}

		transaction = &model.CardTransaction{
			AccountID:  req.AccountID,
			Channel:    enums.Channel_Paynda,
			CardID:     card.ID,
			Status:     enums.TransactionStatus_SUCCEED,
			Type:       transactionType,
			Currency:   wallet.Currency,
			TxAmount:   req.Amount,
			TxCurrency: wallet.Currency,
			RequestID:  req.RequestID,
			RawPayload: []byte(oldAmount.String()),
		}
		if err := u.cardTransactionRepository.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create paynda card balance transfer", "error", err)
			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
