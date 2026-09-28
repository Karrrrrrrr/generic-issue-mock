package biz

import (
	"context"
	"time"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharedbiz "generic-mock/shared/biz"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type UIListCardTransactionsRequest struct {
	AccountID       *model.ID
	ID              *model.ID
	CardID          *model.ID
	AuthorizationID *model.ID
	Type            *common.CardTransactionType
	Status          *common.CardTransactionStatus
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
	Offset          int
	Limit           int
}

type UITransactionStageRequest struct {
	ID        model.ID
	Kind      common.CardTransactionType
	Amount    *decimal.Decimal
	RequestID string
}

type UISimulateRefundRequest struct {
	CardID          *model.ID
	AuthorizationID *model.ID
	Amount          decimal.Decimal
	Currency        *common.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
	RequestID       string
}

func (uc *PingPongUIUsecase) ListCardTransactions(ctx context.Context, req *UIListCardTransactionsRequest) ([]*model.CardTransaction, int64, error) {
	items, err := uc.cardTransactionRepo.List(ctx, &CardTransactionListRequest{
		AccountIDs:       types.PointerSlice(req.AccountID),
		IDs:              types.PointerSlice(req.ID),
		CardIDs:          types.PointerSlice(req.CardID),
		AuthorizationIDs: types.PointerSlice(req.AuthorizationID),
		Types:            types.PointerSlice(req.Type),
		Statuses:         types.PointerSlice(req.Status),
		CreatedFrom:      req.CreatedFrom,
		CreatedTo:        req.CreatedTo,
		Offset:           req.Offset,
		Limit:            &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong UI card transactions", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.cardTransactionRepo.Count(ctx, &CardTransactionCountRequest{
		AccountIDs:       types.PointerSlice(req.AccountID),
		IDs:              types.PointerSlice(req.ID),
		CardIDs:          types.PointerSlice(req.CardID),
		AuthorizationIDs: types.PointerSlice(req.AuthorizationID),
		Types:            types.PointerSlice(req.Type),
		Statuses:         types.PointerSlice(req.Status),
		CreatedFrom:      req.CreatedFrom,
		CreatedTo:        req.CreatedTo,
	})
	if err != nil {
		zap.S().Errorw("count pingpong UI card transactions", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}

func (req *UITransactionStageRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.RequestID == "" || req.Amount != nil && !req.Amount.IsPositive() {
		return pingerrors.ErrInvalid
	}
	switch req.Kind {
	case common.CardTransactionType_CLEAR, common.CardTransactionType_VOID, common.CardTransactionType_REFUND:
		return nil
	default:
		return pingerrors.ErrInvalid
	}
}

func (uc *PingPongUIUsecase) ApplyTransactionStage(ctx context.Context, req *UITransactionStageRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	exists, err := uc.cardTransactionRepo.ExistsForSimulation(ctx, &CardTransactionExistsForSimulationRequest{ID: req.ID})
	if err != nil {
		zap.S().Errorw("check pingpong UI simulation transaction", "error", err)
		return pingerrors.ErrDatabase
	}
	if !exists {
		return pingerrors.ErrNotFound
	}
	transaction, err := uc.cardTransactionRepo.FindForSimulation(ctx, &CardTransactionFindForSimulationRequest{ID: req.ID})
	if err != nil {
		zap.S().Errorw("find pingpong UI simulation transaction", "error", err)
		return pingerrors.ErrDatabase
	}
	if transaction.AuthorizationID <= 0 {
		return pingerrors.ErrInvalid
	}
	if req.Kind == common.CardTransactionType_REFUND {
		if transaction.Type != common.CardTransactionType_CLEAR || transaction.Status != common.TransactionStatus_SUCCEED {
			return pingerrors.ErrInvalid
		}
	} else if transaction.Type != common.CardTransactionType_AUTH || transaction.Status != common.TransactionStatus_AUTHORIZED {
		return pingerrors.ErrInvalid
	}
	amount := transaction.TxAmount
	if req.Amount != nil {
		amount = *req.Amount
	}
	return uc.Stage(ctx, &AuthorizationStageRequest{
		ID:        transaction.AuthorizationID,
		Kind:      req.Kind,
		Amount:    amount,
		RequestID: req.RequestID,
	})
}

func (uc *PingPongUIUsecase) SimulateRefund(ctx context.Context, req *UISimulateRefundRequest) (*model.CardTransaction, error) {
	if req == nil || req.RequestID == "" {
		return nil, pingerrors.ErrInvalid
	}
	result, err := uc.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
		Channel:         common.Channel_PingPong,
		CardID:          req.CardID,
		AuthorizationID: req.AuthorizationID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
		RequestID:       &req.RequestID,
		Notificator:     sharedbiz.NoopNotificator{},
	})
	if err != nil {
		return nil, pingerrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}
