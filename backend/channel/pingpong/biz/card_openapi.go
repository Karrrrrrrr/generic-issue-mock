package biz

import (
	"context"
	"encoding/json"
	"reflect"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"

	"go.uber.org/zap"
)

type FindCardByRequestIDRequest struct {
	AccountID  model.ID
	RequestID  string
	RawRequest []byte
}

type GetCardRequest struct {
	AccountID model.ID
	ID        model.ID
}

type ChangeCardRequest struct {
	AccountID model.ID
	ID        model.ID
	Status    common.CardStatus
}

func (uc *PingPongOpenAPIUsecase) FindCardByRequestID(
	ctx context.Context,
	req *FindCardByRequestIDRequest,
) (*model.Card, bool, error) {
	if req.AccountID <= 0 || req.RequestID == "" {
		return nil, false, pingerrors.ErrInvalid
	}
	exists, err := uc.cardRepo.ExistsRequest(ctx, &CardRequestExistsRequest{
		AccountID: req.AccountID,
		RequestID: req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong card request", "error", err)
		return nil, false, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, false, nil
	}
	card, err := uc.cardRepo.FindRequest(ctx, &CardRequestFindRequest{
		AccountID: req.AccountID,
		RequestID: req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong card request", "error", err)
		return nil, false, pingerrors.ErrDatabase
	}
	var savedRequest, currentRequest any
	if json.Unmarshal(card.RawRequest, &savedRequest) != nil ||
		json.Unmarshal(req.RawRequest, &currentRequest) != nil ||
		!reflect.DeepEqual(savedRequest, currentRequest) {
		return nil, false, pingerrors.ErrConflict
	}
	return card, true, nil
}

func (uc *PingPongOpenAPIUsecase) GetCard(ctx context.Context, req *GetCardRequest) (*model.Card, error) {
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

func (uc *PingPongOpenAPIUsecase) ChangeCard(ctx context.Context, req *ChangeCardRequest) error {
	if req.Status != common.CardStatus_Active && req.Status != common.CardStatus_Frozen && req.Status != common.CardStatus_Deleted {
		return pingerrors.ErrInvalid
	}
	return uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockAccount(ctx, req.AccountID); err != nil {
			return err
		}
		card, err := uc.GetCard(ctx, &GetCardRequest{
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
