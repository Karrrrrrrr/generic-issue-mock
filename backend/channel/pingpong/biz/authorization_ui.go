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

type SimulateAuthorizationRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        common.Currency
	RequestID       string
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
}

type AuthorizationStageRequest struct {
	ID        model.ID
	Amount    decimal.Decimal
	RequestID string
	Kind      common.CardTransactionType
}

type ListAuthorizationsRequest struct {
	ID           *model.ID
	Status       *common.CardTransactionStatus
	MerchantName *string
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	AccountID    *model.ID
	CardID       *model.ID
	Offset       int
	Limit        int
}

func (uc *PingPongUIUsecase) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*model.Authorization, error) {
	if req == nil {
		return nil, pingerrors.ErrInvalid
	}
	result, err := uc.simulator.SimulateAuthorization(ctx, &sharedbiz.SimulateAuthorizationReq{
		Channel:         common.Channel_PingPong,
		CardID:          req.CardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		RequestID:       &req.RequestID,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
		Notificator:     sharedbiz.NoopNotificator{},
	})
	if err != nil {
		return nil, pingerrors.FromSimulation(err)
	}
	return result.Authorization, nil
}

func (uc *PingPongUIUsecase) Stage(ctx context.Context, req *AuthorizationStageRequest) error {
	if req == nil || req.ID <= 0 || !req.Amount.IsPositive() || req.RequestID == "" {
		return pingerrors.ErrInvalid
	}
	var err error
	switch req.Kind {
	case common.CardTransactionType_CLEAR:
		_, err = uc.simulator.SimulateClearing(ctx, &sharedbiz.SimulateClearingReq{
			Channel:         common.Channel_PingPong,
			Amount:          req.Amount,
			Notificator:     sharedbiz.NoopNotificator{},
			RequestID:       &req.RequestID,
			AuthorizationID: req.ID,
		})
	case common.CardTransactionType_VOID:
		_, err = uc.simulator.SimulateReversal(ctx, &sharedbiz.SimulateReversalReq{
			Channel:         common.Channel_PingPong,
			Amount:          req.Amount,
			Notificator:     sharedbiz.NoopNotificator{},
			RequestID:       &req.RequestID,
			AuthorizationID: req.ID,
			Status:          common.TransactionStatus_SUCCEED,
		})
	case common.CardTransactionType_REFUND:
		_, err = uc.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
			Channel:         common.Channel_PingPong,
			Amount:          req.Amount,
			Notificator:     sharedbiz.NoopNotificator{},
			RequestID:       &req.RequestID,
			AuthorizationID: &req.ID,
		})
	default:
		return pingerrors.ErrInvalid
	}
	if err != nil {
		return pingerrors.FromSimulation(err)
	}
	return nil
}

func (uc *PingPongUIUsecase) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) ([]*model.Authorization, int64, error) {
	items, err := uc.authorizationRepo.List(ctx, &AuthorizationListRequest{
		AccountIDs:   types.PointerSlice(req.AccountID),
		CardIDs:      types.PointerSlice(req.CardID),
		IDs:          types.PointerSlice(req.ID),
		Statuses:     types.PointerSlice(req.Status),
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
		Offset:       req.Offset,
		Limit:        &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong authorizations", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.authorizationRepo.Count(ctx, &AuthorizationCountRequest{
		AccountIDs:   types.PointerSlice(req.AccountID),
		CardIDs:      types.PointerSlice(req.CardID),
		IDs:          types.PointerSlice(req.ID),
		Statuses:     types.PointerSlice(req.Status),
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
	})
	if err != nil {
		zap.S().Errorw("count pingpong authorizations", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}

type UIAuthorizationDetailRequest struct {
	AccountID model.ID
	ID        model.ID
}

type UIAuthorizationDetail struct {
	Authorization *model.Authorization
	Card          *model.Card
}

func (uc *PingPongUIUsecase) GetAuthorization(ctx context.Context, req *UIAuthorizationDetailRequest) (*UIAuthorizationDetail, error) {
	if req == nil || req.AccountID <= 0 || req.ID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.authorizationRepo.Exists(ctx, &AuthorizationExistsRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong UI authorization", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	authorization, err := uc.authorizationRepo.Find(ctx, &AuthorizationFindRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong UI authorization", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	card, err := uc.getCard(ctx, &uiCardReference{
		AccountID: authorization.AccountID,
		ID:        authorization.CardID,
	})
	if err != nil {
		return nil, err
	}
	return &UIAuthorizationDetail{
		Authorization: authorization,
		Card:          card,
	}, nil
}

type AuthorizationAmounts struct {
	Settled   decimal.Decimal
	Reversed  decimal.Decimal
	Refunded  decimal.Decimal
	Remaining decimal.Decimal
}

func CalculateAuthorizationAmounts(auth *model.Authorization) AuthorizationAmounts {
	var amounts AuthorizationAmounts
	for _, transaction := range auth.CardTransactions {
		if transaction.Status != common.TransactionStatus_SUCCEED && transaction.Status != common.TransactionStatus_VOID {
			continue
		}
		switch transaction.Type {
		case common.CardTransactionType_CLEAR:
			amounts.Settled = amounts.Settled.Add(transaction.TxAmount)
		case common.CardTransactionType_VOID:
			amounts.Reversed = amounts.Reversed.Add(transaction.TxAmount)
		case common.CardTransactionType_REFUND:
			amounts.Refunded = amounts.Refunded.Add(transaction.TxAmount)
		}
	}
	amounts.Remaining = auth.Amount.Sub(amounts.Settled).Sub(amounts.Reversed)
	return amounts
}
