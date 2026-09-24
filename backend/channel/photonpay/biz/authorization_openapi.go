package biz

import (
	"context"

	photon "generic-mock/channel/photonpay/enums"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/randomx"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type SandboxTransactionRequest struct {
	AccountID           model.ID
	RequestID           string
	CardID              model.ID
	OriginTransactionID *model.ID
	Currency            common.Currency
	Amount              decimal.Decimal
	Type                photon.SandboxTransactionType
	MerchantName        string
	MerchantCountry     string
	MerchantMCC         string
}

func (u *PhotonPayOpenAPIUsecase) SandboxTransaction(ctx context.Context, req *SandboxTransactionRequest) error {
	if !req.Amount.IsPositive() {
		return ErrInvalidOperation
	}
	if req.Type != photon.SandboxTransactionType_Auth && req.Type != photon.SandboxTransactionType_Void && req.Type != photon.SandboxTransactionType_Refund {
		return ErrInvalidOperation
	}
	var originTransactionID model.ID
	if req.OriginTransactionID != nil {
		originTransactionID = *req.OriginTransactionID
		if originTransactionID < 0 {
			return ErrInvalidOperation
		}
	}
	if req.Type == photon.SandboxTransactionType_Void && originTransactionID == 0 {
		return ErrInvalidOperation
	}
	if req.Type == photon.SandboxTransactionType_Auth && originTransactionID != 0 {
		return ErrInvalidOperation
	}
	return u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCard(txCtx, &ResourceRequest{
			AccountID: &req.AccountID,
			ID:        req.CardID,
		}); err != nil {
			return err
		}
		card, err := u.cardRepo.FindCardByAccountID(txCtx, &CardFindCardByAccountIDRequest{
			AccountID: &req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("find photonpay sandbox card", "error", err)
			return ErrDatabaseOperation
		}

		transactionType := photon.SandboxTransactionTypeToGeneric(req.Type)
		var originTransaction *model.CardTransaction
		if originTransactionID != 0 {
			exists, err := u.cardTransactionRepo.ExistByAccountID(txCtx, &CardTransactionExistByAccountIDRequest{
				AccountID: &req.AccountID,
				ID:        originTransactionID,
			})
			if err != nil {
				zap.S().Errorw("check photonpay origin transaction", "error", err)

				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}

			originTransaction, err = u.cardTransactionRepo.FindByAccountID(txCtx, &CardTransactionFindByAccountIDRequest{
				AccountID: &req.AccountID,
				ID:        originTransactionID,
			})
			if err != nil {
				zap.S().Errorw("find photonpay origin transaction", "error", err)

				return ErrDatabaseOperation
			}
			if originTransaction.CardID != card.ID || originTransaction.Type != common.CardTransactionType_AUTH || originTransaction.AuthorizationID == 0 {
				return ErrInvalidOperation
			}
			exists, err = u.authorizationRepo.AuthorizationExists(txCtx, &ExistAuthorizationRequest{
				AccountID: req.AccountID,
				ID:        originTransaction.AuthorizationID,
			})
			if err != nil {
				zap.S().Errorw("check photonpay sandbox authorization", "error", err)
				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}
		}

		var authorizationID model.ID
		if transactionType == common.CardTransactionType_AUTH {
			authorization := &model.Authorization{
				AccountID:         card.AccountID,
				Channel:           common.Channel_PhotonPay,
				CardID:            req.CardID,
				Currency:          req.Currency,
				Amount:            req.Amount,
				MerchantName:      req.MerchantName,
				MerchantCountry:   req.MerchantCountry,
				MerchantMCC:       req.MerchantMCC,
				AuthorizationCode: randomx.Digits(6),
				Status:            common.TransactionStatus_AUTHORIZED,
			}
			if err := u.authorizationRepo.Create(txCtx, authorization); err != nil {
				zap.S().Errorw("create photonpay authorization", "error", err)

				return ErrDatabaseOperation
			}

			authorizationID = authorization.ID
		} else if originTransaction != nil {
			authorizationID = originTransaction.AuthorizationID
		}

		transaction := &model.CardTransaction{
			AccountID:               card.AccountID,
			Channel:                 common.Channel_PhotonPay,
			OriginCardTransactionID: originTransactionID,
			AuthorizationID:         authorizationID,
			CardID:                  req.CardID,
			Status:                  common.TransactionStatus_SUCCEED,
			Type:                    transactionType,
			Currency:                req.Currency,
			TxAmount:                req.Amount,
			TxCurrency:              req.Currency,
			RequestID:               req.RequestID,
			MerchantName:            req.MerchantName,
			MerchantCountry:         req.MerchantCountry,
			MerchantMCC:             req.MerchantMCC,
		}
		if err := u.cardTransactionRepo.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create photonpay sandbox transaction", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
}
