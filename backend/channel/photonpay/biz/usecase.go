package biz

import (
	"context"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Transaction interface {
	InTx(context.Context, func(context.Context) error) error
}

type CardHolderRepository interface {
	Create(context.Context, *model.CardHolder) error
	ExistCardHolderByID(context.Context, model.ID) (bool, error)
	FindCardHolderByID(context.Context, model.ID) (*model.CardHolder, error)
	Save(context.Context, *model.CardHolder) error
	List(context.Context, *ListRequest) ([]*model.CardHolder, error)
}

type CardRepository interface {
	CreateCard(context.Context, *model.Card) error
	ExistCardByID(context.Context, model.ID) (bool, error)
	FindCardByID(context.Context, model.ID) (*model.Card, error)
	ExistCardByRequestID(context.Context, string) (bool, error)
	FindByRequestID(context.Context, string) (*model.Card, error)
	ListCards(context.Context, *ListRequest) ([]*model.Card, error)
	SaveCard(context.Context, *model.Card) error
}

type CardProductRepository interface {
	ExistByPrefix(context.Context, string) (bool, error)
	FindByPrefixForUpdate(context.Context, string) (*model.CardProduct, error)
	List(context.Context) ([]*model.CardProduct, error)
	Save(context.Context, *model.CardProduct) error
}

type CardTransactionRepository interface {
	Create(context.Context, *model.CardTransaction) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardTransaction, error)
	ListTransactions(context.Context, *ListRequest) ([]*model.CardTransaction, error)
}

type AuthorizationRepository interface {
	Create(context.Context, *model.Authorization) error
}

type PhotonPayOpenAPIUsecase struct {
	transaction         Transaction
	cardHolderRepo      CardHolderRepository
	cardRepo            CardRepository
	cardProductRepo     CardProductRepository
	cardTransactionRepo CardTransactionRepository
}

func NewPhotonPayOpenAPIUsecase(injector *do.Injector) (*PhotonPayOpenAPIUsecase, error) {
	return &PhotonPayOpenAPIUsecase{
		transaction:         do.MustInvoke[Transaction](injector),
		cardHolderRepo:      do.MustInvoke[CardHolderRepository](injector),
		cardRepo:            do.MustInvoke[CardRepository](injector),
		cardProductRepo:     do.MustInvoke[CardProductRepository](injector),
		cardTransactionRepo: do.MustInvoke[CardTransactionRepository](injector),
	}, nil
}

type CreateCardHolderRequest struct {
	FirstName              string
	LastName               string
	Email                  string
	Mobile                 string
	MobilePrefix           string
	DateOfBirth            *time.Time
	NationalityCountryCode string
	ResidentialAddress     string
	ResidentialCity        string
	ResidentialCountryCode string
	ResidentialPostalCode  string
	ResidentialState       string
	CertType               string
	CertCountryCode        string
	CertID                 string
	Portrait               string
	ReverseSide            string
}

func (u *PhotonPayOpenAPIUsecase) CreateCardHolder(ctx context.Context, req *CreateCardHolderRequest) (*model.CardHolder, error) {
	holder := &model.CardHolder{
		Channel:                common.Channel_PhotonPay,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		Mobile:                 req.Mobile,
		MobilePrefix:           req.MobilePrefix,
		DateOfBirth:            req.DateOfBirth,
		NationalityCountryCode: req.NationalityCountryCode,
		ResidentialAddress:     req.ResidentialAddress,
		ResidentialCity:        req.ResidentialCity,
		ResidentialCountryCode: req.ResidentialCountryCode,
		ResidentialPostalCode:  req.ResidentialPostalCode,
		ResidentialState:       req.ResidentialState,
		CertType:               req.CertType,
		CertCountryCode:        req.CertCountryCode,
		CertID:                 req.CertID,
		Portrait:               req.Portrait,
		ReverseSide:            req.ReverseSide,
		Status:                 common.CardHolderStatus_Normal,
		ReviewStatus:           common.CardHolderReviewStatus_Approved,
	}
	if err := u.cardHolderRepo.Create(ctx, holder); err != nil {
		zap.S().Errorw("create photonpay card holder", "error", err)

		return nil, ErrDatabaseOperation
	}
	return holder, nil
}

type UpdateCardHolderRequest struct {
	CardholderID model.ID
	Email        *string
	Mobile       *string
	MobilePrefix *string
}

