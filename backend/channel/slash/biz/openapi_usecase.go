package biz

import (
	"context"
	"crypto/rand"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"

	"github.com/samber/do"
	"go.uber.org/zap"
)

type OpenAPIUsecase struct {
	transaction               Transaction
	cardHolderRepository      CardHolderRepository
	cardRepository            CardRepository
	cardProductRepository     CardProductRepository
	cardTransactionRepository CardTransactionRepository
}

func NewOpenAPIUsecase(injector *do.Injector) (*OpenAPIUsecase, error) {
	return &OpenAPIUsecase{
		transaction:               do.MustInvokeNamed[Transaction](injector, "slash.transaction"),
		cardHolderRepository:      do.MustInvokeNamed[CardHolderRepository](injector, "slash.card-holder-repository"),
		cardRepository:            do.MustInvokeNamed[CardRepository](injector, "slash.card-repository"),
		cardProductRepository:     do.MustInvokeNamed[CardProductRepository](injector, "slash.card-product-repository"),
		cardTransactionRepository: do.MustInvokeNamed[CardTransactionRepository](injector, "slash.card-transaction-repository"),
	}, nil
}

type OpenAPIListCardsRequest struct {
	Offset int
	Limit  int
	Status enums.CardStatus
}

func (u *OpenAPIUsecase) ListCards(ctx context.Context, req *OpenAPIListCardsRequest) ([]*model.Card, error) {
	items, err := u.cardRepository.List(ctx, &ListCardsRequest{
		Offset: req.Offset,
		Limit:  req.Limit,
		Status: req.Status,
	})
	if err != nil {
		zap.S().Errorw("list slash openapi cards", "error", err)

		return nil, ErrDatabaseOperation
	}

	return items, nil
}

type OpenAPICreateCardRequest struct {
	CardHolderID  model.ID
	CardProductID model.ID
	Currency      enums.Currency
}

func (u *OpenAPIUsecase) CreateCard(ctx context.Context, req *OpenAPICreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		holderExists, err := u.cardHolderRepository.ExistByID(txCtx, req.CardHolderID)
		if err != nil {
			zap.S().Errorw("check slash openapi card holder", "error", err)

			return ErrDatabaseOperation
		}
		if !holderExists {
			return ErrResourceNotFound
		}

		productExists, err := u.cardProductRepository.ExistByID(txCtx, req.CardProductID)
		if err != nil {
			zap.S().Errorw("check slash openapi card product", "error", err)

			return ErrDatabaseOperation
		}
		if !productExists {
			return ErrResourceNotFound
		}

		product, err := u.cardProductRepository.FindByIDForUpdate(txCtx, req.CardProductID)
		if err != nil {
			zap.S().Errorw("lock slash openapi card product", "error", err)

			return ErrDatabaseOperation
		}

		product.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(product.Prefix, product.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance slash openapi card product sequence", "error", err)

			return ErrDatabaseOperation
		}

		card = &model.Card{
			Channel:                enums.Channel_Slash,
			CardProductID:          product.ID,
			CardBin:                product.Prefix,
			CardNumber:             cardNumber,
			Cvv:                    openAPIRandomDigits(3),
			ExpireAt:               time.Now().UTC().AddDate(2, 0, 0),
			Status:                 enums.CardStatus_Active,
			CardHolderID:           req.CardHolderID,
			FormType:               enums.CardFormType_Virtual,
			CardCurrency:           req.Currency,
			CardScheme:             "VISA",
			CardType:               enums.CardType_Single,
			RequestID:              openAPIRandomDigits(20),
			LastOperationRequestID: openAPIRandomDigits(20),
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

func (u *OpenAPIUsecase) GetCard(ctx context.Context, id model.ID) (*model.Card, error) {
	exists, err := u.cardRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash openapi card", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item, err := u.cardRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash openapi card", "error", err)

		return nil, ErrDatabaseOperation
	}

	return item, nil
}

type OpenAPIUpdateCardRequest struct {
	ID     model.ID
	Status enums.CardStatus
}

func (u *OpenAPIUsecase) UpdateCard(ctx context.Context, req *OpenAPIUpdateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistByID(txCtx, req.ID)
		if err != nil {
			zap.S().Errorw("check slash openapi card", "error", err)

			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		card, err = u.cardRepository.FindByID(txCtx, req.ID)
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

func (u *OpenAPIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	items, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list slash openapi card products", "error", err)

		return nil, ErrDatabaseOperation
	}

	return items, nil
}

type OpenAPIListTransactionsRequest struct {
	Offset int
	Limit  int
	CardID model.ID
}

func (u *OpenAPIUsecase) ListTransactions(ctx context.Context, req *OpenAPIListTransactionsRequest) ([]*model.CardTransaction, error) {
	items, err := u.cardTransactionRepository.List(ctx, &ListCardTransactionsRequest{
		Offset: req.Offset,
		Limit:  req.Limit,
		CardID: req.CardID,
	})
	if err != nil {
		zap.S().Errorw("list slash openapi transactions", "error", err)

		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *OpenAPIUsecase) GetTransaction(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	exists, err := u.cardTransactionRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash openapi transaction", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item, err := u.cardTransactionRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash openapi transaction", "error", err)

		return nil, ErrDatabaseOperation
	}

	return item, nil
}

func openAPIRandomDigits(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	for index := range bytes {
		bytes[index] = '0' + bytes[index]%10
	}

	return string(bytes)
}
