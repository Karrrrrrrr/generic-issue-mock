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

type ListAuthorizationBalancesRequest struct {
	Statuses     []enums.CardTransactionStatus
	AccountID    *model.ID
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	ID           *model.ID
	CardID       *model.ID
	MerchantName *string
}

type PayndaSimulateAuthorizationRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
}

type PayndaSimulateAuthorizationResult struct {
	Authorization   *model.Authorization
	CardTransaction *model.CardTransaction
}

type ClearAuthorizationRequest struct {
	ID     model.ID
	Amount decimal.Decimal
}

type ReverseAuthorizationRequest struct {
	ID     model.ID
	Amount decimal.Decimal
}

type RefundAuthorizationRequest struct {
	ID     model.ID
	Amount decimal.Decimal
}

type applyAuthorizationStepRequest struct {
	ID     model.ID
	Amount decimal.Decimal
	Type   enums.CardTransactionType
}

type AuthorizationBalance struct {
	Authorization *model.Authorization
	Settled       decimal.Decimal
	Reversed      decimal.Decimal
	Refunded      decimal.Decimal
	Remaining     decimal.Decimal
}

type authorizationBalanceRequest struct {
	Authorization *model.Authorization
	Stages        []*model.CardTransaction
}

func authorizationBalance(req *authorizationBalanceRequest) *AuthorizationBalance {
	auth := req.Authorization
	stages := req.Stages
	result := &AuthorizationBalance{
		Authorization: auth,
		Remaining:     auth.Amount,
	}
	for _, stage := range stages {
		switch stage.Type {
		case enums.CardTransactionType_CLEAR:
			if stage.Status == enums.TransactionStatus_SUCCEED {
				result.Settled = result.Settled.Add(stage.TxAmount)
			}
		case enums.CardTransactionType_VOID:
			if stage.Status == enums.TransactionStatus_VOID || stage.Status == enums.TransactionStatus_SUCCEED {
				result.Reversed = result.Reversed.Add(stage.TxAmount)
			}
		case enums.CardTransactionType_REFUND:
			if stage.Status == enums.TransactionStatus_SUCCEED {
				result.Refunded = result.Refunded.Add(stage.TxAmount)
			}
		}
	}
	result.Remaining = auth.Amount.Sub(result.Settled).Sub(result.Reversed)
	return result
}

func (u *PayndaUIUsecase) ListAuthorizations(ctx context.Context, req *PayndaListRequest) ([]*model.Authorization, error) {
	items, err := u.authorizationRepository.List(ctx, &AuthorizationListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list paynda UI authorizations", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	return items, nil
}

func (u *PayndaUIUsecase) SimulateAuthorization(ctx context.Context, req *PayndaSimulateAuthorizationRequest) (*PayndaSimulateAuthorizationResult, error) {
	if req == nil {
		return nil, payndaerrors.ErrInvalidOperation
	}
	result, err := u.simulator.SimulateAuthorization(ctx, &sharedbiz.SimulateAuthorizationReq{
		Channel:         enums.Channel_Paynda,
		CardID:          req.CardID,
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
	return &PayndaSimulateAuthorizationResult{
		Authorization:   result.Authorization,
		CardTransaction: result.CardTransaction,
	}, nil
}

func (u *PayndaUIUsecase) ListAuthorizationBalances(ctx context.Context, req *ListAuthorizationBalancesRequest) ([]*AuthorizationBalance, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, payndaerrors.ErrInvalidOperation
	}
	items, err := u.authorizationRepository.ListAuthorizations(ctx, &AuthorizationListBalancesRequest{
		AccountIDs:   types.PointerSlice(req.AccountID),
		Statuses:     req.Statuses,
		IDs:          types.PointerSlice(req.ID),
		CardIDs:      types.PointerSlice(req.CardID),
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
	})
	if err != nil {
		zap.S().Errorw("list paynda authorization balances", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	results := make([]*AuthorizationBalance, 0, len(items))
	for _, item := range items {
		stages, err := u.cardTransactionRepository.ListStages(ctx, &ListAuthorizationStagesRequest{
			AccountID: item.AccountID,
			ID:        item.ID,
		})
		if err != nil {
			zap.S().Errorw("list paynda authorization stages", "error", err)
			return nil, payndaerrors.ErrDatabaseOperation
		}
		results = append(results, authorizationBalance(&authorizationBalanceRequest{
			Authorization: item,
			Stages:        stages,
		}))
	}
	return results, nil
}

func (u *PayndaUIUsecase) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		ID:     req.ID,
		Amount: req.Amount,
		Type:   enums.CardTransactionType_CLEAR,
	})
}

func (u *PayndaUIUsecase) ReverseAuthorization(ctx context.Context, req *ReverseAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		ID:     req.ID,
		Amount: req.Amount,
		Type:   enums.CardTransactionType_VOID,
	})
}

func (u *PayndaUIUsecase) RefundAuthorization(ctx context.Context, req *RefundAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		ID:     req.ID,
		Amount: req.Amount,
		Type:   enums.CardTransactionType_REFUND,
	})
}

func (u *PayndaUIUsecase) applyAuthorizationStep(ctx context.Context, req *applyAuthorizationStepRequest) (*model.CardTransaction, error) {
	if req == nil {
		return nil, payndaerrors.ErrInvalidOperation
	}
	var err error
	var result *sharedbiz.CardTransactionSimulationResult
	switch req.Type {
	case enums.CardTransactionType_CLEAR:
		result, err = u.simulator.SimulateClearing(ctx, &sharedbiz.SimulateClearingReq{
			Channel:         enums.Channel_Paynda,
			Amount:          req.Amount,
			Notificator:     u,
			AuthorizationID: req.ID,
		})
	case enums.CardTransactionType_VOID:
		result, err = u.simulator.SimulateReversal(ctx, &sharedbiz.SimulateReversalReq{
			Channel:         enums.Channel_Paynda,
			Amount:          req.Amount,
			Notificator:     u,
			AuthorizationID: req.ID,
			Status:          enums.TransactionStatus_VOID,
		})
	case enums.CardTransactionType_REFUND:
		result, err = u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
			Channel:         enums.Channel_Paynda,
			Amount:          req.Amount,
			Notificator:     u,
			AuthorizationID: &req.ID,
		})
	default:
		return nil, payndaerrors.ErrInvalidOperation
	}
	if err != nil {
		return nil, payndaerrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}
