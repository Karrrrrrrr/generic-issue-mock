package biz

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/shopspring/decimal"
)

var ErrNotFound = errors.New("photonpay resource not found")

type CardHolderRepository interface {
	Create(context.Context, *model.CardHolder) error
	FindCardHolderByID(context.Context, model.ID) (*model.CardHolder, error)
	Save(context.Context, *model.CardHolder) error
	List(context.Context, int, int) ([]*model.CardHolder, error)
}

type CardRepository interface {
	CreateCard(context.Context, *model.Card) error
	FindCardByID(context.Context, model.ID) (*model.Card, error)
	FindByRequestID(context.Context, string) (*model.Card, error)
	SaveCard(context.Context, *model.Card) error
}

type TransactionRepository interface {
	ListTransactions(context.Context, int, int) ([]*model.CardTransaction, error)
}

type Usecase struct {
	cardHolderRepo  CardHolderRepository
	cardRepo        CardRepository
	transactionRepo TransactionRepository
}

func NewUsecase(
	cardHolderRepo CardHolderRepository,
	cardRepo CardRepository,
	transactionRepo TransactionRepository,
) *Usecase {
	return &Usecase{
		cardHolderRepo:  cardHolderRepo,
		cardRepo:        cardRepo,
		transactionRepo: transactionRepo,
	}
}

type CreateCardHolderRequest struct {
	FirstName              string
	LastName               string
	Email                  string
	Mobile                 string
	MobilePrefix           string
	DateOfBirth            string
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

func (u *Usecase) CreateCardHolder(ctx context.Context, req *CreateCardHolderRequest) (*model.CardHolder, error) {
	holder := &model.CardHolder{
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
		return nil, err
	}
	return holder, nil
}

func (u *Usecase) UpdateCardHolder(ctx context.Context, holderID model.ID, email *string, mobile *string, mobilePrefix *string) (*model.CardHolder, error) {
	holder, err := u.cardHolderRepo.FindCardHolderByID(ctx, holderID)
	if err != nil {
		return nil, err
	}
	if email != nil {
		holder.Email = *email
	}
	if mobile != nil {
		holder.Mobile = *mobile
	}
	if mobilePrefix != nil {
		holder.MobilePrefix = *mobilePrefix
	}
	if err := u.cardHolderRepo.Save(ctx, holder); err != nil {
		return nil, err
	}
	return holder, nil
}

func (u *Usecase) ListCardHolders(ctx context.Context, page int, size int) ([]*model.CardHolder, error) {
	return u.cardHolderRepo.List(ctx, (page-1)*size, size)
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

func (u *Usecase) OpenCard(ctx context.Context, req *OpenCardRequest) (*model.Card, error) {
	holder, err := u.cardHolderRepo.FindCardHolderByID(ctx, req.CardholderID)
	if err != nil {
		return nil, err
	}
	months := req.ExpirationMonths
	if months == 0 {
		months = 24
	}
	card := &model.Card{
		CardBin:                req.CardBin,
		CardNumber:             req.CardBin + randomDigits(10),
		Cvv:                    randomDigits(3),
		ExpireTime:             time.Now().UTC().AddDate(0, months, 0).Format("01/06"),
		Status:                 common.CardStatus_Active,
		CardHolderID:           holder.ID,
		FormType:               photon.CardFormFactorToGeneric(req.CardFormFactor),
		RequestID:              req.RequestID,
		LastOperationRequestID: req.RequestID,
		LastOperationType:      common.OperationType_OpenCard,
		LastOperationStatus:    common.OperationStatus_Succeed,
		CardCurrency:           req.Currency,
		CardScheme:             req.CardScheme,
		CardType:               photon.CardTypeToGeneric(req.CardType),
	}
	if err := u.cardRepo.CreateCard(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}

func (u *Usecase) GetCard(ctx context.Context, cardID model.ID) (*model.Card, error) {
	return u.cardRepo.FindCardByID(ctx, cardID)
}

func (u *Usecase) GetRequestResult(ctx context.Context, requestID string) (*model.Card, error) {
	return u.cardRepo.FindByRequestID(ctx, requestID)
}

func (u *Usecase) ChangeCardStatus(ctx context.Context, cardID model.ID, requestID string, status common.CardStatus, operation common.OperationType) (*model.Card, error) {
	card, err := u.cardRepo.FindCardByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	card.Status = status
	card.LastOperationRequestID = requestID
	card.LastOperationType = operation
	card.LastOperationStatus = common.OperationStatus_Succeed
	if err := u.cardRepo.SaveCard(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}

func (u *Usecase) ListTransactions(ctx context.Context, page int, size int) ([]*model.CardTransaction, error) {
	return u.transactionRepo.ListTransactions(ctx, (page-1)*size, size)
}

func DefaultBalance() decimal.Decimal {
	return decimal.NewFromInt(1_000_000)
}

func randomDigits(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	for index := range bytes {
		bytes[index] = '0' + bytes[index]%10
	}
	return string(bytes)
}
