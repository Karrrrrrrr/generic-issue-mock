package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"go.uber.org/zap"
)

type ListUICardsRequest struct {
	UIPageRequest
	UITimeRange
	ID         *model.ID
	AccountID  *model.ID
	Status     *enums.CardStatus
	CardNumber *string
}

func (req *ListUICardsRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID, req.AccountID}) ||
		(req.Status != nil && !validUICardStatus(*req.Status)) ||
		!validUIOptionalText(req.CardNumber) {
		return sharederrors.ErrInvalidUIRequest
	}
	if err := req.UITimeRange.Validate(); err != nil {
		return err
	}
	return req.UIPageRequest.Validate()
}

type UICards interface {
	UpdateCardStatus(context.Context, *UpdateUICardStatusRequest) (*model.Card, error)

	GetCard(context.Context, *GetUICardRequest) (*model.Card, error)

	ListCards(context.Context, *ListUICardsRequest) ([]*model.Card, int64, error)
}

func (uc *ui) ListCards(ctx context.Context, req *ListUICardsRequest) ([]*model.Card, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := CardFilters{
		Channel:     uc.channel,
		IDs:         types.PointerSlice(req.ID),
		AccountIDs:  types.PointerSlice(req.AccountID),
		Statuses:    types.PointerSlice(req.Status),
		CardNumber:  req.CardNumber,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
	}
	items, err := uc.cardRepo.List(ctx, &CardListRequest{
		CardFilters: filters,
		Offset:      req.Offset,
		Limit:       req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI card", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.cardRepo.Count(ctx, &CardCountRequest{CardFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI card", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}

type GetUICardRequest struct {
	ID        model.ID
	AccountID model.ID
}

func (req *GetUICardRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.AccountID <= 0 {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func (uc *ui) GetCard(ctx context.Context, req *GetUICardRequest) (*model.Card, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	exists, err := uc.cardRepo.Exist(ctx, &CardExistRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("check shared UI card", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrCardNotFound
	}
	item, err := uc.cardRepo.Find(ctx, &CardFindRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("find shared UI card", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return item, nil
}

type UpdateUICardStatusRequest struct {
	ID        model.ID
	AccountID model.ID
	Status    enums.CardStatus
}

func (req *UpdateUICardStatusRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.AccountID <= 0 ||
		(req.Status != enums.CardStatus_Active && req.Status != enums.CardStatus_Frozen && req.Status != enums.CardStatus_Deleted) {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func (uc *ui) UpdateCardStatus(ctx context.Context, req *UpdateUICardStatusRequest) (*model.Card, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result *model.Card
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockUIAccount(ctx, req.AccountID); err != nil {
			return err
		}
		exists, err := uc.cardRepo.Exist(ctx, &CardExistRequest{
			ID:        req.ID,
			AccountID: req.AccountID,
			Channel:   uc.channel,
		})
		if err != nil {
			zap.S().Errorw("check shared UI status card", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		if !exists {
			return sharederrors.ErrCardNotFound
		}
		result, err = uc.cardRepo.FindByIDWithLock(ctx, &CardFindByIDWithLockRequest{
			ID:        req.ID,
			AccountID: req.AccountID,
			Channel:   uc.channel,
		})
		if err != nil {
			zap.S().Errorw("lock shared UI status card", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		if (result.Status == enums.CardStatus_Deleted || result.Status == enums.CardStatus_Deleteing) && req.Status != enums.CardStatus_Deleted {
			return sharederrors.ErrUICardClosed
		}
		if err := uc.cardRepo.UpdateStatus(ctx, &CardUpdateStatusRequest{
			ID:        result.ID,
			AccountID: result.AccountID,
			Channel:   uc.channel,
			Status:    req.Status,
		}); err != nil {
			zap.S().Errorw("update shared UI card status", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		result.Status = req.Status
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