func (u *PhotonPayOpenAPIUsecase) UpdateCardHolder(ctx context.Context, req *UpdateCardHolderRequest) (*model.CardHolder, error) {
	var holder *model.CardHolder
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCardHolder(txCtx, req.CardholderID); err != nil {
			return err
		}

		var err error
		holder, err = u.cardHolderRepo.FindCardHolderByID(txCtx, req.CardholderID)
		if err != nil {
			zap.S().Errorw("find photonpay card holder", "error", err)

			return ErrDatabaseOperation
		}

		if req.Email != nil {
			holder.Email = *req.Email
		}
		if req.Mobile != nil {
			holder.Mobile = *req.Mobile
		}
		if req.MobilePrefix != nil {
			holder.MobilePrefix = *req.MobilePrefix
		}

		if err := u.cardHolderRepo.Save(txCtx, holder); err != nil {
			zap.S().Errorw("update photonpay card holder", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return holder, nil
}

type ListRequest struct {
	Offset int
	Limit  int
}

func (u *PhotonPayOpenAPIUsecase) ListCardHolders(ctx context.Context, req *ListRequest) ([]*model.CardHolder, error) {
	holders, err := u.cardHolderRepo.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay card holders", "error", err)

		return nil, ErrDatabaseOperation
	}

	return holders, nil
}

type OpenCardRequest struct {
	CardBin          string
	Currency         common.Currency
	CardScheme       string
	CardType         photon.CardType
	CardFormFactor   photon.CardFormFactor
	CardholderID     model.ID
	RequestID        string
	ExpirationMonths int
}

func (u *PhotonPayOpenAPIUsecase) OpenCard(ctx context.Context, req *OpenCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCardHolder(txCtx, req.CardholderID); err != nil {
			return err
		}
		product, err := u.getCardProductByBinPrefix(txCtx, req.CardBin)
		if err != nil {
			return err
		}
		product.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(product.Prefix, product.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepo.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance photonpay card product sequence", "error", err)

			return ErrDatabaseOperation
		}

		months := req.ExpirationMonths
		if months == 0 {
			months = 24
		}
		card = &model.Card{
			Channel:                common.Channel_PhotonPay,
			CardProductID:          product.ID,
			CardBin:                product.Prefix,
			CardNumber:             cardNumber,
			Cvv:                    randomx.Digits(3),
			ExpireAt:               time.Now().UTC().AddDate(0, months, 0),
			Status:                 common.CardStatus_Active,
			CardHolderID:           req.CardholderID,
			FormType:               photon.CardFormFactorToGeneric(req.CardFormFactor),
			RequestID:              req.RequestID,
			LastOperationRequestID: req.RequestID,
			LastOperationType:      common.OperationType_OpenCard,
			LastOperationStatus:    common.OperationStatus_Succeed,
			CardCurrency:           req.Currency,
			CardScheme:             req.CardScheme,
			CardType:               photon.CardTypeToGeneric(req.CardType),
		}

		if err := u.cardRepo.CreateCard(txCtx, card); err != nil {
			zap.S().Errorw("create photonpay card", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PhotonPayOpenAPIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	products, err := u.cardProductRepo.List(ctx)
	if err != nil {
		zap.S().Errorw("list photonpay card products", "error", err)
		return nil, ErrDatabaseOperation
	}

	return products, nil
}

func (u *PhotonPayOpenAPIUsecase) GetCard(ctx context.Context, cardID model.ID) (*model.Card, error) {
	if err := u.requireCard(ctx, cardID); err != nil {
		return nil, err
	}

	card, err := u.cardRepo.FindCardByID(ctx, cardID)
	if err != nil {
		zap.S().Errorw("find photonpay card", "error", err)

		return nil, ErrDatabaseOperation
	}

	return card, nil
}

func (u *PhotonPayOpenAPIUsecase) ListCards(ctx context.Context, req *ListRequest) ([]*model.Card, error) {
	cards, err := u.cardRepo.ListCards(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay cards", "error", err)

		return nil, ErrDatabaseOperation
	}

	return cards, nil
}

func (u *PhotonPayOpenAPIUsecase) GetRequestResult(ctx context.Context, requestID string) (*model.Card, error) {
	exists, err := u.cardRepo.ExistCardByRequestID(ctx, requestID)
	if err != nil {
		zap.S().Errorw("check photonpay card request", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	card, err := u.cardRepo.FindByRequestID(ctx, requestID)
	if err != nil {
		zap.S().Errorw("find photonpay card request", "error", err)

		return nil, ErrDatabaseOperation
	}

	return card, nil
}

type ChangeCardStatusRequest struct {
	CardID    model.ID
	RequestID string
	Status    common.CardStatus
	Operation common.OperationType
}

func (u *PhotonPayOpenAPIUsecase) ChangeCardStatus(ctx context.Context, req *ChangeCardStatusRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCard(txCtx, req.CardID); err != nil {
			return err
		}

		var err error
		card, err = u.cardRepo.FindCardByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("find photonpay card", "error", err)

			return ErrDatabaseOperation
		}

		card.Status = req.Status
		card.LastOperationRequestID = req.RequestID
		card.LastOperationType = req.Operation
		card.LastOperationStatus = common.OperationStatus_Succeed

		if err := u.cardRepo.SaveCard(txCtx, card); err != nil {
			zap.S().Errorw("update photonpay card status", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PhotonPayOpenAPIUsecase) ListTransactions(ctx context.Context, req *ListRequest) ([]*model.CardTransaction, error) {
	transactions, err := u.cardTransactionRepo.ListTransactions(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay card transactions", "error", err)

		return nil, ErrDatabaseOperation
	}

	return transactions, nil
}

func DefaultBalance() decimal.Decimal {
	return decimal.NewFromInt(1_000_000)
}

func (u *PhotonPayOpenAPIUsecase) requireCardHolder(ctx context.Context, id model.ID) error {
	exists, err := u.cardHolderRepo.ExistCardHolderByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay card holder", "error", err)

		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}

func (u *PhotonPayOpenAPIUsecase) requireCard(ctx context.Context, id model.ID) error {
	exists, err := u.cardRepo.ExistCardByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay card", "error", err)

		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}

func (u *PhotonPayOpenAPIUsecase) getCardProductByBinPrefix(
	ctx context.Context,
	prefix string,
) (*model.CardProduct, error) {
	exists, err := u.cardProductRepo.ExistByPrefix(ctx, prefix)
	if err != nil {
		zap.S().Errorw("check photonpay card product", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	product, err := u.cardProductRepo.FindByPrefixForUpdate(ctx, prefix)
	if err != nil {
		zap.S().Errorw("lock photonpay card product", "error", err)
		return nil, ErrDatabaseOperation
	}

	return product, nil
}
