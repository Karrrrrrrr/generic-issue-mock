package biz

import (
	"context"
	"time"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type UIListCardsRequest struct {
	AccountID   *model.ID
	ID          *model.ID
	Status      *common.CardStatus
	CardNumber  *string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Offset      int
	Limit       int
}

type UIChangeCardRequest struct {
	AccountID model.ID
	ID        model.ID
	Status    common.CardStatus
}

type uiCardReference struct {
	AccountID model.ID
	ID        model.ID
}

func (uc *PingPongUIUsecase) ListCards(ctx context.Context, req *UIListCardsRequest) ([]*model.Card, int64, error) {
	items, err := uc.cardRepo.List(ctx, &CardListRequest{
		AccountIDs:  types.PointerSlice(req.AccountID),
		IDs:         types.PointerSlice(req.ID),
		Statuses:    types.PointerSlice(req.Status),
		CardNumber:  req.CardNumber,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		Offset:      req.Offset,
		Limit:       &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong UI cards", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.cardRepo.Count(ctx, &CardCountRequest{
		AccountIDs:  types.PointerSlice(req.AccountID),
		IDs:         types.PointerSlice(req.ID),
		Statuses:    types.PointerSlice(req.Status),
		CardNumber:  req.CardNumber,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
	})
	if err != nil {
		zap.S().Errorw("count pingpong UI cards", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}

func (uc *PingPongUIUsecase) ChangeCard(ctx context.Context, req *UIChangeCardRequest) error {
	if req.Status != common.CardStatus_Active && req.Status != common.CardStatus_Frozen && req.Status != common.CardStatus_Deleted {
		return pingerrors.ErrInvalid
	}
	return uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockAccount(ctx, req.AccountID); err != nil {
			return err
		}
		card, err := uc.getCard(ctx, &uiCardReference{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			return err
		}
		if card.Status == common.CardStatus_Deleted && req.Status != card.Status {
			return pingerrors.ErrClosed
		}
		card.Status = req.Status
		if err := uc.cardRepo.Save(ctx, card); err != nil {
			zap.S().Errorw("change pingpong card status", "error", err)
			return pingerrors.ErrDatabase
		}
		return nil
	})
}

func (uc *PingPongUIUsecase) getCard(ctx context.Context, req *uiCardReference) (*model.Card, error) {
	if req.AccountID <= 0 || req.ID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.cardRepo.Exists(ctx, &CardExistsRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong card", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	card, err := uc.cardRepo.Find(ctx, &CardFindRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong card", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return card, nil
}
