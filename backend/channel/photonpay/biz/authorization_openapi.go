package biz

import (
	"context"

	photon "generic-mock/channel/photonpay/enums"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	sharedbiz "generic-mock/shared/biz"

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

func (req *SandboxTransactionRequest) Validate() error {
	if req == nil || req.AccountID <= 0 || req.CardID <= 0 || req.RequestID == "" ||
		!req.Amount.IsPositive() || req.Currency == "" || req.MerchantName == "" ||
		req.MerchantCountry == "" || req.MerchantMCC == "" ||
		(req.OriginTransactionID != nil && *req.OriginTransactionID <= 0) {
		return photonpayerrors.ErrInvalidOperation
	}
	switch req.Type {
	case photon.SandboxTransactionType_Auth:
		if req.OriginTransactionID != nil {
			return photonpayerrors.ErrInvalidOperation
		}
	case photon.SandboxTransactionType_Void:
		if req.OriginTransactionID == nil {
			return photonpayerrors.ErrInvalidOperation
		}
	case photon.SandboxTransactionType_Refund:
	default:
		return photonpayerrors.ErrInvalidOperation
	}
	return nil
}

func (u *PhotonPayOpenAPIUsecase) SandboxTransaction(ctx context.Context, req *SandboxTransactionRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	var authorizationID *model.ID
	if req.OriginTransactionID != nil {
		exists, err := u.cardTransactionRepo.ExistByAccountID(ctx, &CardTransactionExistByAccountIDRequest{
			AccountID: &req.AccountID,
			ID:        *req.OriginTransactionID,
		})
		if err != nil {
			zap.S().Errorw("check photonpay sandbox origin transaction", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		if !exists {
			return photonpayerrors.ErrResourceNotFound
		}
		origin, err := u.cardTransactionRepo.FindByAccountID(ctx, &CardTransactionFindByAccountIDRequest{
			AccountID: &req.AccountID,
			ID:        *req.OriginTransactionID,
		})
		if err != nil {
			zap.S().Errorw("find photonpay sandbox origin transaction", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		if origin.CardID != req.CardID || origin.Currency != req.Currency || origin.AuthorizationID <= 0 ||
			origin.Type != common.CardTransactionType_AUTH || origin.Status != common.TransactionStatus_AUTHORIZED {
			return photonpayerrors.ErrInvalidOperation
		}
		authorizationID = &origin.AuthorizationID
	}
	var err error
	switch req.Type {
	case photon.SandboxTransactionType_Auth:
		_, err = u.simulator.SimulateAuthorization(ctx, &sharedbiz.SimulateAuthorizationReq{
			AccountID:       req.AccountID,
			Channel:         common.Channel_PhotonPay,
			CardID:          req.CardID,
			Amount:          req.Amount,
			Currency:        req.Currency,
			RequestID:       &req.RequestID,
			MerchantName:    &req.MerchantName,
			MerchantCountry: &req.MerchantCountry,
			MerchantMCC:     &req.MerchantMCC,
			Notificator:     sharedbiz.NoopNotificator{},
		})
	case photon.SandboxTransactionType_Void:
		_, err = u.simulator.SimulateReversal(ctx, &sharedbiz.SimulateReversalReq{
			AccountID:       req.AccountID,
			Channel:         common.Channel_PhotonPay,
			CardID:          req.CardID,
			AuthorizationID: *authorizationID,
			Amount:          req.Amount,
			Status:          common.TransactionStatus_VOID,
			RequestID:       &req.RequestID,
			Notificator:     sharedbiz.NoopNotificator{},
		})
	case photon.SandboxTransactionType_Refund:
		_, err = u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
			AccountID:       req.AccountID,
			Channel:         common.Channel_PhotonPay,
			CardID:          req.CardID,
			AuthorizationID: authorizationID,
			Amount:          req.Amount,
			Currency:        req.Currency,
			RequestID:       &req.RequestID,
			MerchantName:    &req.MerchantName,
			MerchantCountry: &req.MerchantCountry,
			MerchantMCC:     &req.MerchantMCC,
			Notificator:     sharedbiz.NoopNotificator{},
		})
	}
	return photonpayerrors.FromSimulation(err)
}
