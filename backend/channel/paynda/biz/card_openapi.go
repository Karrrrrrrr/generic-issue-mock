package biz

import (
	"context"

	paynda "generic-mock/channel/paynda/enums"
	payndaerrors "generic-mock/channel/paynda/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type PayndaRequestLookup struct {
	AccountID model.ID
	RequestID string
}

type PayndaUpdateCardStatusRequest struct {
	AccountID model.ID
	CardID    model.ID
	RequestID string
	Status    paynda.CardStatus
}

type PayndaRequestResult struct {
	Card         *model.Card
	Transaction  *model.CardTransaction
	IsCardCreate bool
}

func (u *PayndaOpenAPIUsecase) GetCard(ctx context.Context, req *PayndaResourceRequest) (*model.Card, error) {
	if err := u.requireCard(ctx, req); err != nil {
		return nil, err
	}

	card, err := u.cardRepository.FindByID(ctx, &CardFindByIDRequest{
		AccountID: &req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find paynda card", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return card, nil
}

func (u *PayndaOpenAPIUsecase) ListCards(ctx context.Context, req *PayndaListRequest) ([]*model.Card, error) {
	items, err := u.cardRepository.List(ctx, &CardListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list paynda cards", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaOpenAPIUsecase) GetCardBalance(ctx context.Context, req *PayndaResourceRequest) (*model.Wallet, error) {
	card, err := u.GetCard(ctx, req)
	if err != nil {
		return nil, err
	}
	if card.WalletID == 0 {
		return nil, payndaerrors.ErrResourceNotFound
	}

	walletRequest := &PayndaResourceRequest{
		AccountID: req.AccountID,
		ID:        card.WalletID,
	}
	exists, err := u.walletRepository.ExistByID(ctx, &WalletExistByIDRequest{
		AccountID: &walletRequest.AccountID,
		ID:        walletRequest.ID,
	})
	if err != nil {
		zap.S().Errorw("check paynda card wallet", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, payndaerrors.ErrResourceNotFound
	}

	wallet, err := u.walletRepository.FindByID(ctx, &WalletFindByIDRequest{
		AccountID: &walletRequest.AccountID,
		ID:        walletRequest.ID,
	})
	if err != nil {
		zap.S().Errorw("find paynda card wallet", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return wallet, nil
}

func (u *PayndaOpenAPIUsecase) UpdateCardStatus(
	ctx context.Context,
	req *PayndaUpdateCardStatusRequest,
) (*model.Card, error) {
	if req.AccountID <= 0 || req.CardID <= 0 {
		return nil, payndaerrors.ErrInvalidOperation
	}
	if req.RequestID == "" {
		req.RequestID = randomx.Digits(20)
	}

	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistForStatusChange(txCtx, &CardStatusExistsRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("check paynda card status change", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		if !exists {
			return payndaerrors.ErrResourceNotFound
		}
		card, err = u.cardRepository.LockForStatusChange(txCtx, &CardStatusLockRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("lock paynda card status change", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		nextStatus := paynda.ConvertCardStatusToGenericCardStatus(req.Status)
		if (card.Status == enums.CardStatus_Deleted && nextStatus != enums.CardStatus_Deleted) ||
			(card.Status == enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleted) {
			return payndaerrors.ErrCardClosed
		}
		card.Status = nextStatus
		card.LastOperationRequestID = req.RequestID
		card.LastOperationType = enums.OperationType_UpdateCard
		card.LastOperationStatus = enums.OperationStatus_Succeed
		if err := u.cardRepository.SaveStatus(txCtx, &CardStatusSaveRequest{
			AccountID:              card.AccountID,
			ID:                     card.ID,
			Status:                 card.Status,
			LastOperationRequestID: card.LastOperationRequestID,
			LastOperationType:      card.LastOperationType,
			LastOperationStatus:    card.LastOperationStatus,
		}); err != nil {
			zap.S().Errorw("update paynda card status", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PayndaOpenAPIUsecase) ReleaseCard(ctx context.Context, req *PayndaUpdateCardStatusRequest) (*model.Card, error) {
	req.Status = paynda.CardStatus_Deleted
	return u.UpdateCardStatus(ctx, req)
}

func (u *PayndaOpenAPIUsecase) ListCardBalanceUpdates(
	ctx context.Context,
	req *PayndaListRequest,
) ([]*model.CardTransaction, error) {
	items, err := u.cardTransactionRepository.List(ctx, &CardTransactionListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Offset:     req.Offset,
		Limit:      req.Limit,
		Types: []enums.CardTransactionType{
			enums.CardTransactionType_FundIn,
			enums.CardTransactionType_FundOut,
		},
	})
	if err != nil {
		zap.S().Errorw("list paynda card balance updates", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaOpenAPIUsecase) FindRequestResult(
	ctx context.Context,
	req *PayndaRequestLookup,
) (*PayndaRequestResult, error) {
	exists, err := u.cardRepository.ExistByRequestID(ctx, (*CardExistByRequestIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check paynda create card request", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if exists {
		card, err := u.cardRepository.FindByRequestID(ctx, (*CardFindByRequestIDRequest)(req))
		if err != nil {
			zap.S().Errorw("find paynda create card request", "error", err)
			return nil, payndaerrors.ErrDatabaseOperation
		}
		return &PayndaRequestResult{
			Card:         card,
			IsCardCreate: true,
		}, nil
	}

	exists, err = u.cardRepository.ExistByLastOperationRequestID(ctx, (*CardExistByLastOperationRequestIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check paynda card operation request", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if exists {
		card, err := u.cardRepository.FindByLastOperationRequestID(ctx, (*CardFindByLastOperationRequestIDRequest)(req))
		if err != nil {
			zap.S().Errorw("find paynda card operation request", "error", err)
			return nil, payndaerrors.ErrDatabaseOperation
		}
		return &PayndaRequestResult{Card: card}, nil
	}

	exists, err = u.cardTransactionRepository.ExistByRequestID(ctx, (*CardTransactionExistByRequestIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check paynda card balance transfer request", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, payndaerrors.ErrResourceNotFound
	}
	transaction, err := u.cardTransactionRepository.FindByRequestID(ctx, (*CardTransactionFindByRequestIDRequest)(req))
	if err != nil {
		zap.S().Errorw("find paynda card balance transfer request", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return &PayndaRequestResult{Transaction: transaction}, nil
}

func (u *PayndaOpenAPIUsecase) requireCard(ctx context.Context, req *PayndaResourceRequest) error {
	exists, err := u.cardRepository.ExistByID(ctx, &CardExistByIDRequest{
		AccountID: &req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check paynda card", "error", err)
		return payndaerrors.ErrDatabaseOperation
	}
	if !exists {
		return payndaerrors.ErrResourceNotFound
	}

	return nil
}
