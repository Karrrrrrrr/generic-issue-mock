package biz

import (
	"context"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type PhotonPayUIUsecase struct {
	transaction         PhotonPayTransaction
	cardHolderRepo      CardHolderRepository
	cardRepo            CardRepository
	cardProductRepo     CardProductRepository
	authorizationRepo   AuthorizationRepository
	cardTransactionRepo CardTransactionRepository
	webhookRepo         WebhookConfigRepository
}

func NewPhotonPayUIUsecase(injector *do.Injector) (*PhotonPayUIUsecase, error) {
	return &PhotonPayUIUsecase{
		transaction:         do.MustInvoke[PhotonPayTransaction](injector),
		cardHolderRepo:      do.MustInvoke[CardHolderRepository](injector),
		cardRepo:            do.MustInvoke[CardRepository](injector),
		cardProductRepo:     do.MustInvoke[CardProductRepository](injector),
		authorizationRepo:   do.MustInvoke[AuthorizationRepository](injector),
		cardTransactionRepo: do.MustInvoke[CardTransactionRepository](injector),
		webhookRepo:         do.MustInvoke[WebhookConfigRepository](injector),
	}, nil
}

type UICreateWebhookRequest struct {
	Event, TargetURL string
	Enabled          bool
}
type UIUpdateWebhookRequest struct {
	ID        model.ID
	TargetURL string
	Enabled   bool
}

func (u *PhotonPayUIUsecase) CreateWebhook(ctx context.Context, req *UICreateWebhookRequest) (*model.WebhookConfig, error) {
	item := &model.WebhookConfig{Channel: enums.Channel_PhotonPay, Event: req.Event, TargetURL: req.TargetURL, Enabled: req.Enabled}
	if err := u.webhookRepo.Create(ctx, item); err != nil {
		zap.S().Errorw("create photonpay UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}
func (u *PhotonPayUIUsecase) ListWebhooks(ctx context.Context) ([]*model.WebhookConfig, error) {
	items, err := u.webhookRepo.List(ctx)
	if err != nil {
		zap.S().Errorw("list photonpay UI webhooks", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}
func (u *PhotonPayUIUsecase) UpdateWebhook(ctx context.Context, req *UIUpdateWebhookRequest) (*model.WebhookConfig, error) {
	exists, err := u.webhookRepo.ExistByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check photonpay UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := u.webhookRepo.FindByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find photonpay UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.TargetURL, item.Enabled = req.TargetURL, req.Enabled
	if err := u.webhookRepo.Save(ctx, item); err != nil {
		zap.S().Errorw("update photonpay UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}
func (u *PhotonPayUIUsecase) DeleteWebhook(ctx context.Context, id model.ID) error {
	exists, err := u.webhookRepo.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay UI webhook", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	if err := u.webhookRepo.Delete(ctx, id); err != nil {
		zap.S().Errorw("delete photonpay UI webhook", "error", err)
		return ErrDatabaseOperation
	}
	return nil
}

type UICreateCardHolderRequest struct {
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (u *PhotonPayUIUsecase) CreateCardHolder(ctx context.Context, req *UICreateCardHolderRequest) (*model.CardHolder, error) {
	dateOfBirth := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	holder := &model.CardHolder{
		Channel:                enums.Channel_PhotonPay,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		Mobile:                 req.Mobile,
		MobilePrefix:           photon.DefaultMobilePrefix,
		DateOfBirth:            &dateOfBirth,
		NationalityCountryCode: photon.DefaultNationalityCountryCode,
		Status:                 enums.CardHolderStatus_Normal,
		ReviewStatus:           enums.CardHolderReviewStatus_Approved,
	}
	if err := u.cardHolderRepo.Create(ctx, holder); err != nil {
		zap.S().Errorw("create photonpay UI card holder", "error", err)

		return nil, ErrDatabaseOperation
	}

	return holder, nil
}

func (u *PhotonPayUIUsecase) ListCardHolders(ctx context.Context, req *ListRequest) ([]*model.CardHolder, error) {
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

func (u *PhotonPayUIUsecase) OpenCard(ctx context.Context, req *UIOpenCardRequest) (*model.Card, error) {
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
			Cvv:                    randomx.Digits(3),
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

func (u *PhotonPayUIUsecase) ListCards(ctx context.Context, req *ListRequest) ([]*model.Card, error) {
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

func (u *PhotonPayUIUsecase) ChangeCardStatus(ctx context.Context, req *UIChangeCardStatusRequest) (*model.Card, error) {
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

func (u *PhotonPayUIUsecase) ListTransactions(ctx context.Context, req *ListRequest) ([]*model.CardTransaction, error) {
	transactions, err := u.cardTransactionRepo.ListTransactions(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay UI card transactions", "error", err)

		return nil, ErrDatabaseOperation
	}

	return transactions, nil
}

func (u *PhotonPayUIUsecase) ListAuthorizations(ctx context.Context, req *ListRequest) ([]*model.Authorization, error) {
	items, err := u.authorizationRepo.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay UI authorizations", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

type UISimulateAuthorizationRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

type UISimulateAuthorizationResult struct {
	Authorization   *model.Authorization
	CardTransaction *model.CardTransaction
}

type UISimulateRefundRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

func (u *PhotonPayUIUsecase) SimulateAuthorization(ctx context.Context, req *UISimulateAuthorizationRequest) (*UISimulateAuthorizationResult, error) {
	var result *UISimulateAuthorizationResult
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.getCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}

		now := time.Now().UTC()
		authorization := &model.Authorization{
			Channel:           enums.Channel_PhotonPay,
			CardID:            card.ID,
			Currency:          req.Currency,
			Amount:            req.Amount,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: randomx.Digits(6),
			Status:            enums.TransactionStatus_AUTHORIZED,
			OccurredAt:        now,
		}
		if err := u.authorizationRepo.Create(txCtx, authorization); err != nil {
			zap.S().Errorw("create photonpay UI authorization", "error", err)

			return ErrDatabaseOperation
		}

		transaction := &model.CardTransaction{
			Channel:           enums.Channel_PhotonPay,
			AuthorizationID:   authorization.ID,
			CardID:            card.ID,
			Status:            enums.TransactionStatus_AUTHORIZED,
			Type:              enums.CardTransactionType_AUTH,
			Currency:          req.Currency,
			TxAmount:          req.Amount,
			TxCurrency:        req.Currency,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: authorization.AuthorizationCode,
			OccurredAt:        now,
		}
		if err := u.cardTransactionRepo.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create photonpay UI authorization transaction", "error", err)

			return ErrDatabaseOperation
		}

		result = &UISimulateAuthorizationResult{
			Authorization:   authorization,
			CardTransaction: transaction,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SimulateRefund creates a posted refund directly for an active card.
func (u *PhotonPayUIUsecase) SimulateRefund(ctx context.Context, req *UISimulateRefundRequest) (*model.CardTransaction, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var transaction *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.getCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}
		transaction = &model.CardTransaction{
			Channel:           enums.Channel_PhotonPay,
			CardID:            card.ID,
			Status:            enums.TransactionStatus_SUCCEED,
			Type:              enums.CardTransactionType_REFUND,
			Currency:          req.Currency,
			TxAmount:          req.Amount,
			TxCurrency:        req.Currency,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: randomx.Digits(6),
			OccurredAt:        time.Now().UTC(),
		}
		if err := u.cardTransactionRepo.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create photonpay UI simulated refund", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

type UIApplyTransactionStepRequest struct {
	CardTransactionID model.ID
	Type              enums.CardTransactionType
	Amount            decimal.Decimal
}

func (u *PhotonPayUIUsecase) ApplyTransactionStep(ctx context.Context, req *UIApplyTransactionStepRequest) (*model.CardTransaction, error) {
	var next *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		origin, err := u.getCardTransaction(txCtx, req.CardTransactionID)
		if err != nil {
			return err
		}

		amount := req.Amount
		if amount.IsZero() {
			amount = origin.TxAmount
		}
		if !amount.IsPositive() || amount.GreaterThan(origin.TxAmount) {
			return ErrInvalidOperation
		}

		status := enums.TransactionStatus_SUCCEED
		if req.Type == enums.CardTransactionType_VOID {
			status = enums.TransactionStatus_VOID
		}

		next = &model.CardTransaction{
			Channel:                 enums.Channel_PhotonPay,
			OriginCardTransactionID: origin.ID,
			AuthorizationID:         origin.AuthorizationID,
			CardID:                  origin.CardID,
			Status:                  status,
			Type:                    req.Type,
			Currency:                origin.Currency,
			TxAmount:                amount,
			TxCurrency:              origin.TxCurrency,
			MerchantName:            origin.MerchantName,
			MerchantCountry:         origin.MerchantCountry,
			MerchantMCC:             origin.MerchantMCC,
			AuthorizationCode:       origin.AuthorizationCode,
			OccurredAt:              time.Now().UTC(),
		}
		if err := u.cardTransactionRepo.Create(txCtx, next); err != nil {
			zap.S().Errorw("create photonpay UI card transaction step", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return next, nil
}

func (u *PhotonPayUIUsecase) getCard(ctx context.Context, id model.ID) (*model.Card, error) {
	exists, err := u.cardRepo.ExistCardByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay UI card", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	card, err := u.cardRepo.FindCardByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find photonpay UI card", "error", err)

		return nil, ErrDatabaseOperation
	}

	return card, nil
}

func (u *PhotonPayUIUsecase) getCardTransaction(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	exists, err := u.cardTransactionRepo.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay UI card transaction", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	transaction, err := u.cardTransactionRepo.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find photonpay UI card transaction", "error", err)

		return nil, ErrDatabaseOperation
	}

	return transaction, nil
}
