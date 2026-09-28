package biz

import (
	"context"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharedbiz "generic-mock/shared/biz"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type SimulateAuthorizationRequest struct {
	AccountID    model.ID
	CardID       model.ID
	Amount       decimal.Decimal
	Currency     common.Currency
	RequestID    string
	MerchantName *string
}

type AuthorizationStageRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
	RequestID string
	Kind      common.CardTransactionType
}

type ListAuthorizationsRequest struct {
	AccountID *model.ID
	CardID    *model.ID
	Offset    int
	Limit     int
}

func (uc *PingPongUIUsecase) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*model.Authorization, error) {
	if req == nil {
		return nil, pingerrors.ErrInvalid
	}
	result, err := uc.simulator.SimulateAuthorization(ctx, &sharedbiz.SimulateAuthorizationReq{
		AccountID:    req.AccountID,
		Channel:      common.Channel_PingPong,
		CardID:       req.CardID,
		Amount:       req.Amount,
		Currency:     req.Currency,
		RequestID:    &req.RequestID,
		MerchantName: req.MerchantName,
		Notificator:  sharedbiz.NoopNotificator{},
	})
	if err != nil {
		return nil, pingerrors.FromSimulation(err)
	}
	return result.Authorization, nil
}

type authorizationReference struct {
	AccountID model.ID
	ID        model.ID
}

func (uc *PingPongUIUsecase) authorization(ctx context.Context, req *authorizationReference) (*model.Authorization, error) {
	if req.AccountID <= 0 || req.ID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.authorizationRepo.Exists(ctx, &AuthorizationExistsRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong authorization", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	item, err := uc.authorizationRepo.Find(ctx, &AuthorizationFindRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong authorization", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return item, nil
}

func (uc *PingPongUIUsecase) Stage(ctx context.Context, req *AuthorizationStageRequest) error {
	if req == nil || req.AccountID <= 0 || req.ID <= 0 || !req.Amount.IsPositive() || req.RequestID == "" {
		return pingerrors.ErrInvalid
	}
	auth, err := uc.authorization(ctx, &authorizationReference{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		return err
	}
	switch req.Kind {
	case common.CardTransactionType_CLEAR:
		_, err = uc.simulator.SimulateClearing(ctx, &sharedbiz.SimulateClearingReq{
			AccountID:       req.AccountID,
			Channel:         common.Channel_PingPong,
			CardID:          auth.CardID,
			Amount:          req.Amount,
			Notificator:     sharedbiz.NoopNotificator{},
			RequestID:       &req.RequestID,
			AuthorizationID: auth.ID,
		})
	case common.CardTransactionType_VOID:
		_, err = uc.simulator.SimulateReversal(ctx, &sharedbiz.SimulateReversalReq{
			AccountID:       req.AccountID,
			Channel:         common.Channel_PingPong,
			CardID:          auth.CardID,
			Amount:          req.Amount,
			Notificator:     sharedbiz.NoopNotificator{},
			RequestID:       &req.RequestID,
			AuthorizationID: auth.ID,
			Status:          common.TransactionStatus_SUCCEED,
		})
	case common.CardTransactionType_REFUND:
		_, err = uc.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
			AccountID:       req.AccountID,
			Channel:         common.Channel_PingPong,
			CardID:          auth.CardID,
			Amount:          req.Amount,
			Notificator:     sharedbiz.NoopNotificator{},
			RequestID:       &req.RequestID,
			AuthorizationID: &auth.ID,
			Currency:        auth.Currency,
		})
	default:
		return pingerrors.ErrInvalid
	}
	if err != nil {
		return pingerrors.FromSimulation(err)
	}
	return nil
}

func CalculateRemainingAuthorization(auth *model.Authorization) decimal.Decimal {
	remaining := auth.Amount
	for _, stage := range auth.CardTransactions {
		if stage.Status == common.TransactionStatus_SUCCEED && (stage.Type == common.CardTransactionType_CLEAR || stage.Type == common.CardTransactionType_VOID) {
			remaining = remaining.Sub(stage.TxAmount)
		}
	}
	return remaining
}

func (uc *PingPongUIUsecase) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) ([]*model.Authorization, int64, error) {
	items, err := uc.authorizationRepo.List(ctx, &AuthorizationListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		CardIDs:    types.PointerSlice(req.CardID),
		Offset:     req.Offset,
		Limit:      &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong authorizations", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.authorizationRepo.Count(ctx, &AuthorizationCountRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		CardIDs:    types.PointerSlice(req.CardID),
	})
	if err != nil {
		zap.S().Errorw("count pingpong authorizations", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}
