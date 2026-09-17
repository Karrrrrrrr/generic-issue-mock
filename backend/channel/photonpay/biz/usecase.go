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

type PhotonPayTransaction interface {
	InTx(context.Context, func(context.Context) error) error
}

type AccountRepository interface {
	Create(context.Context, *model.Account) error
	Exist(context.Context, model.ID) (bool, error)
	Find(context.Context, model.ID) (*model.Account, error)
	Count(context.Context) (int64, error)
	List(context.Context, *ListRequest) ([]*model.Account, error)
	Save(context.Context, *model.Account) error
}

type WalletRepository interface {
	Create(context.Context, *model.Wallet) error
	FindByIDForUpdate(context.Context, *ResourceRequest) (*model.Wallet, error)
	Save(context.Context, *model.Wallet) error
}

type CardHolderRepository interface {
	Create(context.Context, *model.CardHolder) error
	ExistCardHolderByID(context.Context, model.ID) (bool, error)
	FindCardHolderByID(context.Context, model.ID) (*model.CardHolder, error)
	ExistCardHolderByAccountID(context.Context, *ResourceRequest) (bool, error)
	FindCardHolderByAccountID(context.Context, *ResourceRequest) (*model.CardHolder, error)
	Save(context.Context, *model.CardHolder) error
	List(context.Context, *ListRequest) ([]*model.CardHolder, error)
}

type CardRepository interface {
	CreateCard(context.Context, *model.Card) error
	ExistCardByID(context.Context, model.ID) (bool, error)
	FindCardByID(context.Context, model.ID) (*model.Card, error)
	ExistCardByRequestID(context.Context, string) (bool, error)
	FindByRequestID(context.Context, string) (*model.Card, error)
	ExistCardByLastOperationRequestID(context.Context, string) (bool, error)
	FindByLastOperationRequestID(context.Context, string) (*model.Card, error)
	ExistCardByRequestIDForAccount(context.Context, *RequestResultResourceRequest) (bool, error)
	FindCardByRequestIDForAccount(context.Context, *RequestResultResourceRequest) (*model.Card, error)
	ExistCardByLastOperationRequestIDForAccount(context.Context, *RequestResultResourceRequest) (bool, error)
	FindCardByLastOperationRequestIDForAccount(context.Context, *RequestResultResourceRequest) (*model.Card, error)
	ListCards(context.Context, *ListRequest) ([]*model.Card, error)
	SaveCard(context.Context, *model.Card) error
	ExistCardByAccountID(context.Context, *ResourceRequest) (bool, error)
	FindCardByAccountID(context.Context, *ResourceRequest) (*model.Card, error)
}

type CardProductRepository interface {
	ExistByPrefix(context.Context, string) (bool, error)
	FindByPrefixForUpdate(context.Context, string) (*model.CardProduct, error)
	ExistByPrefixForAccount(context.Context, *CardProductResourceRequest) (bool, error)
	FindByPrefixForAccountForUpdate(context.Context, *CardProductResourceRequest) (*model.CardProduct, error)
	List(context.Context) ([]*model.CardProduct, error)
	Save(context.Context, *model.CardProduct) error
	ListByAccountID(context.Context, model.ID) ([]*model.CardProduct, error)
}

type CardTransactionRepository interface {
	Create(context.Context, *model.CardTransaction) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardTransaction, error)
	ExistByAccountID(context.Context, *ResourceRequest) (bool, error)
	FindByAccountID(context.Context, *ResourceRequest) (*model.CardTransaction, error)
	ListTransactions(context.Context, *ListRequest) ([]*model.CardTransaction, error)
}

type AuthorizationRepository interface {
	Create(context.Context, *model.Authorization) error
	List(context.Context, *ListRequest) ([]*model.Authorization, error)
}

type WebhookConfigRepository interface {
	Create(context.Context, *model.WebhookConfig) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.WebhookConfig, error)
	List(context.Context) ([]*model.WebhookConfig, error)
	Save(context.Context, *model.WebhookConfig) error
	Delete(context.Context, model.ID) error
}

type WebhookRecordRepository interface {
	Create(context.Context, *model.WebhookRecord) error
	Exist(context.Context, model.ID) (bool, error)
	Find(context.Context, model.ID) (*model.WebhookRecord, error)
	Count(context.Context) (int64, error)
	List(context.Context, *ListRequest) ([]*model.WebhookRecord, error)
	Save(context.Context, *model.WebhookRecord) error
}

type WebhookDeliveryResult struct {
	StatusCode      int
	ResponseBody    string
	RequestHeaders  []byte
	ResponseHeaders []byte
}

type WebhookClient interface {
	Deliver(context.Context, *PhotonPayWebhookDeliveryRequest) (*WebhookDeliveryResult, error)
}

type PhotonPayWebhookDeliveryRequest struct {
	TargetURL      string
	Payload        []byte
	RequestHeaders []byte
	NotifyCategory string
	NotifyType     string
	PublishedAt    string
}

