package biz

import (
	"context"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharedbiz "generic-mock/shared/biz"

	"go.uber.org/zap"
)

type OpenAPIListCardsRequest struct {
	AccountID model.ID
	Offset    int
	Limit     int
	Status    *enums.CardStatus
}

type OpenAPICardRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type OpenAPIUpdateCardRequest struct {
	AccountID model.ID
	ID        model.ID
	Status    enums.CardStatus
}

func (u *SlashOpenAPIUsecase) ListCards(ctx context.Context, req *OpenAPIListCardsRequest) ([]*model.Card, error) {
	items, err := u.cardRepository.List(ctx, &CardListRequest{
		AccountIDs: []model.ID{req.AccountID},
		Offset:     req.Offset,
		Limit:      req.Limit,
		Statuses:   types.PointerSlice(req.Status),
	})
	if err != nil {
		zap.S().Errorw("list slash openapi cards", "error", err)

		return nil, slasherrors.ErrDatabaseOperation
	}

	return items, nil
}

func (u *SlashOpenAPIUsecase) FindCardByRequestID(
	ctx context.Context,
	req *OpenAPICardRequestIDRequest,
) (*model.Card, bool, error) {
	if req.RequestID == "" {
		return nil, false, slasherrors.ErrInvalidOperation
	}
	exists, err := u.sharedCardRepository.ExistByRequestID(ctx, &sharedbiz.CardExistByRequestIDRequest{
		AccountID: req.AccountID,
		Channel:   enums.Channel_Slash,
		RequestID: req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("check slash openapi card request", "error", err)
		return nil, false, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, false, nil
	}
	card, err := u.sharedCardRepository.FindByRequestID(ctx, &sharedbiz.CardFindByRequestIDRequest{
		AccountID: req.AccountID,
		Channel:   enums.Channel_Slash,
		RequestID: req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("find slash openapi card request", "error", err)
		return nil, false, slasherrors.ErrDatabaseOperation
	}
	return card, true, nil
}

func (u *SlashOpenAPIUsecase) GetCard(ctx context.Context, req *ResourceRequest) (*model.Card, error) {
	exists, err := u.cardRepository.ExistByAccountID(ctx, (*CardExistByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check slash openapi card", "error", err)

		return nil, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, slasherrors.ErrResourceNotFound
	}

	item, err := u.cardRepository.FindByAccountID(ctx, (*CardFindByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("find slash openapi card", "error", err)

		return nil, slasherrors.ErrDatabaseOperation
	}

	return item, nil
}

func (u *SlashOpenAPIUsecase) UpdateCard(ctx context.Context, req *OpenAPIUpdateCardRequest) (*model.Card, error) {
	if req.AccountID <= 0 || req.ID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistForStatusChange(txCtx, &CardStatusExistsRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			zap.S().Errorw("check slash card status change", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		if !exists {
			return slasherrors.ErrResourceNotFound
		}
		card, err = u.cardRepository.LockForStatusChange(txCtx, &CardStatusLockRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			zap.S().Errorw("lock slash card status change", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		nextStatus := req.Status
		if (card.Status == enums.CardStatus_Deleted && nextStatus != enums.CardStatus_Deleted) ||
			(card.Status == enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleted) {
			return slasherrors.ErrCardClosed
		}
		card.Status = nextStatus
		if err := u.cardRepository.SaveStatus(txCtx, &CardStatusSaveRequest{
			AccountID: card.AccountID,
			ID:        card.ID,
			Status:    card.Status,
		}); err != nil {
			zap.S().Errorw("update slash openapi card", "error", err)

			return slasherrors.ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}
