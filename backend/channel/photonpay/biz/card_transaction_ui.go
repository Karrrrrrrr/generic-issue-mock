package biz

import (
	"context"
	"time"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharedbiz "generic-mock/shared/biz"

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
	CardID          *model.ID
	Amount          decimal.Decimal
	Currency        *enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
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
		return nil, 0, photonpayerrors.ErrInvalidOperation
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

		return nil, 0, photonpayerrors.ErrDatabaseOperation
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
		return nil, 0, photonpayerrors.ErrDatabaseOperation
	}

	return transactions, total, nil
}

func (u *PhotonPayUIUsecase) SimulateRefund(ctx context.Context, req *UISimulateRefundRequest) (*model.CardTransaction, error) {
	if req == nil {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	result, err := u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
		Channel:         enums.Channel_PhotonPay,
		CardID:          req.CardID,
		AuthorizationID: req.AuthorizationID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
		Notificator:     u,
	})
	if err != nil {
		return nil, photonpayerrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}

func (u *PhotonPayUIUsecase) ApplyTransactionStep(ctx context.Context, req *UIApplyTransactionStepRequest) (*model.CardTransaction, error) {
	if req == nil || req.CardTransactionID <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	exists, err := u.cardTransactionRepo.ExistByID(ctx, req.CardTransactionID)
	if err != nil {
		zap.S().Errorw("check photonpay simulation transaction", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}
	origin, err := u.cardTransactionRepo.FindByID(ctx, req.CardTransactionID)
	if err != nil {
		zap.S().Errorw("find photonpay simulation transaction", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	amount := origin.TxAmount
	if req.Amount != nil {
		amount = *req.Amount
	}
	if origin.AuthorizationID <= 0 || !amount.IsPositive() {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	if (req.Type == enums.CardTransactionType_CLEAR || req.Type == enums.CardTransactionType_VOID) &&
		(origin.Type != enums.CardTransactionType_AUTH || origin.Status != enums.TransactionStatus_AUTHORIZED) {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	var result *sharedbiz.CardTransactionSimulationResult
	switch req.Type {
	case enums.CardTransactionType_CLEAR:
		result, err = u.simulator.SimulateClearing(ctx, &sharedbiz.SimulateClearingReq{
			Channel:         enums.Channel_PhotonPay,
			Amount:          amount,
			Notificator:     u,
			AuthorizationID: origin.AuthorizationID,
		})
	case enums.CardTransactionType_VOID:
		result, err = u.simulator.SimulateReversal(ctx, &sharedbiz.SimulateReversalReq{
			Channel:         enums.Channel_PhotonPay,
			Amount:          amount,
			Notificator:     u,
			AuthorizationID: origin.AuthorizationID,
			Status:          enums.TransactionStatus_VOID,
		})
	case enums.CardTransactionType_REFUND:
		result, err = u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
			Channel:         enums.Channel_PhotonPay,
			Amount:          amount,
			Notificator:     u,
			AuthorizationID: &origin.AuthorizationID,
		})
	default:
		return nil, photonpayerrors.ErrInvalidOperation
	}
	if err != nil {
		return nil, photonpayerrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}

func (u *PhotonPayUIUsecase) getCardTransaction(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	exists, err := u.cardTransactionRepo.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay UI card transaction", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}

	transaction, err := u.cardTransactionRepo.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find photonpay UI card transaction", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return transaction, nil
}