type PhotonPayOpenAPIUsecase struct {
	transaction         PhotonPayTransaction
	cardHolderRepo      CardHolderRepository
	cardRepo            CardRepository
	cardProductRepo     CardProductRepository
	authorizationRepo   AuthorizationRepository
	cardTransactionRepo CardTransactionRepository
}

func NewPhotonPayOpenAPIUsecase(injector *do.Injector) (*PhotonPayOpenAPIUsecase, error) {
	return &PhotonPayOpenAPIUsecase{
		transaction:         do.MustInvoke[PhotonPayTransaction](injector),
		cardHolderRepo:      do.MustInvoke[CardHolderRepository](injector),
		cardRepo:            do.MustInvoke[CardRepository](injector),
		cardProductRepo:     do.MustInvoke[CardProductRepository](injector),
		authorizationRepo:   do.MustInvoke[AuthorizationRepository](injector),
		cardTransactionRepo: do.MustInvoke[CardTransactionRepository](injector),
	}, nil
}

type CreateCardHolderRequest struct {
	AccountID              model.ID
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
		AccountID:              req.AccountID,
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
	AccountID    model.ID
	CardholderID model.ID
	Email        *string
	Mobile       *string
	MobilePrefix *string
}

func (u *PhotonPayOpenAPIUsecase) UpdateCardHolder(ctx context.Context, req *UpdateCardHolderRequest) (*model.CardHolder, error) {
	var holder *model.CardHolder
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		resource := &ResourceRequest{AccountID: &req.AccountID, ID: req.CardholderID}
		if err := u.requireCardHolder(txCtx, resource); err != nil {
			return err
		}

		var err error
		holder, err = u.cardHolderRepo.FindCardHolderByAccountID(txCtx, resource)
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
	AccountID model.ID
	Offset    int
	Limit     int
}

type ResourceRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardProductResourceRequest struct {
	AccountID model.ID
	Prefix    string
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
	AccountID        model.ID
	CardBin          string
	Currency         common.Currency
	CardScheme       common.CardScheme
	CardType         photon.CardType
	CardFormFactor   photon.CardFormFactor
	CardholderID     model.ID
	RequestID        string
	ExpirationMonths int
}

func (u *PhotonPayOpenAPIUsecase) OpenCard(ctx context.Context, req *OpenCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		holderResource := &ResourceRequest{AccountID: &req.AccountID, ID: req.CardholderID}
		if err := u.requireCardHolder(txCtx, holderResource); err != nil {
			return err
		}
		product, err := u.getCardProductByBinPrefix(txCtx, &CardProductResourceRequest{
			AccountID: req.AccountID,
			Prefix:    req.CardBin,
		})
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
			AccountID:              req.AccountID,
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

func (u *PhotonPayOpenAPIUsecase) ListCardProducts(ctx context.Context, accountID model.ID) ([]*model.CardProduct, error) {
	products, err := u.cardProductRepo.ListByAccountID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("list photonpay card products", "error", err)
		return nil, ErrDatabaseOperation
	}

	return products, nil
}

