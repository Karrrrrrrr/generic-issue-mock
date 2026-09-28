package biz

import (
	"context"
	"time"

	payndaerrors "generic-mock/channel/paynda/errors"
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

type PayndaListTransactionsRequest struct {
	PayndaListRequest
	CardID         *model.ID
	StartCreatedAt *time.Time
	EndCreatedAt   *time.Time
	Types          []enums.CardTransactionType
}

type PayndaSimulateRefundRequest struct {
	AuthorizationID *model.ID
	CardID          *model.ID
	Amount          decimal.Decimal
	Currency        *enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
}

type PayndaUIApplyTransactionStepRequest struct {
	CardTransactionID model.ID
	Type              enums.CardTransactionType
	Amount            *decimal.Decimal
}

func (u *PayndaUIUsecase) ListTransactions(ctx context.Context, req *ListUITransactionsRequest) ([]*model.CardTransaction, int64, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, 0, payndaerrors.ErrInvalidOperation
	}
	items, err := u.cardTransactionRepository.List(ctx, &CardTransactionListRequest{
		AccountIDs:       types.PointerSlice(req.AccountID),
		IDs:              types.PointerSlice(req.ID),
		Statuses:         req.Statuses,
		CreatedFrom:      req.CreatedFrom,
		CreatedTo:        req.CreatedTo,
		CardIDs:          types.PointerSlice(req.CardID),
		AuthorizationIDs: types.PointerSlice(req.AuthorizationID),
		Types:            req.Types,
		Offset:           req.Offset,
		Limit:            req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list paynda UI transactions", "error", err)
		return nil, 0, payndaerrors.ErrDatabaseOperation
	}
	total, err := u.cardTransactionRepository.Count(ctx, &CardTransactionCountRequest{
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
		zap.S().Errorw("count paynda UI card transactions", "error", err)
		return nil, 0, payndaerrors.ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *PayndaUIUsecase) SimulateRefund(ctx context.Context, req *PayndaSimulateRefundRequest) (*model.CardTransaction, error) {
	if req == nil {
		return nil, payndaerrors.ErrInvalidOperation
	}
	result, err := u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
		Channel:         enums.Channel_Paynda,
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
		return nil, payndaerrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}

func (u *PayndaUIUsecase) ApplyTransactionStep(ctx context.Context, req *PayndaUIApplyTransactionStepRequest) (*model.CardTransaction, error) {
	if req == nil || req.CardTransactionID <= 0 {
		return nil, payndaerrors.ErrInvalidOperation
	}
	exists, err := u.cardTransactionRepository.ExistByID(ctx, req.CardTransactionID)
	if err != nil {
		zap.S().Errorw("check paynda simulation transaction", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, payndaerrors.ErrResourceNotFound
	}
	origin, err := u.cardTransactionRepository.FindByID(ctx, req.CardTransactionID)
	if err != nil {
		zap.S().Errorw("find paynda simulation transaction", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	amount := origin.TxAmount
	if req.Amount != nil {
		amount = *req.Amount
	}
	if origin.AuthorizationID <= 0 || !amount.IsPositive() {
		return nil, payndaerrors.ErrInvalidOperation
	}
	if (req.Type == enums.CardTransactionType_CLEAR || req.Type == enums.CardTransactionType_VOID) &&
		(origin.Type != enums.CardTransactionType_AUTH || origin.Status != enums.TransactionStatus_AUTHORIZED) {
		return nil, payndaerrors.ErrInvalidOperation
	}
	var result *sharedbiz.CardTransactionSimulationResult
	switch req.Type {
	case enums.CardTransactionType_CLEAR:
		result, err = u.simulator.SimulateClearing(ctx, &sharedbiz.SimulateClearingReq{
			Channel:         enums.Channel_Paynda,
			Amount:          amount,
			Notificator:     u,
			AuthorizationID: origin.AuthorizationID,
		})
	case enums.CardTransactionType_VOID:
		result, err = u.simulator.SimulateReversal(ctx, &sharedbiz.SimulateReversalReq{
			Channel:         enums.Channel_Paynda,
			Amount:          amount,
			Notificator:     u,
			AuthorizationID: origin.AuthorizationID,
			Status:          enums.TransactionStatus_VOID,
		})
	case enums.CardTransactionType_REFUND:
		result, err = u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
			Channel:         enums.Channel_Paynda,
			Amount:          amount,
			Notificator:     u,
			AuthorizationID: &origin.AuthorizationID,
		})
	default:
		return nil, payndaerrors.ErrInvalidOperation
	}
	if err != nil {
		return nil, payndaerrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}
