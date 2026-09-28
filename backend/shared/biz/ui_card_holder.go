package biz

import (
	"context"

	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"go.uber.org/zap"
)

type ListUICardHoldersRequest struct {
	UIPageRequest
	ID        *model.ID
	AccountID *model.ID
}

func (req *ListUICardHoldersRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID, req.AccountID}) {
		return sharederrors.ErrInvalidUIRequest
	}
	return req.UIPageRequest.Validate()
}

type UICardHolders interface {
	ListCardHolders(context.Context, *ListUICardHoldersRequest) ([]*model.CardHolder, int64, error)
}

func (uc *ui) ListCardHolders(ctx context.Context, req *ListUICardHoldersRequest) ([]*model.CardHolder, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := CardHolderFilters{
		Channel:    uc.channel,
		IDs:        types.PointerSlice(req.ID),
		AccountIDs: types.PointerSlice(req.AccountID),
	}
	items, err := uc.cardHolderRepo.List(ctx, &CardHolderListRequest{
		CardHolderFilters: filters,
		Offset:            req.Offset,
		Limit:             req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI card_holder", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.cardHolderRepo.Count(ctx, &CardHolderCountRequest{CardHolderFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI card_holder", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}
