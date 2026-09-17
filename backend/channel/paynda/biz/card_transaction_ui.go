package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type PayndaListTransactionsRequest struct {
	PayndaListRequest
	CardID         *model.ID
	StartCreatedAt *time.Time
	EndCreatedAt   *time.Time
	Types          []enums.CardTransactionType
}

type PayndaSimulateRefundRequest struct {
	AuthorizationID *model.ID
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

type PayndaUIApplyTransactionStepRequest struct {
	CardTransactionID model.ID
	Type              enums.CardTransactionType
	Amount            *decimal.Decimal
}

func (u *PayndaUIUsecase) ListTransactions(ctx context.Context, req *PayndaListRequest) ([]*model.CardTransaction, error) {
	items, err := u.cardTransactionRepository.List(ctx, &CardTransactionListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Offset:     req.Offset,
		Limit:      req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list paynda UI transactions", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

// SimulateRefund creates a posted refund directly for an active card.
func (u *PayndaUIUsecase) SimulateRefund(ctx context.Context, req *PayndaSimulateRefundRequest) (*model.CardTransaction, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var transaction *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistByID(txCtx, &CardExistByIDRequest{ID: req.CardID})
		if err != nil {
			zap.S().Errorw("check paynda UI card for simulated refund", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		card, err := u.cardRepository.FindByID(txCtx, &CardFindByIDRequest{ID: req.CardID})
		if err != nil {
			zap.S().Errorw("find paynda UI card for simulated refund", "error", err)
			return ErrDatabaseOperation
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}
		var authorizationID model.ID
		if req.AuthorizationID != nil && *req.AuthorizationID != 0 {
			if *req.AuthorizationID < 0 {
				return ErrInvalidOperation
			}
			exists, err := u.authorizationRepository.AuthorizationExists(txCtx, &ExistAuthorizationRequest{
				AccountID: card.AccountID,
				ID:        *req.AuthorizationID,
			})
			if err != nil {
				zap.S().Errorw("check paynda refund authorization", "error", err)
				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}
			auth, err := u.authorizationRepository.LockAuthorization(txCtx, &LockAuthorizationRequest{
				AccountID: card.AccountID,
				ID:        *req.AuthorizationID,
			})
			if err != nil {
				zap.S().Errorw("find paynda refund authorization", "error", err)
				return ErrDatabaseOperation
			}
			if auth.CardID != card.ID || auth.Currency != req.Currency {
				return ErrInvalidOperation
			}
			authorizationID = auth.ID
		}
		wallet, err := u.walletRepository.LockWallet(txCtx, &LockWalletRequest{
			AccountID: card.AccountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock paynda refund wallet", "error", err)
			return ErrDatabaseOperation
		}
		if wallet.Currency != req.Currency {
			return ErrInvalidOperation
		}
		wallet.Amount = wallet.Amount.Add(req.Amount)
		wallet.In = wallet.In.Add(req.Amount)
		if err := u.walletRepository.SaveWallet(txCtx, wallet); err != nil {
			zap.S().Errorw("credit paynda refund wallet", "error", err)
			return ErrDatabaseOperation
		}

		transaction = &model.CardTransaction{
			AuthorizationID:   authorizationID,
			Account:           card.Account,
			AccountID:         card.AccountID,
			Channel:           enums.Channel_Paynda,
			CardID:            card.ID,
			Status:            enums.TransactionStatus_SUCCEED,
			Type:              enums.CardTransactionType_REFUND,
			Currency:          req.Currency,
			TxAmount:          req.Amount,
			TxCurrency:        req.Currency,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: randomx.Digits(6),
		}
		if err := u.cardTransactionRepository.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create paynda UI simulated refund", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	u.dispatchTransaction(ctx, transaction)
	return transaction, nil
}

func (u *PayndaUIUsecase) ApplyTransactionStep(ctx context.Context, req *PayndaUIApplyTransactionStepRequest) (*model.CardTransaction, error) {
	var next *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardTransactionRepository.ExistByID(txCtx, req.CardTransactionID)
		if err != nil {
			zap.S().Errorw("check paynda UI transaction", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		origin, err := u.cardTransactionRepository.FindByID(txCtx, req.CardTransactionID)
		if err != nil {
			zap.S().Errorw("find paynda UI transaction", "error", err)
			return ErrDatabaseOperation
		}
		amount := origin.TxAmount
		if req.Amount != nil {
			amount = *req.Amount
		}
		if !amount.IsPositive() {
			return ErrInvalidOperation
		}
		if origin.CardID == 0 {
			return ErrInvalidOperation
		}

		card, err := u.cardRepository.FindByID(txCtx, &CardFindByIDRequest{
			AccountID: &origin.AccountID,
			ID:        origin.CardID,
		})
		if err != nil {
			zap.S().Errorw("find paynda UI transaction card", "error", err)
			return ErrDatabaseOperation
		}
		if card.WalletID == 0 {
			return ErrResourceNotFound
		}
		wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, &WalletFindByIDForUpdateRequest{
			AccountID: &card.AccountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock paynda UI card wallet", "error", err)
			return ErrDatabaseOperation
		}

		if origin.AuthorizationID <= 0 {
			return ErrInvalidOperation
		}
		existsAuth, err := u.authorizationRepository.AuthorizationExists(txCtx, &ExistAuthorizationRequest{
			AccountID: origin.AccountID,
			ID:        origin.AuthorizationID,
		})
		if err != nil {
			zap.S().Errorw("check paynda transaction authorization", "error", err)
			return ErrDatabaseOperation
		}
		if !existsAuth {
			return ErrResourceNotFound
		}
		auth, err := u.authorizationRepository.LockAuthorization(txCtx, &LockAuthorizationRequest{
			AccountID: origin.AccountID,
			ID:        origin.AuthorizationID,
		})
		if err != nil {
			zap.S().Errorw("lock paynda transaction authorization", "error", err)
			return ErrDatabaseOperation
		}
		if auth.CardID != origin.CardID || auth.Currency != origin.Currency {
			return ErrInvalidOperation
		}
		status := enums.TransactionStatus_SUCCEED
		switch req.Type {
		case enums.CardTransactionType_CLEAR:
			if origin.Type != enums.CardTransactionType_AUTH || origin.Status != enums.TransactionStatus_AUTHORIZED {
				return ErrInvalidOperation
			}
			wallet.Amount = wallet.Amount.Sub(amount)
			wallet.Out = wallet.Out.Add(amount)
			origin.Status = enums.TransactionStatus_SUCCEED
		case enums.CardTransactionType_VOID:
			if origin.Type != enums.CardTransactionType_AUTH || origin.Status != enums.TransactionStatus_AUTHORIZED {
				return ErrInvalidOperation
			}
			status = enums.TransactionStatus_VOID
			origin.Status = enums.TransactionStatus_VOID
		case enums.CardTransactionType_REFUND:
			if origin.AuthorizationID <= 0 {
				return ErrInvalidOperation
			}
			wallet.Amount = wallet.Amount.Add(amount)
			wallet.In = wallet.In.Add(amount)
		default:
			return ErrInvalidOperation
		}
		if err := u.walletRepository.Save(txCtx, wallet); err != nil {
			zap.S().Errorw("save paynda UI card wallet", "error", err)
			return ErrDatabaseOperation
		}
		if err := u.cardTransactionRepository.Save(txCtx, origin); err != nil {
			zap.S().Errorw("update paynda UI origin transaction", "error", err)
			return ErrDatabaseOperation
		}
		next = &model.CardTransaction{
			Account:                 origin.Account,
			AccountID:               origin.AccountID,
			Channel:                 enums.Channel_Paynda,
			OriginCardTransactionID: origin.ID,
			AuthorizationID:         origin.AuthorizationID,
			CardID:                  origin.CardID,
			Status:                  status,
			Type:                    req.Type,
			Currency:                origin.Currency,
			TxAmount:                amount,
			TxCurrency:              origin.TxCurrency,
			MerchantName:            origin.MerchantName,
			MerchantCountry:         origin.MerchantCountry,
			MerchantMCC:             origin.MerchantMCC,
			AuthorizationCode:       origin.AuthorizationCode,
		}
		if err := u.cardTransactionRepository.Create(txCtx, next); err != nil {
			zap.S().Errorw("create paynda UI transaction step", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	u.dispatchTransaction(ctx, next)
	return next, nil
}
