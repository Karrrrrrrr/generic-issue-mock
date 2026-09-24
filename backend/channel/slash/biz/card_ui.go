package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
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
	ID     model.ID
	Status enums.CardStatus
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
			return ErrDatabaseOperation
		}

		if holder.AccountID != req.AccountID {
			return ErrResourceNotFound
		}
		products, err := u.cardProductRepository.List(txCtx)
		if err != nil {
			zap.S().Errorw("list slash channel products", "error", err)
			return ErrDatabaseOperation
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
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		product, err := u.cardProductRepository.FindByIDForUpdate(txCtx, productID)
		if err != nil {
			zap.S().Errorw("lock slash channel product", "error", err)
			return ErrDatabaseOperation
		}

		product.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(product.Prefix, product.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance slash card product sequence", "error", err)
			return ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: holder.AccountID,
			Channel:   enums.Channel_Slash,
			Amount:    decimal.Zero,
			Type:      enums.WalletType_Card,
			Currency:  req.Currency,
		}
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create slash card wallet", "error", err)
			return ErrDatabaseOperation
		}

		card = &model.Card{
			AccountID:              holder.AccountID,
			Channel:                enums.Channel_Slash,
			CardProductID:          product.ID,
			CardBin:                product.Prefix,
			CardNumber:             cardNumber,
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
			return ErrDatabaseOperation
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
		return nil, 0, ErrInvalidOperation
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
		return nil, 0, ErrDatabaseOperation
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
		return nil, 0, ErrDatabaseOperation
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
		return nil, ErrDatabaseOperation
	}

	return card, nil
}

func (u *SlashUIUsecase) UpdateCardStatus(ctx context.Context, req *UpdateCardStatusRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCard(txCtx, req.ID); err != nil {
			return err
		}
		var err error
		card, err = u.cardRepository.FindByID(txCtx, req.ID)
		if err != nil {
			zap.S().Errorw("find slash card", "error", err)
			return ErrDatabaseOperation
		}
		card.Status = req.Status
		if err := u.cardRepository.Save(txCtx, card); err != nil {
			zap.S().Errorw("update slash card status", "error", err)
			return ErrDatabaseOperation
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
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	return nil
}
