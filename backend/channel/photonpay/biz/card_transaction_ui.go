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

type ListUITransactionsRequest struct {
	AccountID       *model.ID
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
	Offset          int
	Limit           int
	ID              *model.ID
	CardID          *model.ID
	AuthorizationID *model.ID
	Types           []enums.CardTransactionType
	Statuses        []enums.CardTransactionStatus
}

type UISimulateRefundRequest struct {
	AuthorizationID *model.ID
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

type UIApplyTransactionStepRequest struct {
	CardTransactionID model.ID
	Type              enums.CardTransactionType
	Amount            *decimal.Decimal
}

func (u *PhotonPayUIUsecase) ListTransactions(ctx context.Context, req *ListUITransactionsRequest) ([]*model.CardTransaction, int64, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, 0, ErrInvalidOperation
	}
	transactions, err := u.cardTransactionRepo.ListTransactions(ctx, &CardTransactionListTransactionsRequest{
		AccountIDs:       types.PointerSlice(req.AccountID),
		IDs:              types.PointerSlice(req.ID),
		Statuses:         req.Statuses,
		CreatedFrom:      req.CreatedFrom,
		CreatedTo:        req.CreatedTo,
		CardIDs:          types.PointerSlice(req.CardID),
		AuthorizationIDs: types.PointerSlice(req.AuthorizationID),
		Types:            req.Types,
		Limit:            req.Limit,
		Offset:           req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay UI card transactions", "error", err)

		return nil, 0, ErrDatabaseOperation
	}

	total, err := u.cardTransactionRepo.Count(ctx, &CardTransactionCountRequest{
		AccountIDs:       types.PointerSlice(req.AccountID),
		IDs:              types.PointerSlice(req.ID),
		Statuses:         req.Statuses,
		CreatedFrom:      req.CreatedFrom,
		CreatedTo:        req.CreatedTo,
		CardIDs:          types.PointerSlice(req.CardID),
		AuthorizationIDs: types.PointerSlice(req.AuthorizationID),
		Types:            req.Types,
	})
	if err != nil {
		zap.S().Errorw("count photonpay UI card transactions", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return transactions, total, nil
}

// SimulateRefund creates a posted refund directly for an active card.
func (u *PhotonPayUIUsecase) SimulateRefund(ctx context.Context, req *UISimulateRefundRequest) (*model.CardTransaction, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var transaction *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.getCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}
		var authorizationID model.ID
		if req.AuthorizationID != nil && *req.AuthorizationID != 0 {
			if *req.AuthorizationID < 0 {
				return ErrInvalidOperation
			}
			exists, err := u.authorizationRepo.AuthorizationExists(txCtx, &ExistAuthorizationRequest{
				AccountID: card.AccountID,
				ID:        *req.AuthorizationID,
			})
			if err != nil {
				zap.S().Errorw("check photonpay refund authorization", "error", err)
				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}
			auth, err := u.authorizationRepo.LockAuthorization(txCtx, &LockAuthorizationRequest{
				AccountID: card.AccountID,
				ID:        *req.AuthorizationID,
			})
			if err != nil {
				zap.S().Errorw("find photonpay refund authorization", "error", err)
				return ErrDatabaseOperation
			}
			if auth.CardID != card.ID || auth.Currency != req.Currency {
				return ErrInvalidOperation
			}
			authorizationID = auth.ID
		}
		wallet, err := u.walletRepo.LockWallet(txCtx, &LockWalletRequest{
			AccountID: card.AccountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay refund wallet", "error", err)
			return ErrDatabaseOperation
		}
		if wallet.Currency != req.Currency {
			return ErrInvalidOperation
		}
		wallet.Available = wallet.Available.Add(req.Amount)
		wallet.In = wallet.In.Add(req.Amount)
		if err := u.walletRepo.SaveWallet(txCtx, wallet); err != nil {
			zap.S().Errorw("credit photonpay refund wallet", "error", err)
			return ErrDatabaseOperation
		}

		transaction = &model.CardTransaction{
			AuthorizationID:   authorizationID,
			Account:           card.Account,
			AccountID:         card.AccountID,
			Channel:           enums.Channel_PhotonPay,
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
		if err := u.cardTransactionRepo.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create photonpay UI simulated refund", "error", err)
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

func (u *PhotonPayUIUsecase) ApplyTransactionStep(ctx context.Context, req *UIApplyTransactionStepRequest) (*model.CardTransaction, error) {
	var next *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		origin, err := u.getCardTransaction(txCtx, req.CardTransactionID)
		if err != nil {
			return err
		}

		amount := origin.TxAmount
		if req.Amount != nil {
			amount = *req.Amount
		}
		if !amount.IsPositive() {
			return ErrInvalidOperation
		}

		if origin.AuthorizationID <= 0 {
			return ErrInvalidOperation
		}
		existsAuth, err := u.authorizationRepo.AuthorizationExists(txCtx, &ExistAuthorizationRequest{
			AccountID: origin.AccountID,
			ID:        origin.AuthorizationID,
		})
		if err != nil {
			zap.S().Errorw("check photonpay transaction authorization", "error", err)
			return ErrDatabaseOperation
		}
		if !existsAuth {
			return ErrResourceNotFound
		}
		auth, err := u.authorizationRepo.LockAuthorization(txCtx, &LockAuthorizationRequest{
			AccountID: origin.AccountID,
			ID:        origin.AuthorizationID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay transaction authorization", "error", err)
			return ErrDatabaseOperation
		}
		if auth.CardID != origin.CardID || auth.Currency != origin.Currency {
			return ErrInvalidOperation
		}
		status := enums.TransactionStatus_SUCCEED
		if req.Type == enums.CardTransactionType_VOID {
			status = enums.TransactionStatus_VOID
		}

		if req.Type != enums.CardTransactionType_VOID && req.Type != enums.CardTransactionType_REFUND && req.Type != enums.CardTransactionType_CLEAR {
			return ErrInvalidOperation
		}
		if req.Type == enums.CardTransactionType_VOID && origin.Type != enums.CardTransactionType_AUTH {
			return ErrInvalidOperation
		}
		if req.Type != enums.CardTransactionType_VOID {
			card, err := u.cardRepo.FindCard(txCtx, &FindCardRequest{
				AccountID: origin.AccountID,
				ID:        origin.CardID,
			})
			if err != nil {
				zap.S().Errorw("find photonpay transaction card", "error", err)
				return ErrDatabaseOperation
			}
			wallet, err := u.walletRepo.LockWallet(txCtx, &LockWalletRequest{
				AccountID: origin.AccountID,
				ID:        card.WalletID,
			})
			if err != nil {
				zap.S().Errorw("lock photonpay transaction wallet", "error", err)
				return ErrDatabaseOperation
			}
			if wallet.Currency != origin.Currency {
				return ErrInvalidOperation
			}
			if req.Type == enums.CardTransactionType_REFUND {
				wallet.Available = wallet.Available.Add(amount)
				wallet.In = wallet.In.Add(amount)
			} else {
				wallet.Available = wallet.Available.Sub(amount)
				wallet.Out = wallet.Out.Add(amount)
			}
			if err := u.walletRepo.SaveWallet(txCtx, wallet); err != nil {
				zap.S().Errorw("save photonpay transaction wallet", "error", err)
				return ErrDatabaseOperation
			}
		}
		next = &model.CardTransaction{
			Account:                 origin.Account,
			AccountID:               origin.AccountID,
			Channel:                 enums.Channel_PhotonPay,
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
		if err := u.cardTransactionRepo.Create(txCtx, next); err != nil {
			zap.S().Errorw("create photonpay UI card transaction step", "error", err)

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

func (u *PhotonPayUIUsecase) getCardTransaction(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	exists, err := u.cardTransactionRepo.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay UI card transaction", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	transaction, err := u.cardTransactionRepo.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find photonpay UI card transaction", "error", err)

		return nil, ErrDatabaseOperation
	}

	return transaction, nil
}
