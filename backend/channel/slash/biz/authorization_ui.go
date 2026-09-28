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

type ListAuthorizationBalancesRequest struct {
	Statuses     []enums.CardTransactionStatus
	AccountID    *model.ID
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	ID           *model.ID
	CardID       *model.ID
	MerchantName *string
}

type ListAuthorizationsRequest struct {
	AccountID *model.ID
	Offset    int
	Limit     int
	ID        *model.ID
	CardID    *model.ID
	Status    *enums.CardTransactionStatus
}

type SimulateAuthorizationRequest struct {
	AccountID       model.ID
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
	Notificator     sharedbiz.CardTransactionNotificator
}

type SimulateAuthorizationResult struct {
	Authorization   *model.Authorization
	CardTransaction *model.CardTransaction
}

type ClearAuthorizationRequest struct {
	AccountID   model.ID
	ID          model.ID
	Amount      decimal.Decimal
	Notificator sharedbiz.CardTransactionNotificator
}

type ReverseAuthorizationRequest struct {
	AccountID   model.ID
	ID          model.ID
	Amount      decimal.Decimal
	Notificator sharedbiz.CardTransactionNotificator
}

type RefundAuthorizationRequest struct {
	AccountID   model.ID
	ID          model.ID
	Amount      decimal.Decimal
	Notificator sharedbiz.CardTransactionNotificator
}

type applyAuthorizationStepRequest struct {
	AccountID   model.ID
	ID          model.ID
	Amount      decimal.Decimal
	Type        enums.CardTransactionType
	Notificator sharedbiz.CardTransactionNotificator
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

func (u *SlashUIUsecase) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*SimulateAuthorizationResult, error) {
	if req == nil {
		return nil, slasherrors.ErrInvalidOperation
	}
	result, err := u.simulator.SimulateAuthorization(ctx, &sharedbiz.SimulateAuthorizationReq{
		AccountID:       req.AccountID,
		Channel:         enums.Channel_Slash,
		CardID:          req.CardID,
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
	return &SimulateAuthorizationResult{
		Authorization:   result.Authorization,
		CardTransaction: result.CardTransaction,
	}, nil
}

func (u *SlashUIUsecase) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) ([]*model.Authorization, int64, error) {
	items, err := u.authorizationRepository.List(ctx, &AuthorizationListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		CardIDs:    types.PointerSlice(req.CardID),
		IDs:        types.PointerSlice(req.ID),
		Limit:      req.Limit,
		Offset:     req.Offset,
		Statuses:   types.PointerSlice(req.Status),
	})
	if err != nil {
		zap.S().Errorw("list slash authorizations", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
	}
	total, err := u.authorizationRepository.Count(ctx, &AuthorizationCountRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		CardIDs:    types.PointerSlice(req.CardID),
		IDs:        types.PointerSlice(req.ID),
		Limit:      req.Limit,
		Offset:     req.Offset,
		Statuses:   types.PointerSlice(req.Status),
	})
	if err != nil {
		zap.S().Errorw("count slash authorizations", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *SlashUIUsecase) GetAuthorization(ctx context.Context, id model.ID) (*model.Authorization, error) {
	if err := u.requireAuthorization(ctx, id); err != nil {
		return nil, err
	}
	item, err := u.authorizationRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash authorization", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}

	return item, nil
}

func (u *SlashUIUsecase) requireAuthorization(ctx context.Context, id model.ID) error {
	exists, err := u.authorizationRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash authorization", "error", err)
		return slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return slasherrors.ErrResourceNotFound
	}
	return nil
}

func (u *SlashUIUsecase) ListAuthorizationBalances(ctx context.Context, req *ListAuthorizationBalancesRequest) ([]*AuthorizationBalance, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, slasherrors.ErrInvalidOperation
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
		zap.S().Errorw("list slash authorization balances", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	results := make([]*AuthorizationBalance, 0, len(items))
	for _, item := range items {
		stages, err := u.cardTransactionRepository.ListStages(ctx, &ListAuthorizationStagesRequest{
			AccountID: item.AccountID,
			ID:        item.ID,
		})
		if err != nil {
			zap.S().Errorw("list slash authorization stages", "error", err)
			return nil, slasherrors.ErrDatabaseOperation
		}
		results = append(results, authorizationBalance(&authorizationBalanceRequest{
			Authorization: item,
			Stages:        stages,
		}))
	}
	return results, nil
}

func (u *SlashUIUsecase) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		AccountID:   req.AccountID,
		ID:          req.ID,
		Amount:      req.Amount,
		Type:        enums.CardTransactionType_CLEAR,
		Notificator: req.Notificator,
	})
}

func (u *SlashUIUsecase) ReverseAuthorization(ctx context.Context, req *ReverseAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		AccountID:   req.AccountID,
		ID:          req.ID,
		Amount:      req.Amount,
		Type:        enums.CardTransactionType_VOID,
		Notificator: req.Notificator,
	})
}

func (u *SlashUIUsecase) RefundAuthorization(ctx context.Context, req *RefundAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		AccountID:   req.AccountID,
		ID:          req.ID,
		Amount:      req.Amount,
		Type:        enums.CardTransactionType_REFUND,
		Notificator: req.Notificator,
	})
}

func (u *SlashUIUsecase) applyAuthorizationStep(ctx context.Context, req *applyAuthorizationStepRequest) (*model.CardTransaction, error) {
	if req.AccountID <= 0 || req.ID <= 0 || !req.Amount.IsPositive() {
		return nil, slasherrors.ErrInvalidOperation
	}
	exists, err := u.authorizationRepository.AuthorizationExists(ctx, &ExistAuthorizationRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check slash simulation authorization", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, slasherrors.ErrResourceNotFound
	}
	auth, err := u.authorizationRepository.FindAuthorizationDetail(ctx, &FindAuthorizationDetailRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find slash simulation authorization", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	var result *sharedbiz.CardTransactionSimulationResult
	switch req.Type {
	case enums.CardTransactionType_CLEAR:
		result, err = u.simulator.SimulateClearing(ctx, &sharedbiz.SimulateClearingReq{
			AccountID:       req.AccountID,
			Channel:         enums.Channel_Slash,
			CardID:          auth.CardID,
			Amount:          req.Amount,
			Notificator:     req.Notificator,
			AuthorizationID: auth.ID,
		})
	case enums.CardTransactionType_VOID:
		result, err = u.simulator.SimulateReversal(ctx, &sharedbiz.SimulateReversalReq{
			AccountID:       req.AccountID,
			Channel:         enums.Channel_Slash,
			CardID:          auth.CardID,
			Amount:          req.Amount,
			Notificator:     req.Notificator,
			AuthorizationID: auth.ID,
			Status:          enums.TransactionStatus_VOID,
		})
	case enums.CardTransactionType_REFUND:
		result, err = u.simulator.SimulateRefund(ctx, &sharedbiz.SimulateRefundReq{
			AccountID:       req.AccountID,
			Channel:         enums.Channel_Slash,
			CardID:          auth.CardID,
			Amount:          req.Amount,
			Notificator:     req.Notificator,
			AuthorizationID: &auth.ID,
			Currency:        auth.Currency,
		})
	default:
		return nil, slasherrors.ErrInvalidOperation
	}
	if err != nil {
		return nil, slasherrors.FromSimulation(err)
	}
	return result.CardTransaction, nil
}
