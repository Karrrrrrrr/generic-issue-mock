package biz

import (
	"context"
	"time"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/cardwallet"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type ListCardsRequest struct {
	AccountID   *model.ID
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Offset      int
	Limit       int
	ID          *model.ID
	CardNumber  *string
	Statuses    []enums.CardStatus
}

type CreateCardRequest struct {
	AccountID     model.ID
	CardHolderID  model.ID
	CardProductID *model.ID
	Currency      enums.Currency
}

type UpdateCardStatusRequest struct {
	AccountID model.ID
	ID        model.ID
	Status    enums.CardStatus
}

func (u *SlashUIUsecase) CreateCard(ctx context.Context, req *CreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCardHolder(txCtx, req.CardHolderID); err != nil {
			return err
		}
		holder, err := u.cardHolderRepository.FindByID(txCtx, req.CardHolderID)
		if err != nil {
			zap.S().Errorw("find slash card holder", "error", err)
			return slasherrors.ErrDatabaseOperation
		}

		if holder.AccountID != req.AccountID {
			return slasherrors.ErrResourceNotFound
		}
		products, err := u.cardProductRepository.List(txCtx)
		if err != nil {
			zap.S().Errorw("list slash channel products", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		productID := types.Value(req.CardProductID)
		if req.CardProductID == nil {
			for _, candidate := range products {
				if candidate.IsDefault {
					productID = candidate.ID
					break
				}
			}
		}
		exists, err := u.cardProductRepository.ExistByID(txCtx, productID)
		if err != nil {
			zap.S().Errorw("check slash channel product", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		if !exists {
			return slasherrors.ErrResourceNotFound
		}
		product, err := u.cardProductRepository.FindByIDForUpdate(txCtx, productID)
		if err != nil {
			zap.S().Errorw("lock slash channel product", "error", err)
			return slasherrors.ErrDatabaseOperation
		}

		product.NextCardNumber++
		generatedCard, ok := cardnumber.Generate(cardnumber.GenerateRequest{
			Channel:  enums.Channel_Slash,
			Prefix:   product.Prefix,
			Sequence: product.NextCardNumber,
		})
		if !ok {
			return slasherrors.ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance slash card product sequence", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		assignment, ok := cardwallet.Prepare(cardwallet.PrepareRequest{
			AccountID: req.AccountID,
			Channel:   enums.Channel_Slash,
			CardType:  enums.CardType_Single,
			Currency:  req.Currency,
		})
		if !ok {
			return slasherrors.ErrInvalidOperation
		}
		wallet := assignment.Wallet
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create slash card wallet", "error", err)
			return slasherrors.ErrDatabaseOperation
		}

		card = &model.Card{
			AccountID:              holder.AccountID,
			Channel:                enums.Channel_Slash,
			CardProductID:          product.ID,
			CardBin:                generatedCard.Bin,
			CardNumber:             generatedCard.Number,
			Cvv:                    randomx.Digits(3),
			ExpireAt:               time.Now().UTC().AddDate(2, 0, 0),
			Status:                 enums.CardStatus_Active,
			WalletID:               wallet.ID,
			CardHolderID:           req.CardHolderID,
			FormType:               enums.CardFormType_Virtual,
			CardCurrency:           req.Currency,
			CardScheme:             enums.CardScheme_Visa,
			CardType:               enums.CardType_Single,
			RequestID:              randomx.Digits(20),
			LastOperationRequestID: randomx.Digits(20),
		}
		if err := u.cardRepository.Create(txCtx, card); err != nil {
			zap.S().Errorw("create slash card", "error", err)
			return slasherrors.ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *SlashUIUsecase) ListCards(ctx context.Context, req *ListCardsRequest) ([]*model.Card, int64, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, 0, slasherrors.ErrInvalidOperation
	}
	items, err := u.cardRepository.List(ctx, &CardListRequest{
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
		zap.S().Errorw("list slash cards", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
	}
	total, err := u.cardRepository.Count(ctx, &CardCountRequest{
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
		zap.S().Errorw("count slash cards", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *SlashUIUsecase) GetCard(ctx context.Context, id model.ID) (*model.Card, error) {
	if err := u.requireCard(ctx, id); err != nil {
		return nil, err
	}
	card, err := u.cardRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash card", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}

	return card, nil
}

func (u *SlashUIUsecase) UpdateCardStatus(ctx context.Context, req *UpdateCardStatusRequest) (*model.Card, error) {
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
			zap.S().Errorw("update slash card status", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *SlashUIUsecase) requireCard(ctx context.Context, id model.ID) error {
	exists, err := u.cardRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card", "error", err)
		return slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return slasherrors.ErrResourceNotFound
	}
	return nil
}
