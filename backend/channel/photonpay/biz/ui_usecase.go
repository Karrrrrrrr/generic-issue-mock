package biz

import (
	"context"
	"crypto/rand"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"

	"github.com/samber/do"
	"go.uber.org/zap"
)

type UIUsecase struct {
	transaction         Transaction
	cardHolderRepo      CardHolderRepository
	cardRepo            CardRepository
	cardProductRepo     CardProductRepository
	cardTransactionRepo CardTransactionRepository
}

func NewUIUsecase(injector *do.Injector) (*UIUsecase, error) {
	return &UIUsecase{
		transaction:         do.MustInvoke[Transaction](injector),
		cardHolderRepo:      do.MustInvoke[CardHolderRepository](injector),
		cardRepo:            do.MustInvoke[CardRepository](injector),
		cardProductRepo:     do.MustInvoke[CardProductRepository](injector),
		cardTransactionRepo: do.MustInvoke[CardTransactionRepository](injector),
	}, nil
}

type UICreateCardHolderRequest struct {
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (u *UIUsecase) CreateCardHolder(ctx context.Context, req *UICreateCardHolderRequest) (*model.CardHolder, error) {
	dateOfBirth := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	holder := &model.CardHolder{
		Channel:                enums.Channel_PhotonPay,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		Mobile:                 req.Mobile,
		MobilePrefix:           "+1",
		DateOfBirth:            &dateOfBirth,
		NationalityCountryCode: "US",
		Status:                 enums.CardHolderStatus_Normal,
		ReviewStatus:           enums.CardHolderReviewStatus_Approved,
	}
	if err := u.cardHolderRepo.Create(ctx, holder); err != nil {
		zap.S().Errorw("create photonpay UI card holder", "error", err)

		return nil, ErrDatabaseOperation
	}

	return holder, nil
}

func (u *UIUsecase) ListCardHolders(ctx context.Context, req *ListRequest) ([]*model.CardHolder, error) {
	holders, err := u.cardHolderRepo.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay UI card holders", "error", err)

		return nil, ErrDatabaseOperation
	}

	return holders, nil
}

type UIOpenCardRequest struct {
	CardHolderID model.ID
	Currency     enums.Currency
	RequestID    string
}

func (u *UIUsecase) OpenCard(ctx context.Context, req *UIOpenCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardHolderRepo.ExistCardHolderByID(txCtx, req.CardHolderID)
		if err != nil {
			zap.S().Errorw("check photonpay UI card holder", "error", err)

			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		productExists, err := u.cardProductRepo.ExistByPrefix(txCtx, photon.DefaultCardBin)
		if err != nil {
			zap.S().Errorw("check photonpay UI card product", "error", err)

			return ErrDatabaseOperation
		}
		if !productExists {
			return ErrResourceNotFound
		}

		product, err := u.cardProductRepo.FindByPrefixForUpdate(txCtx, photon.DefaultCardBin)
		if err != nil {
			zap.S().Errorw("lock photonpay UI card product", "error", err)

			return ErrDatabaseOperation
		}

		product.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(product.Prefix, product.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepo.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance photonpay UI card product sequence", "error", err)

			return ErrDatabaseOperation
		}

		card = &model.Card{
			Channel:                enums.Channel_PhotonPay,
			CardProductID:          product.ID,
			CardBin:                product.Prefix,
			CardNumber:             cardNumber,
			Cvv:                    uiRandomDigits(3),
			ExpireAt:               time.Now().UTC().AddDate(0, 24, 0),
			Status:                 enums.CardStatus_Active,
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

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *UIUsecase) ListCards(ctx context.Context, req *ListRequest) ([]*model.Card, error) {
	cards, err := u.cardRepo.ListCards(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay UI cards", "error", err)

		return nil, ErrDatabaseOperation
	}

	return cards, nil
}

type UIChangeCardStatusRequest struct {
	CardID model.ID
	Status enums.CardStatus
}

func (u *UIUsecase) ChangeCardStatus(ctx context.Context, req *UIChangeCardStatusRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepo.ExistCardByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("check photonpay UI card", "error", err)

			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		card, err = u.cardRepo.FindCardByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("find photonpay UI card", "error", err)

			return ErrDatabaseOperation
		}

		card.Status = req.Status
		if err := u.cardRepo.SaveCard(txCtx, card); err != nil {
			zap.S().Errorw("update photonpay UI card status", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *UIUsecase) ListTransactions(ctx context.Context, req *ListRequest) ([]*model.CardTransaction, error) {
	transactions, err := u.cardTransactionRepo.ListTransactions(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay UI card transactions", "error", err)

		return nil, ErrDatabaseOperation
	}

	return transactions, nil
}

func uiRandomDigits(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	for index := range bytes {
		bytes[index] = '0' + bytes[index]%10
	}

	return string(bytes)
}
