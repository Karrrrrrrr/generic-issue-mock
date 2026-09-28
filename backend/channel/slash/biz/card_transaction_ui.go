package biz

import (
	"context"
	"time"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharedbiz "generic-mock/shared/biz"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type ListCardTransactionsRequest struct {
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

type SimulateRefundRequest struct {
	AuthorizationID *model.ID
	CardID          *model.ID
	Amount          decimal.Decimal
	Currency        *enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
	Notificator     sharedbiz.CardTransactionNotificator
}

type ApplyTransactionStepRequest struct {
	CardTransactionID model.ID
	Type              enums.CardTransactionType
	Amount            *decimal.Decimal
	Notificator       sharedbiz.CardTransactionNotificator
}

func (u *SlashUIUsecase) SimulateRefund(ctx context.Context, req *SimulateRefundRequest) (*model.CardTransaction, error) {
	if req == nil {
		return nil, slasherrors.ErrInvalidOperation
	}
	result, err := u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
		Channel:         enums.Channel_Slash,
		CardID:          req.CardID,
		AuthorizationID: req.AuthorizationID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
		Notificator:     req.Notificator,
	})
	if err != nil {
		return nil, slasherrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}

func (u *SlashUIUsecase) ListCardTransactions(ctx context.Context, req *ListCardTransactionsRequest) ([]*model.CardTransaction, int64, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, 0, slasherrors.ErrInvalidOperation
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
		Limit:            req.Limit,
		Offset:           req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list slash card transactions", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
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
		Limit:            req.Limit,
		Offset:           req.Offset,
	})
	if err != nil {
		zap.S().Errorw("count slash card transactions", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *SlashUIUsecase) GetCardTransaction(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	if err := u.requireCardTransaction(ctx, id); err != nil {
		return nil, err
	}
	item, err := u.cardTransactionRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash card transaction", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}

	return item, nil
}

func (u *SlashUIUsecase) ApplyTransactionStep(ctx context.Context, req *ApplyTransactionStepRequest) (*model.CardTransaction, error) {
	if req == nil || req.CardTransactionID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	exists, err := u.cardTransactionRepository.ExistByID(ctx, req.CardTransactionID)
	if err != nil {
		zap.S().Errorw("check slash simulation transaction", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, slasherrors.ErrResourceNotFound
	}
	origin, err := u.cardTransactionRepository.FindByID(ctx, req.CardTransactionID)
	if err != nil {
		zap.S().Errorw("find slash simulation transaction", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	amount := origin.TxAmount
	if req.Amount != nil {
		amount = *req.Amount
	}
	if origin.AuthorizationID <= 0 || !amount.IsPositive() {
		return nil, slasherrors.ErrInvalidOperation
	}
	if (req.Type == enums.CardTransactionType_CLEAR || req.Type == enums.CardTransactionType_VOID) &&
		(origin.Type != enums.CardTransactionType_AUTH || origin.Status != enums.TransactionStatus_AUTHORIZED) {
		return nil, slasherrors.ErrInvalidOperation
	}
	var result *sharedbiz.CardTransactionSimulationResult
	switch req.Type {
	case enums.CardTransactionType_CLEAR:
		result, err = u.simulator.SimulateClearing(ctx, &sharedbiz.SimulateClearingReq{
			Channel:         enums.Channel_Slash,
			Amount:          amount,
			Notificator:     req.Notificator,
			AuthorizationID: origin.AuthorizationID,
		})
	case enums.CardTransactionType_VOID:
		result, err = u.simulator.SimulateReversal(ctx, &sharedbiz.SimulateReversalReq{
			Channel:         enums.Channel_Slash,
			Amount:          amount,
			Notificator:     req.Notificator,
			AuthorizationID: origin.AuthorizationID,
			Status:          enums.TransactionStatus_VOID,
		})
	case enums.CardTransactionType_REFUND:
		result, err = u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
			Channel:         enums.Channel_Slash,
			Amount:          amount,
			Notificator:     req.Notificator,
			AuthorizationID: &origin.AuthorizationID,
		})
	default:
		return nil, slasherrors.ErrInvalidOperation
	}
	if err != nil {
		return nil, slasherrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}

func (u *SlashUIUsecase) requireCardTransaction(ctx context.Context, id model.ID) error {
	exists, err := u.cardTransactionRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card transaction", "error", err)
		return slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return slasherrors.ErrResourceNotFound
	}
	return nil
}
