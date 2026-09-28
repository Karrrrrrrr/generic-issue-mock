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

type ListAuthorizationBalancesRequest struct {
	Statuses     []enums.CardTransactionStatus
	AccountID    *model.ID
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	ID           *model.ID
	CardID       *model.ID
	MerchantName *string
}

type UISimulateAuthorizationRequest struct {
	AccountID       model.ID
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
}

type UISimulateAuthorizationResult struct {
	Authorization   *model.Authorization
	CardTransaction *model.CardTransaction
}

type ClearAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
}

type ReverseAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
}

type RefundAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
}

type applyAuthorizationStepRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
	Type      enums.CardTransactionType
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

func (u *PhotonPayUIUsecase) ListAuthorizations(ctx context.Context, req *ListRequest) ([]*model.Authorization, error) {
	items, err := u.authorizationRepo.List(ctx, &AuthorizationListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay UI authorizations", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	return items, nil
}

func (u *PhotonPayUIUsecase) SimulateAuthorization(ctx context.Context, req *UISimulateAuthorizationRequest) (*UISimulateAuthorizationResult, error) {
	if req == nil {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	result, err := u.simulator.SimulateAuthorization(ctx, &sharedbiz.SimulateAuthorizationReq{
		AccountID:       req.AccountID,
		Channel:         enums.Channel_PhotonPay,
		CardID:          req.CardID,
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
	return &UISimulateAuthorizationResult{
		Authorization:   result.Authorization,
		CardTransaction: result.CardTransaction,
	}, nil
}

func (u *PhotonPayUIUsecase) ListAuthorizationBalances(ctx context.Context, req *ListAuthorizationBalancesRequest) ([]*AuthorizationBalance, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	items, err := u.authorizationRepo.ListAuthorizations(ctx, &AuthorizationListBalancesRequest{
		AccountIDs:   types.PointerSlice(req.AccountID),
		Statuses:     req.Statuses,
		IDs:          types.PointerSlice(req.ID),
		CardIDs:      types.PointerSlice(req.CardID),
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
	})
	if err != nil {
		zap.S().Errorw("list photonpay authorization balances", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	results := make([]*AuthorizationBalance, 0, len(items))
	for _, item := range items {
		stages, err := u.cardTransactionRepo.ListStages(ctx, &ListAuthorizationStagesRequest{
			AccountID: item.AccountID,
			ID:        item.ID,
		})
		if err != nil {
			zap.S().Errorw("list photonpay authorization stages", "error", err)
			return nil, photonpayerrors.ErrDatabaseOperation
		}
		results = append(results, authorizationBalance(&authorizationBalanceRequest{
			Authorization: item,
			Stages:        stages,
		}))
	}
	return results, nil
}

func (u *PhotonPayUIUsecase) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
		Amount:    req.Amount,
		Type:      enums.CardTransactionType_CLEAR,
	})
}

func (u *PhotonPayUIUsecase) ReverseAuthorization(ctx context.Context, req *ReverseAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
		Amount:    req.Amount,
		Type:      enums.CardTransactionType_VOID,
	})
}

func (u *PhotonPayUIUsecase) RefundAuthorization(ctx context.Context, req *RefundAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
		Amount:    req.Amount,
		Type:      enums.CardTransactionType_REFUND,
	})
}

func (u *PhotonPayUIUsecase) applyAuthorizationStep(ctx context.Context, req *applyAuthorizationStepRequest) (*model.CardTransaction, error) {
	if req.AccountID <= 0 || req.ID <= 0 || !req.Amount.IsPositive() {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	exists, err := u.authorizationRepo.AuthorizationExists(ctx, &ExistAuthorizationRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check photonpay simulation authorization", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}
	auth, err := u.authorizationRepo.FindAuthorizationDetail(ctx, &FindAuthorizationDetailRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find photonpay simulation authorization", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	var result *sharedbiz.CardTransactionSimulationResult
	switch req.Type {
	case enums.CardTransactionType_CLEAR:
		result, err = u.simulator.SimulateClearing(ctx, &sharedbiz.SimulateClearingReq{
			AccountID:       req.AccountID,
			Channel:         enums.Channel_PhotonPay,
			CardID:          auth.CardID,
			Amount:          req.Amount,
			Notificator:     u,
			AuthorizationID: auth.ID,
		})
	case enums.CardTransactionType_VOID:
		result, err = u.simulator.SimulateReversal(ctx, &sharedbiz.SimulateReversalReq{
			AccountID:       req.AccountID,
			Channel:         enums.Channel_PhotonPay,
			CardID:          auth.CardID,
			Amount:          req.Amount,
			Notificator:     u,
			AuthorizationID: auth.ID,
			Status:          enums.TransactionStatus_VOID,
		})
	case enums.CardTransactionType_REFUND:
		result, err = u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
			AccountID:       req.AccountID,
			Channel:         enums.Channel_PhotonPay,
			CardID:          auth.CardID,
			Amount:          req.Amount,
			Notificator:     u,
			AuthorizationID: &auth.ID,
			Currency:        auth.Currency,
		})
	default:
		return nil, photonpayerrors.ErrInvalidOperation
	}
	if err != nil {
		return nil, photonpayerrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}
