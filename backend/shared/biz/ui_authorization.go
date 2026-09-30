package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type ListUIAuthorizationsRequest struct {
	UIPageRequest
	UITimeRange
	ID           *model.ID
	AccountID    *model.ID
	CardID       *model.ID
	Status       *enums.CardTransactionStatus
	MerchantName *string
}

func (req *ListUIAuthorizationsRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID, req.AccountID, req.CardID}) ||
		(req.Status != nil && !validUITransactionStatus(*req.Status)) ||
		!validUIOptionalText(req.MerchantName) {
		return sharederrors.ErrInvalidUIRequest
	}
	if err := req.UITimeRange.Validate(); err != nil {
		return err
	}
	return req.UIPageRequest.Validate()
}

type UIAuthorizations interface {
	GetAuthorizationDetail(context.Context, *GetUIAuthorizationDetailRequest) (*UIAuthorizationDetail, error)
	GetAuthorization(context.Context, *GetUIAuthorizationRequest) (*model.Authorization, error)

	ListAuthorizations(context.Context, *ListUIAuthorizationsRequest) ([]*model.Authorization, int64, error)
}

func (uc *ui) ListAuthorizations(ctx context.Context, req *ListUIAuthorizationsRequest) ([]*model.Authorization, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := AuthorizationFilters{
		Channel:      uc.channel,
		IDs:          types.PointerSlice(req.ID),
		AccountIDs:   types.PointerSlice(req.AccountID),
		CardIDs:      types.PointerSlice(req.CardID),
		Statuses:     types.PointerSlice(req.Status),
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
	}
	items, err := uc.authorizationRepo.List(ctx, &AuthorizationListRequest{
		AuthorizationFilters: filters,
		Offset:               req.Offset,
		Limit:                req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI authorization", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.authorizationRepo.Count(ctx, &AuthorizationCountRequest{AuthorizationFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI authorization", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}

type GetUIAuthorizationRequest struct {
	ID        model.ID
	AccountID model.ID
}

func (req *GetUIAuthorizationRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.AccountID <= 0 {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func (uc *ui) GetAuthorization(ctx context.Context, req *GetUIAuthorizationRequest) (*model.Authorization, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	exists, err := uc.authorizationRepo.Exist(ctx, &AuthorizationExistRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("check shared UI authorization", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrAuthorizationNotFound
	}
	item, err := uc.authorizationRepo.Find(ctx, &AuthorizationFindRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("find shared UI authorization", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return item, nil
}

type UIAuthorizationAmounts struct {
	Settled   decimal.Decimal
	Reversed  decimal.Decimal
	Refunded  decimal.Decimal
	Remaining decimal.Decimal
}

func CalculateUIAuthorizationAmounts(item *model.Authorization) UIAuthorizationAmounts {
	result := UIAuthorizationAmounts{Remaining: item.Amount}
	if item.Status == enums.TransactionStatus_FAILED {
		result.Remaining = decimal.Zero
		return result
	}
	for _, stage := range item.CardTransactions {
		if stage.AccountID != item.AccountID || stage.Channel != item.Channel || stage.AuthorizationID != item.ID || stage.CardID != item.CardID {
			continue
		}
		switch stage.Type {
		case enums.CardTransactionType_CLEAR:
			if stage.Status == enums.TransactionStatus_SUCCEED {
				result.Settled = result.Settled.Add(stage.TxAmount)
			}
		case enums.CardTransactionType_VOID:
			if stage.Status == enums.TransactionStatus_SUCCEED || stage.Status == enums.TransactionStatus_VOID {
				result.Reversed = result.Reversed.Add(stage.TxAmount)
			}
		case enums.CardTransactionType_REFUND:
			if stage.Status == enums.TransactionStatus_SUCCEED {
				result.Refunded = result.Refunded.Add(stage.TxAmount)
			}
		}
	}
	result.Remaining = item.Amount.Sub(result.Settled).Sub(result.Reversed)
	return result
}

type GetUIAuthorizationDetailRequest struct {
	ID        model.ID
	AccountID model.ID
}

type UIAuthorizationDetail struct {
	Authorization *model.Authorization
	Card          *model.Card
}

func (req *GetUIAuthorizationDetailRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.AccountID <= 0 {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func (uc *ui) GetAuthorizationDetail(ctx context.Context, req *GetUIAuthorizationDetailRequest) (*UIAuthorizationDetail, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	authorization, err := uc.GetAuthorization(ctx, &GetUIAuthorizationRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	card, err := uc.GetCard(ctx, &GetUICardRequest{
		ID:        authorization.CardID,
		AccountID: authorization.AccountID,
	})
	if err != nil {
		return nil, err
	}
	return &UIAuthorizationDetail{
		Authorization: authorization,
		Card:          card,
	}, nil
}
