package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type OpenAPIListCardsRequest struct {
	AccountID model.ID
	Offset    int
	Limit     int
	Status    *enums.CardStatus
}

type OpenAPICreateCardRequest struct {
	AccountID     model.ID
	CardHolderID  model.ID
	CardProductID model.ID
	Currency      enums.Currency
	RequestID     string
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

		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *SlashOpenAPIUsecase) CreateCard(ctx context.Context, req *OpenAPICreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if req.CardHolderID != 0 {
			holderExists, err := u.cardHolderRepository.ExistByID(txCtx, req.CardHolderID)
			if err != nil {
				zap.S().Errorw("check slash openapi card holder", "error", err)

				return ErrDatabaseOperation
			}
			if !holderExists {
				return ErrResourceNotFound
			}
		}

		productExists, err := u.cardProductRepository.ExistByID(txCtx, req.CardProductID)
		if err != nil {
			zap.S().Errorw("check slash openapi card product", "error", err)

			return ErrDatabaseOperation
		}
		if !productExists {
			return ErrResourceNotFound
		}

		cardProduct, err := u.cardProductRepository.FindByIDForUpdate(txCtx, req.CardProductID)
		if err != nil {
			zap.S().Errorw("lock slash openapi card product", "error", err)

			return ErrDatabaseOperation
		}

		cardProduct.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(cardProduct.Prefix, cardProduct.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, cardProduct); err != nil {
			zap.S().Errorw("advance slash openapi card product sequence", "error", err)

			return ErrDatabaseOperation
		}

		card = &model.Card{
			Channel:                enums.Channel_Slash,
			AccountID:              req.AccountID,
			CardProductID:          cardProduct.ID,
			CardBin:                cardProduct.Prefix,
			CardNumber:             cardNumber,
			Cvv:                    randomx.Digits(3),
			ExpireAt:               time.Now().UTC().AddDate(2, 0, 0),
			Status:                 enums.CardStatus_Active,
			CardHolderID:           req.CardHolderID,
			FormType:               enums.CardFormType_Virtual,
			CardCurrency:           req.Currency,
			CardScheme:             enums.CardScheme_Visa,
			CardType:               enums.CardType_Single,
			RequestID:              req.RequestID,
			LastOperationRequestID: req.RequestID,
		}
		if err := u.cardRepository.Create(txCtx, card); err != nil {
			zap.S().Errorw("create slash openapi card", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *SlashOpenAPIUsecase) GetCard(ctx context.Context, req *ResourceRequest) (*model.Card, error) {
	exists, err := u.cardRepository.ExistByAccountID(ctx, (*CardExistByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check slash openapi card", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item, err := u.cardRepository.FindByAccountID(ctx, (*CardFindByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("find slash openapi card", "error", err)

		return nil, ErrDatabaseOperation
	}

	return item, nil
}

func (u *SlashOpenAPIUsecase) UpdateCard(ctx context.Context, req *OpenAPIUpdateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		resource := &ResourceRequest{AccountID: &req.AccountID, ID: req.ID}
		exists, err := u.cardRepository.ExistByAccountID(txCtx, (*CardExistByAccountIDRequest)(resource))
		if err != nil {
			zap.S().Errorw("check slash openapi card", "error", err)

			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		card, err = u.cardRepository.FindByAccountID(txCtx, (*CardFindByAccountIDRequest)(resource))
		if err != nil {
			zap.S().Errorw("find slash openapi card", "error", err)

			return ErrDatabaseOperation
		}

		card.Status = req.Status
		if err := u.cardRepository.Save(txCtx, card); err != nil {
			zap.S().Errorw("update slash openapi card", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}