func (u *PhotonPayOpenAPIUsecase) GetCard(ctx context.Context, req *ResourceRequest) (*model.Card, error) {
	if err := u.requireCard(ctx, req); err != nil {
		return nil, err
	}

	card, err := u.cardRepo.FindCardByAccountID(ctx, req)
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

type RequestResultResourceRequest struct {
	AccountID model.ID
	RequestID string
}

func (u *PhotonPayOpenAPIUsecase) GetRequestResult(ctx context.Context, req *RequestResultResourceRequest) (*model.Card, error) {
	exists, err := u.cardRepo.ExistCardByRequestIDForAccount(ctx, req)
	if err != nil {
		zap.S().Errorw("check photonpay card request", "error", err)

		return nil, ErrDatabaseOperation
	}
	if exists {
		card, err := u.cardRepo.FindCardByRequestIDForAccount(ctx, req)
		if err != nil {
			zap.S().Errorw("find photonpay card request", "error", err)

			return nil, ErrDatabaseOperation
		}

		return card, nil
	}

	exists, err = u.cardRepo.ExistCardByLastOperationRequestIDForAccount(ctx, req)
	if err != nil {
		zap.S().Errorw("check photonpay card operation request", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	card, err := u.cardRepo.FindCardByLastOperationRequestIDForAccount(ctx, req)
	if err != nil {
		zap.S().Errorw("find photonpay card operation request", "error", err)

		return nil, ErrDatabaseOperation
	}

	return card, nil
}

type ChangeCardStatusRequest struct {
	AccountID model.ID
	CardID    model.ID
	RequestID string
	Status    common.CardStatus
	Operation common.OperationType
}

func (u *PhotonPayOpenAPIUsecase) ChangeCardStatus(ctx context.Context, req *ChangeCardStatusRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCard(txCtx, &ResourceRequest{AccountID: &req.AccountID, ID: req.CardID}); err != nil {
			return err
		}

		var err error
		card, err = u.cardRepo.FindCardByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("find photonpay card", "error", err)

			return ErrDatabaseOperation
		}

		card.Status = req.Status
		if req.RequestID != "" {
			card.LastOperationRequestID = req.RequestID
		} else {
			card.LastOperationRequestID = randomx.Digits(20)
		}
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

type UpdateCardRequest struct {
	AccountID model.ID
	CardID    model.ID
	RequestID string
}

func (u *PhotonPayOpenAPIUsecase) UpdateCard(ctx context.Context, req *UpdateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCard(txCtx, &ResourceRequest{AccountID: &req.AccountID, ID: req.CardID}); err != nil {
			return err
		}

		var err error
		card, err = u.cardRepo.FindCardByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("find photonpay card for update", "error", err)

			return ErrDatabaseOperation
		}

		card.LastOperationRequestID = req.RequestID
		card.LastOperationType = common.OperationType_UpdateCard
		card.LastOperationStatus = common.OperationStatus_Succeed
		if err := u.cardRepo.SaveCard(txCtx, card); err != nil {
			zap.S().Errorw("update photonpay card", "error", err)

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

type SandboxTransactionRequest struct {
	AccountID           model.ID
	RequestID           string
	CardID              model.ID
	OriginTransactionID model.ID
	Currency            common.Currency
	Amount              decimal.Decimal
	Type                photon.SandboxTransactionType
	MerchantName        string
	MerchantCountry     string
	MerchantMCC         string
}

func (u *PhotonPayOpenAPIUsecase) SandboxTransaction(ctx context.Context, req *SandboxTransactionRequest) error {
	return u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCard(txCtx, &ResourceRequest{AccountID: &req.AccountID, ID: req.CardID}); err != nil {
			return err
		}
		card, err := u.cardRepo.FindCardByAccountID(txCtx, &ResourceRequest{
			AccountID: &req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("find photonpay sandbox card", "error", err)
			return ErrDatabaseOperation
		}

		transactionType := photon.SandboxTransactionTypeToGeneric(req.Type)
		var originTransaction *model.CardTransaction
		if transactionType != common.CardTransactionType_AUTH {
			originResource := &ResourceRequest{AccountID: &req.AccountID, ID: req.OriginTransactionID}
			exists, err := u.cardTransactionRepo.ExistByAccountID(txCtx, originResource)
			if err != nil {
				zap.S().Errorw("check photonpay origin transaction", "error", err)

				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}

			originTransaction, err = u.cardTransactionRepo.FindByAccountID(txCtx, originResource)
			if err != nil {
				zap.S().Errorw("find photonpay origin transaction", "error", err)

				return ErrDatabaseOperation
			}
		}

		var authorizationID model.ID
		if transactionType == common.CardTransactionType_AUTH {
			authorization := &model.Authorization{
				AccountID:         card.AccountID,
				Channel:           common.Channel_PhotonPay,
				CardID:            req.CardID,
				Currency:          req.Currency,
				Amount:            req.Amount,
				MerchantName:      req.MerchantName,
				MerchantCountry:   req.MerchantCountry,
				MerchantMCC:       req.MerchantMCC,
				AuthorizationCode: randomx.Digits(6),
				Status:            common.TransactionStatus_AUTHORIZED,
			}
			if err := u.authorizationRepo.Create(txCtx, authorization); err != nil {
				zap.S().Errorw("create photonpay authorization", "error", err)

				return ErrDatabaseOperation
			}

			authorizationID = authorization.ID
		} else {
			authorizationID = originTransaction.AuthorizationID
		}

		transaction := &model.CardTransaction{
			AccountID:               card.AccountID,
			Channel:                 common.Channel_PhotonPay,
			OriginCardTransactionID: req.OriginTransactionID,
			AuthorizationID:         authorizationID,
			CardID:                  req.CardID,
			Status:                  common.TransactionStatus_SUCCEED,
			Type:                    transactionType,
			Currency:                req.Currency,
			TxAmount:                req.Amount,
			TxCurrency:              req.Currency,
			RequestID:               req.RequestID,
			MerchantName:            req.MerchantName,
			MerchantCountry:         req.MerchantCountry,
			MerchantMCC:             req.MerchantMCC,
		}
		if err := u.cardTransactionRepo.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create photonpay sandbox transaction", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
}

func DefaultBalance() decimal.Decimal {
	return decimal.NewFromInt(1_000_000)
}

func (u *PhotonPayOpenAPIUsecase) requireCardHolder(
	ctx context.Context,
	req *ResourceRequest,
) error {
	exists, err := u.cardHolderRepo.ExistCardHolderByAccountID(ctx, req)
	if err != nil {
		zap.S().Errorw("check photonpay card holder", "error", err)

		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}

func (u *PhotonPayOpenAPIUsecase) requireCard(ctx context.Context, req *ResourceRequest) error {
	exists, err := u.cardRepo.ExistCardByAccountID(ctx, req)
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
	req *CardProductResourceRequest,
) (*model.CardProduct, error) {
	exists, err := u.cardProductRepo.ExistByPrefixForAccount(ctx, req)
	if err != nil {
		zap.S().Errorw("check photonpay card product", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	product, err := u.cardProductRepo.FindByPrefixForAccountForUpdate(ctx, req)
	if err != nil {
		zap.S().Errorw("lock photonpay card product", "error", err)
		return nil, ErrDatabaseOperation
	}

	return product, nil
}
