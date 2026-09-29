package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"go.uber.org/zap"
)

type ListUICardTransactionsRequest struct {
	UIPageRequest
	UITimeRange
	ID              *model.ID
	AccountID       *model.ID
	CardID          *model.ID
	AuthorizationID *model.ID
	Status          *enums.CardTransactionStatus
	Type            *enums.CardTransactionType
}

func (req *ListUICardTransactionsRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID, req.AccountID, req.CardID, req.AuthorizationID}) ||
		(req.Status != nil && !validUITransactionStatus(*req.Status)) ||
		(req.Type != nil && !validUITransactionType(*req.Type)) {
		return sharederrors.ErrInvalidUIRequest
	}
	if err := req.UITimeRange.Validate(); err != nil {
		return err
	}
	return req.UIPageRequest.Validate()
}

type UICardTransactions interface {
	GetCardTransaction(context.Context, *GetUICardTransactionRequest) (*model.CardTransaction, error)

	ListCardTransactions(context.Context, *ListUICardTransactionsRequest) ([]*model.CardTransaction, int64, error)
}

func (uc *ui) ListCardTransactions(ctx context.Context, req *ListUICardTransactionsRequest) ([]*model.CardTransaction, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := CardTransactionFilters{
		Channel:          uc.channel,
		IDs:              types.PointerSlice(req.ID),
		AccountIDs:       types.PointerSlice(req.AccountID),
		CardIDs:          types.PointerSlice(req.CardID),
		AuthorizationIDs: types.PointerSlice(req.AuthorizationID),
		Statuses:         types.PointerSlice(req.Status),
		Types:            types.PointerSlice(req.Type),
		CreatedFrom:      req.CreatedFrom,
		CreatedTo:        req.CreatedTo,
	}
	items, err := uc.cardTransactionRepo.List(ctx, &CardTransactionListRequest{
		CardTransactionFilters: filters,
		Offset:                 req.Offset,
		Limit:                  &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI card_transaction", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.cardTransactionRepo.Count(ctx, &CardTransactionCountRequest{CardTransactionFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI card_transaction", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}

type GetUICardTransactionRequest struct {
	ID        model.ID
	AccountID model.ID
}

func (req *GetUICardTransactionRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.AccountID <= 0 {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func (uc *ui) GetCardTransaction(ctx context.Context, req *GetUICardTransactionRequest) (*model.CardTransaction, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	exists, err := uc.cardTransactionRepo.Exist(ctx, &CardTransactionExistRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("check shared UI card_transaction", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrUITransactionNotFound
	}
	item, err := uc.cardTransactionRepo.Find(ctx, &CardTransactionFindRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("find shared UI card_transaction", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return item, nil
}
