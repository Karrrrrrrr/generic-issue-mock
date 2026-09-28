package biz

import (
	"context"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type ListUICardsRequest struct {
	AccountID   *model.ID
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Offset      int
	Limit       int
	ID          *model.ID
	CardNumber  *string
	Statuses    []enums.CardStatus
}

type UIOpenCardRequest struct {
	CardProductID model.ID
	AccountID     model.ID
	CardHolderID  model.ID
	Currency      enums.Currency
	RequestID     string
}

type UIChangeCardStatusRequest struct {
	AccountID model.ID
	CardID    model.ID
	Status    enums.CardStatus
}

func (req *UIOpenCardRequest) Validate() error {
	if req == nil || req.CardProductID <= 0 {
		return photonpayerrors.ErrInvalidOperation
	}
	return nil
}

func (u *PhotonPayUIUsecase) OpenCard(ctx context.Context, req *UIOpenCardRequest) (*model.Card, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardHolderRepo.ExistCardHolderByAccountID(txCtx, &CardHolderExistCardHolderByAccountIDRequest{
			AccountID: &req.AccountID,
			ID:        req.CardHolderID,
		})
		if err != nil {
			zap.S().Errorw("check photonpay UI card holder", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}
		if !exists {
			return photonpayerrors.ErrResourceNotFound
		}

		productExists, err := u.cardProductRepo.ExistByID(txCtx, &CardProductExistByIDRequest{ID: req.CardProductID})
		if err != nil {
			zap.S().Errorw("check photonpay UI card product", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}
		if !productExists {
			return photonpayerrors.ErrResourceNotFound
		}

		product, err := u.cardProductRepo.FindByIDForUpdate(txCtx, &CardProductFindByIDForUpdateRequest{ID: req.CardProductID})
		if err != nil {
			zap.S().Errorw("lock photonpay UI card product", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}

		product.NextCardNumber++
		generatedCard, ok := cardnumber.Generate(cardnumber.GenerateRequest{
			Channel:  enums.Channel_PhotonPay,
			Prefix:   product.Prefix,
			Sequence: product.NextCardNumber,
		})
		if !ok {
			return photonpayerrors.ErrInvalidOperation
		}
		if err := u.cardProductRepo.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance photonpay UI card product sequence", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}
		virtualAccount, err := u.virtualAccountRepo.FindByAccountID(txCtx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay UI account virtual account", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}

		card = &model.Card{
			AccountID:              req.AccountID,
			Channel:                enums.Channel_PhotonPay,
			CardProductID:          product.ID,
			CardBin:                generatedCard.Bin,
			CardNumber:             generatedCard.Number,
			Cvv:                    randomx.Digits(3),
			ExpireAt:               time.Now().UTC().AddDate(0, 24, 0),
			Status:                 enums.CardStatus_Active,
			VirtualAccountID:       &virtualAccount.ID,
			WalletID:               virtualAccount.WalletID,
			CardHolderID:           req.CardHolderID,
			FormType:               enums.CardFormType_Virtual,
			RequestID:              req.RequestID,
			LastOperationRequestID: req.RequestID,
			LastOperationType:      enums.OperationType_OpenCard,
			LastOperationStatus:    enums.OperationStatus_Succeed,
			CardCurrency:           req.Currency,
			CardScheme:             photon.CardScheme,
			CardType:               enums.CardType_Share,
		}
		if err := u.cardRepo.CreateCard(txCtx, card); err != nil {
			zap.S().Errorw("create photonpay UI card", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	u.dispatchCardStatus(ctx, card)

	return card, nil
}

func (u *PhotonPayUIUsecase) ListCards(ctx context.Context, req *ListUICardsRequest) ([]*model.Card, int64, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, 0, photonpayerrors.ErrInvalidOperation
	}
	cards, err := u.cardRepo.ListCards(ctx, &CardListCardsRequest{
		AccountIDs:  types.PointerSlice(req.AccountID),
		IDs:         types.PointerSlice(req.ID),
		Statuses:    req.Statuses,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		CardNumber:  req.CardNumber,
		Limit:       req.Limit,
		Offset:      req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay UI cards", "error", err)

		return nil, 0, photonpayerrors.ErrDatabaseOperation
	}

	total, err := u.cardRepo.Count(ctx, &CardCountRequest{
		AccountIDs:  types.PointerSlice(req.AccountID),
		IDs:         types.PointerSlice(req.ID),
		Statuses:    req.Statuses,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		CardNumber:  req.CardNumber,
	})
	if err != nil {
		zap.S().Errorw("count photonpay UI cards", "error", err)
		return nil, 0, photonpayerrors.ErrDatabaseOperation
	}

	return cards, total, nil
}

func (u *PhotonPayUIUsecase) ChangeCardStatus(ctx context.Context, req *UIChangeCardStatusRequest) (*model.Card, error) {
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
		if (card.Status == enums.CardStatus_Deleted && nextStatus != enums.CardStatus_Deleted) ||
			(card.Status == enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleted) {
			return photonpayerrors.ErrCardClosed
		}
		card.Status = nextStatus
		if err := u.cardRepo.SaveStatus(txCtx, &CardStatusSaveRequest{
			AccountID:              card.AccountID,
			ID:                     card.ID,
			Status:                 card.Status,
			LastOperationRequestID: card.LastOperationRequestID,
			LastOperationType:      card.LastOperationType,
			LastOperationStatus:    card.LastOperationStatus,
		}); err != nil {
			zap.S().Errorw("update photonpay UI card status", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PhotonPayUIUsecase) getCard(ctx context.Context, id model.ID) (*model.Card, error) {
	exists, err := u.cardRepo.ExistCardByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay UI card", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}

	card, err := u.cardRepo.FindCardByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find photonpay UI card", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return card, nil
}
