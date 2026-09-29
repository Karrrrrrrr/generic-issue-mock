package biz

import (
	"context"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type RequestResultResourceRequest struct {
	AccountID model.ID
	RequestID string
}

type ChangeCardStatusRequest struct {
	AccountID model.ID
	CardID    model.ID
	RequestID string
	Status    common.CardStatus
	Operation common.OperationType
}

type UpdateCardRequest struct {
	AccountID model.ID
	CardID    model.ID
	RequestID string
}

func (u *PhotonPayOpenAPIUsecase) GetCard(ctx context.Context, req *ResourceRequest) (*model.Card, error) {
	if err := u.requireCard(ctx, req); err != nil {
		return nil, err
	}

	card, err := u.cardRepo.FindCardByAccountID(ctx, (*CardFindCardByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("find photonpay card", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return card, nil
}

func (u *PhotonPayOpenAPIUsecase) ListCards(ctx context.Context, req *ListRequest) ([]*model.Card, error) {
	cards, err := u.cardRepo.ListCards(ctx, &CardListCardsRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay cards", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return cards, nil
}

func (u *PhotonPayOpenAPIUsecase) FindCardProductByBinPrefix(
	ctx context.Context,
	prefix string,
) (*model.CardProduct, error) {
	exists, err := u.cardProductRepo.ExistByPrefix(ctx, prefix)
	if err != nil {
		zap.S().Errorw("check photonpay card product", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		zap.S().Warnw(
			"photonpay card product not found for card bin",
			"card_bin",
			prefix,
		)
		return nil, photonpayerrors.ErrCardBinNotFound
	}
	product, err := u.cardProductRepo.FindByPrefix(ctx, prefix)
	if err != nil {
		zap.S().Errorw("find photonpay card product", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	return product, nil
}

func (u *PhotonPayOpenAPIUsecase) FindDefaultVirtualAccount(
	ctx context.Context,
	accountID model.ID,
) (*model.VirtualAccount, error) {
	virtualAccounts, err := u.virtualAccountRepo.ListVirtualAccounts(ctx, &VirtualAccountListRequest{
		AccountIDs: []model.ID{accountID},
	})
	if err != nil {
		zap.S().Errorw(
			"list photonpay account virtual accounts",
			"account_id",
			accountID,
			"error",
			err,
		)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if len(virtualAccounts) == 0 {
		zap.S().Errorw(
			"photonpay default virtual account not found",
			"account_id",
			accountID,
		)
		return nil, photonpayerrors.ErrDefaultVirtualAccountNotFound
	}
	return virtualAccounts[len(virtualAccounts)-1], nil
}

func (u *PhotonPayOpenAPIUsecase) GetRequestResult(ctx context.Context, req *RequestResultResourceRequest) (*model.Card, error) {
	exists, err := u.cardRepo.ExistCardByRequestIDForAccount(ctx, (*CardExistCardByRequestIDForAccountRequest)(req))
	if err != nil {
		zap.S().Errorw("check photonpay card request", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if exists {
		card, err := u.cardRepo.FindCardByRequestIDForAccount(ctx, (*CardFindCardByRequestIDForAccountRequest)(req))
		if err != nil {
			zap.S().Errorw("find photonpay card request", "error", err)

			return nil, photonpayerrors.ErrDatabaseOperation
		}

		return card, nil
	}

	exists, err = u.cardRepo.ExistCardByLastOperationRequestIDForAccount(ctx, (*CardExistCardByLastOperationRequestIDForAccountRequest)(req))
	if err != nil {
		zap.S().Errorw("check photonpay card operation request", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}

	card, err := u.cardRepo.FindCardByLastOperationRequestIDForAccount(ctx, (*CardFindCardByLastOperationRequestIDForAccountRequest)(req))
	if err != nil {
		zap.S().Errorw("find photonpay card operation request", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return card, nil
}

func (u *PhotonPayOpenAPIUsecase) ChangeCardStatus(ctx context.Context, req *ChangeCardStatusRequest) (*model.Card, error) {
	if req.AccountID <= 0 || req.CardID <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepo.ExistForStatusChange(txCtx, &CardStatusExistsRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("check photonpay card status change", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		if !exists {
			return photonpayerrors.ErrResourceNotFound
		}
		card, err = u.cardRepo.LockForStatusChange(txCtx, &CardStatusLockRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay card status change", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		nextStatus := req.Status
		if (card.Status == common.CardStatus_Deleted && nextStatus != common.CardStatus_Deleted) ||
			(card.Status == common.CardStatus_Deleteing && nextStatus != common.CardStatus_Deleteing && nextStatus != common.CardStatus_Deleted) {
			return photonpayerrors.ErrCardClosed
		}
		card.Status = nextStatus
		if req.RequestID != "" {
			card.LastOperationRequestID = req.RequestID
		} else {
			card.LastOperationRequestID = randomx.Digits(20)
		}
		card.LastOperationType = req.Operation
		card.LastOperationStatus = common.OperationStatus_Succeed

		if err := u.cardRepo.SaveStatus(txCtx, &CardStatusSaveRequest{
			AccountID:              card.AccountID,
			ID:                     card.ID,
			Status:                 card.Status,
			LastOperationRequestID: card.LastOperationRequestID,
			LastOperationType:      card.LastOperationType,
			LastOperationStatus:    card.LastOperationStatus,
		}); err != nil {
			zap.S().Errorw("update photonpay card status", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PhotonPayOpenAPIUsecase) UpdateCard(ctx context.Context, req *UpdateCardRequest) (*model.Card, error) {
	if req.AccountID <= 0 || req.CardID <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCard(txCtx, &ResourceRequest{
			AccountID: &req.AccountID,
			ID:        req.CardID,
		}); err != nil {
			return err
		}

		var err error
		card, err = u.cardRepo.FindCard(txCtx, &FindCardRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("find photonpay card for update", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}

		card.LastOperationRequestID = req.RequestID
		card.LastOperationType = common.OperationType_UpdateCard
		card.LastOperationStatus = common.OperationStatus_Succeed
		if err := u.cardRepo.UpdateOperation(txCtx, &CardOperationUpdateRequest{
			AccountID: card.AccountID,
			ID:        card.ID,
			RequestID: card.LastOperationRequestID,
			Type:      card.LastOperationType,
			Status:    card.LastOperationStatus,
		}); err != nil {
			zap.S().Errorw("update photonpay card", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PhotonPayOpenAPIUsecase) requireCard(ctx context.Context, req *ResourceRequest) error {
	exists, err := u.cardRepo.ExistCardByAccountID(ctx, (*CardExistCardByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check photonpay card", "error", err)

		return photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return photonpayerrors.ErrResourceNotFound
	}

	return nil
}
