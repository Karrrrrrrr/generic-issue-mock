package biz

import (
	"context"
	"encoding/json"
	"strconv"
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
	transaction             PhotonPayTransaction
	cardHolderRepo          CardHolderRepository
	cardRepo                CardRepository
	cardProductRepo         CardProductRepository
	authorizationRepo       AuthorizationRepository
	cardTransactionRepo     CardTransactionRepository
	webhookRepo             WebhookConfigRepository
	authorizationConfigRepo AuthorizationConfigRepository
	webhookRecordRepo       WebhookRecordRepository
	webhookClient           WebhookClient
	accountRepo             AccountRepository
	walletRepo              WalletRepository
	virtualAccountRepo      VirtualAccountRepository
}

func NewPhotonPayUIUsecase(injector *do.Injector) (*PhotonPayUIUsecase, error) {
	return &PhotonPayUIUsecase{
		transaction:             do.MustInvoke[PhotonPayTransaction](injector),
		cardHolderRepo:          do.MustInvoke[CardHolderRepository](injector),
		cardRepo:                do.MustInvoke[CardRepository](injector),
		cardProductRepo:         do.MustInvoke[CardProductRepository](injector),
		authorizationRepo:       do.MustInvoke[AuthorizationRepository](injector),
		cardTransactionRepo:     do.MustInvoke[CardTransactionRepository](injector),
		webhookRepo:             do.MustInvoke[WebhookConfigRepository](injector),
		authorizationConfigRepo: do.MustInvoke[AuthorizationConfigRepository](injector),
		webhookRecordRepo:       do.MustInvoke[WebhookRecordRepository](injector),
		webhookClient:           do.MustInvoke[WebhookClient](injector),
		accountRepo:             do.MustInvoke[AccountRepository](injector),
		walletRepo:              do.MustInvoke[WalletRepository](injector),
		virtualAccountRepo:      do.MustInvoke[VirtualAccountRepository](injector),
	}, nil
}

type UIUpdateAuthorizationConfigRequest struct {
	AccountID        model.ID
	TargetURL        string
	Enabled          bool
	TimeoutMillis    int
	FallbackBehavior string
}

func (u *PhotonPayUIUsecase) GetAuthorizationConfig(ctx context.Context, accountID model.ID) (*model.AuthorizationConfig, error) {
	accountExists, err := u.accountRepo.Exist(ctx, accountID)
	if err != nil {
		zap.S().Errorw("check photonpay authorization config account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !accountExists {
		return nil, ErrResourceNotFound
	}
	exists, err := u.authorizationConfigRepo.ExistByAccountID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("check photonpay authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := u.authorizationConfigRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("find photonpay authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *PhotonPayUIUsecase) UpdateAuthorizationConfig(ctx context.Context, req *UIUpdateAuthorizationConfigRequest) (*model.AuthorizationConfig, error) {
	accountExists, err := u.accountRepo.Exist(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check photonpay authorization config account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !accountExists {
		return nil, ErrResourceNotFound
	}
	exists, err := u.authorizationConfigRepo.ExistByAccountID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check photonpay authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		item := &model.AuthorizationConfig{
			AccountID:        req.AccountID,
			Channel:          enums.Channel_PhotonPay,
			TargetURL:        req.TargetURL,
			Enabled:          req.Enabled,
			TimeoutMillis:    req.TimeoutMillis,
			FallbackBehavior: req.FallbackBehavior,
		}
		if err := u.authorizationConfigRepo.Create(ctx, item); err != nil {
			zap.S().Errorw("create photonpay authorization config", "error", err)
			return nil, ErrDatabaseOperation
		}
		return item, nil
	}
	item, err := u.authorizationConfigRepo.FindByAccountID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("find photonpay authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.TargetURL = req.TargetURL
	item.Enabled = req.Enabled
	item.TimeoutMillis = req.TimeoutMillis
	item.FallbackBehavior = req.FallbackBehavior
	if err := u.authorizationConfigRepo.Save(ctx, item); err != nil {
		zap.S().Errorw("save photonpay authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

type UICreateVirtualAccountRequest struct {
	AccountID model.ID
	Name      string
}

func (u *PhotonPayUIUsecase) CreateVirtualAccount(
	ctx context.Context,
	req *UICreateVirtualAccountRequest,
) (*model.VirtualAccount, error) {
	var item *model.VirtualAccount
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.accountRepo.Exist(txCtx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check photonpay virtual account", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		wallet := &model.Wallet{
			AccountID: req.AccountID,
			Channel:   enums.Channel_PhotonPay,
			Type:      enums.WalletType_VirtualAccount,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepo.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create photonpay virtual account wallet", "error", err)
			return ErrDatabaseOperation
		}

		item = &model.VirtualAccount{
			AccountID: req.AccountID,
			Channel:   enums.Channel_PhotonPay,
			WalletID:  wallet.ID,
			Name:      req.Name,
			Wallet:    wallet,
		}
		if err := u.virtualAccountRepo.Create(txCtx, item); err != nil {
			zap.S().Errorw("create photonpay virtual account", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (u *PhotonPayUIUsecase) ListVirtualAccounts(ctx context.Context) ([]*model.VirtualAccount, error) {
	items, err := u.virtualAccountRepo.List(ctx)
	if err != nil {
		zap.S().Errorw("list photonpay virtual accounts", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

type UICreateAccountRequest struct{ Name string }

func (u *PhotonPayUIUsecase) CreateAccount(ctx context.Context, req *UICreateAccountRequest) (*model.Account, error) {
	var item *model.Account
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		item = &model.Account{Channel: enums.Channel_PhotonPay, Name: req.Name}
		if err := u.accountRepo.Create(txCtx, item); err != nil {
			zap.S().Errorw("create photonpay UI account", "error", err)
			return ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: item.ID,
			Channel:   enums.Channel_PhotonPay,
			Type:      enums.WalletType_Account,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepo.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create photonpay UI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		item.WalletID = wallet.ID
		if err := u.accountRepo.Save(txCtx, item); err != nil {
			zap.S().Errorw("attach photonpay UI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		virtualWallet := &model.Wallet{
			AccountID: item.ID,
			Channel:   enums.Channel_PhotonPay,
			Type:      enums.WalletType_VirtualAccount,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepo.Create(txCtx, virtualWallet); err != nil {
			zap.S().Errorw("create photonpay account virtual wallet", "error", err)
			return ErrDatabaseOperation
		}
		virtualAccount := &model.VirtualAccount{
			AccountID: item.ID,
			Channel:   enums.Channel_PhotonPay,
			WalletID:  virtualWallet.ID,
			Name:      item.Name,
		}
		if err := u.virtualAccountRepo.Create(txCtx, virtualAccount); err != nil {
			zap.S().Errorw("create photonpay account virtual account", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (u *PhotonPayUIUsecase) ListAccounts(ctx context.Context, req *ListRequest) ([]*model.Account, int64, error) {
	items, err := u.accountRepo.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay UI accounts", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.accountRepo.Count(ctx)
	if err != nil {
		zap.S().Errorw("count photonpay UI accounts", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	return items, total, nil
}

type UIUpdateAccountRequest struct {
	ID   model.ID
	Name string
}

func (u *PhotonPayUIUsecase) UpdateAccount(ctx context.Context, req *UIUpdateAccountRequest) (*model.Account, error) {
	exists, err := u.accountRepo.Exist(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check photonpay UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := u.accountRepo.Find(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find photonpay UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.Name = req.Name
	if err := u.accountRepo.Save(ctx, item); err != nil {
		zap.S().Errorw("rename photonpay UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *PhotonPayUIUsecase) ListWebhookRecords(
	ctx context.Context,
	req *ListRequest,
) ([]*model.WebhookRecord, int64, error) {
	items, err := u.webhookRecordRepo.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay webhook records", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	total, err := u.webhookRecordRepo.Count(ctx)
	if err != nil {
		zap.S().Errorw("count photonpay webhook records", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *PhotonPayUIUsecase) ReplayWebhookRecord(
	ctx context.Context,
	id model.ID,
) (*model.WebhookRecord, error) {
	exists, err := u.webhookRecordRepo.Exist(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay webhook record", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	original, err := u.webhookRecordRepo.Find(ctx, id)
	if err != nil {
		zap.S().Errorw("find photonpay webhook record", "error", err)
		return nil, ErrDatabaseOperation
	}
	replay := &model.WebhookRecord{
		WebhookConfigID: original.WebhookConfigID,
		AccountID:       original.AccountID,
		Channel:         enums.Channel_PhotonPay,
		Event:           original.Event,
		TargetURL:       original.TargetURL,
		SourceID:        original.SourceID,
		Payload:         original.Payload,
		RequestHeaders:  original.RequestHeaders,
		Status:          enums.WebhookDeliveryStatus_Pending,
		AttemptCount:    original.AttemptCount + 1,
	}
	if err := u.webhookRecordRepo.Create(ctx, replay); err != nil {
		zap.S().Errorw("create photonpay webhook replay record", "error", err)
		return nil, ErrDatabaseOperation
	}

	result, deliveryErr := u.webhookClient.Deliver(ctx, &PhotonPayWebhookDeliveryRequest{
		TargetURL:      replay.TargetURL,
		Payload:        replay.Payload,
		RequestHeaders: replay.RequestHeaders,
	})
	if deliveryErr != nil {
		replay.Status = enums.WebhookDeliveryStatus_Failed
		replay.ErrorMessage = deliveryErr.Error()
	} else {
		replay.StatusCode = result.StatusCode
		replay.ResponseBody = result.ResponseBody
		replay.RequestHeaders = result.RequestHeaders
		replay.ResponseHeaders = result.ResponseHeaders
		if result.StatusCode >= 200 && result.StatusCode < 300 && photonPayWebhookAcknowledged(result.ResponseBody) {
			deliveredAt := time.Now().UTC()
			replay.Status = enums.WebhookDeliveryStatus_Succeeded
			replay.DeliveredAt = &deliveredAt
		} else {
			replay.Status = enums.WebhookDeliveryStatus_Failed
			replay.ErrorMessage = "unexpected PhotonPay webhook response"
		}
	}
	if err := u.webhookRecordRepo.Save(ctx, replay); err != nil {
		zap.S().Errorw("save photonpay webhook replay record", "error", err)
		return nil, ErrDatabaseOperation
	}

	return replay, nil
}

type UICreateWebhookRequest struct {
	AccountID model.ID
	Event     photon.WebhookEvent
	TargetURL string
	Enabled   bool
}
type UIUpdateWebhookRequest struct {
	ID        model.ID
	TargetURL string
	Enabled   bool
}

func (u *PhotonPayUIUsecase) CreateWebhook(ctx context.Context, req *UICreateWebhookRequest) (*model.WebhookConfig, error) {
	if !req.Event.Valid() {
		return nil, ErrInvalidOperation
	}
	exists, err := u.accountRepo.Exist(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check photonpay webhook account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item := &model.WebhookConfig{
		AccountID: req.AccountID,
		Channel:   enums.Channel_PhotonPay,
		Event:     string(req.Event),
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	}
	if err := u.webhookRepo.Create(ctx, item); err != nil {
		zap.S().Errorw("create photonpay UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}
func (u *PhotonPayUIUsecase) ListWebhooks(
	ctx context.Context,
	req *WebhookListRequest,
) ([]*model.WebhookConfig, error) {
	items, err := u.webhookRepo.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list photonpay UI webhooks", "error", err)
		return nil, ErrDatabaseOperation
	}
	result := make([]*model.WebhookConfig, 0, len(items))
	for _, item := range items {
		if photon.WebhookEvent(item.Event).Valid() {
			result = append(result, item)
		}
	}
	return result, nil
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
	AccountID model.ID
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (u *PhotonPayUIUsecase) CreateCardHolder(ctx context.Context, req *UICreateCardHolderRequest) (*model.CardHolder, error) {
	dateOfBirth := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	holder := &model.CardHolder{
		AccountID:              req.AccountID,
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
	AccountID    model.ID
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
		virtualAccount, err := u.virtualAccountRepo.FindByAccountID(txCtx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay UI account virtual account", "error", err)
			return ErrDatabaseOperation
		}

		card = &model.Card{
			AccountID:              req.AccountID,
			Channel:                enums.Channel_PhotonPay,
			CardProductID:          product.ID,
			CardBin:                product.Prefix,
			CardNumber:             cardNumber,
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

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	u.dispatchCardStatus(ctx, card)

	return card, nil
}

type UIFundCardRequest struct {
	CardID model.ID
	Amount decimal.Decimal
}

func (u *PhotonPayUIUsecase) FundCard(ctx context.Context, req *UIFundCardRequest) (*model.Card, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}

	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		var err error
		card, err = u.getCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.WalletID == 0 {
			return ErrResourceNotFound
		}

		account, err := u.accountRepo.Find(txCtx, card.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay UI card account", "error", err)

			return ErrDatabaseOperation
		}
		accountID := card.AccountID
		source, err := u.walletRepo.FindByIDForUpdate(txCtx, &ResourceRequest{
			AccountID: &accountID,
			ID:        account.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay UI account wallet", "error", err)

			return ErrDatabaseOperation
		}
		target, err := u.walletRepo.FindByIDForUpdate(txCtx, &ResourceRequest{
			AccountID: &accountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay UI card wallet", "error", err)

			return ErrDatabaseOperation
		}
		if source.Amount.LessThan(req.Amount) {
			return ErrInvalidOperation
		}
		source.Amount = source.Amount.Sub(req.Amount)
		source.Out = source.Out.Add(req.Amount)
		target.Amount = target.Amount.Add(req.Amount)
		target.In = target.In.Add(req.Amount)
		if err := u.walletRepo.Save(txCtx, source); err != nil {
			zap.S().Errorw("save photonpay UI account wallet", "error", err)

			return ErrDatabaseOperation
		}
		if err := u.walletRepo.Save(txCtx, target); err != nil {
			zap.S().Errorw("save photonpay UI card wallet", "error", err)

			return ErrDatabaseOperation
		}
		card.Wallet = target

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
		if card.WalletID == 0 {
			return ErrResourceNotFound
		}
		accountID := card.AccountID
		wallet, err := u.walletRepo.FindByIDForUpdate(txCtx, &ResourceRequest{
			AccountID: &accountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay UI authorization wallet", "error", err)

			return ErrDatabaseOperation
		}
		if wallet.Amount.LessThan(req.Amount) {
			return ErrInvalidOperation
		}

		authorization := &model.Authorization{
			AccountID:         card.AccountID,
			Channel:           enums.Channel_PhotonPay,
			CardID:            card.ID,
			Currency:          req.Currency,
			Amount:            req.Amount,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: randomx.Digits(6),
			Status:            enums.TransactionStatus_AUTHORIZED,
		}
		if err := u.authorizationRepo.Create(txCtx, authorization); err != nil {
			zap.S().Errorw("create photonpay UI authorization", "error", err)

			return ErrDatabaseOperation
		}

		transaction := &model.CardTransaction{
			AccountID:         card.AccountID,
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
	u.dispatchTransaction(ctx, result.CardTransaction)

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
			AccountID:         card.AccountID,
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
	u.dispatchTransaction(ctx, transaction)
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
			AccountID:               origin.AccountID,
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
	u.dispatchTransaction(ctx, next)

	return next, nil
}

func (u *PhotonPayUIUsecase) dispatchTransaction(ctx context.Context, transaction *model.CardTransaction) {
	u.dispatch(ctx, photon.WebhookEventFromGenericTransactionType(transaction.Type), transaction.ID, transaction)
}

func (u *PhotonPayUIUsecase) dispatchCardStatus(ctx context.Context, card *model.Card) {
	payload, err := json.Marshal(struct {
		CardID     string            `json:"cardId"`
		CardStatus photon.CardStatus `json:"cardStatus"`
	}{
		CardID:     strconv.FormatInt(card.ID, 10),
		CardStatus: photon.CardStatusFromGeneric(card.Status),
	})
	if err != nil {
		zap.S().Errorw("marshal photonpay card status webhook payload", "error", err)
		return
	}
	configs, err := u.webhookRepo.List(ctx, &WebhookListRequest{
		AccountID: card.AccountID,
	})
	if err != nil {
		zap.S().Errorw("list photonpay card status webhook configs", "error", err)
		return
	}
	for _, config := range configs {
		if !config.Enabled || config.Event != string(photon.WebhookEventCardStatusUpdate) {
			continue
		}
		record := &model.WebhookRecord{
			WebhookConfigID: config.ID,
			AccountID:       config.AccountID,
			Channel:         enums.Channel_PhotonPay,
			Event:           string(photon.WebhookEventCardStatusUpdate),
			TargetURL:       config.TargetURL,
			SourceID:        strconv.FormatInt(card.ID, 10),
			Payload:         payload,
			Status:          enums.WebhookDeliveryStatus_Pending,
			AttemptCount:    1,
		}
		if err := u.webhookRecordRepo.Create(ctx, record); err != nil {
			zap.S().Errorw("create photonpay card status webhook record", "error", err)
			continue
		}
		result, deliveryErr := u.webhookClient.Deliver(ctx, &PhotonPayWebhookDeliveryRequest{
			TargetURL:      config.TargetURL,
			Payload:        payload,
			NotifyCategory: string(photon.WebhookNotificationCategoryIssuingCard),
			NotifyType:     string(photon.WebhookEventCardStatusUpdate),
			PublishedAt:    time.Now().UTC().Format(time.RFC3339),
		})
		if deliveryErr != nil {
			record.Status = enums.WebhookDeliveryStatus_Failed
			record.ErrorMessage = deliveryErr.Error()
			zap.S().Errorw("deliver photonpay card status webhook", "error", deliveryErr, "webhook_record_id", record.ID)
		} else {
			record.StatusCode = result.StatusCode
			record.ResponseBody = result.ResponseBody
			record.RequestHeaders = result.RequestHeaders
			record.ResponseHeaders = result.ResponseHeaders
			if result.StatusCode >= 200 && result.StatusCode < 300 && photonPayWebhookAcknowledged(result.ResponseBody) {
				deliveredAt := time.Now().UTC()
				record.Status = enums.WebhookDeliveryStatus_Succeeded
				record.DeliveredAt = &deliveredAt
			} else {
				record.Status = enums.WebhookDeliveryStatus_Failed
				record.ErrorMessage = "unexpected PhotonPay webhook response"
			}
		}
		if err := u.webhookRecordRepo.Save(ctx, record); err != nil {
			zap.S().Errorw("save photonpay card status webhook record", "error", err)
		}
	}
}

func (u *PhotonPayUIUsecase) dispatch(ctx context.Context, event photon.WebhookEvent, sourceID model.ID, transaction *model.CardTransaction) {
	payload, err := u.photonPayWebhookPayload(ctx, transaction)
	if err != nil {
		zap.S().Errorw("marshal photonpay webhook payload", "error", err)
		return
	}
	configs, err := u.webhookRepo.List(ctx, &WebhookListRequest{
		AccountID: transaction.AccountID,
	})
	if err != nil {
		zap.S().Errorw("list photonpay webhook configs", "error", err)
		return
	}
	for _, config := range configs {
		if !config.Enabled || config.Event != string(event) {
			continue
		}
		record := &model.WebhookRecord{
			WebhookConfigID: config.ID,
			AccountID:       config.AccountID,
			Channel:         enums.Channel_PhotonPay,
			Event:           string(event),
			TargetURL:       config.TargetURL,
			SourceID:        strconv.FormatInt(int64(sourceID), 10),
			Payload:         payload,
			Status:          enums.WebhookDeliveryStatus_Pending,
			AttemptCount:    1,
		}
		if err := u.webhookRecordRepo.Create(ctx, record); err != nil {
			zap.S().Errorw("create photonpay webhook record", "error", err)
			continue
		}
		result, deliveryErr := u.webhookClient.Deliver(ctx, &PhotonPayWebhookDeliveryRequest{
			TargetURL:      config.TargetURL,
			Payload:        payload,
			NotifyCategory: string(photonPayWebhookCategory(transaction)),
			NotifyType:     string(event),
			PublishedAt:    transaction.CreatedAt.UTC().Format(time.RFC3339),
		})
		if deliveryErr != nil {
			record.Status = enums.WebhookDeliveryStatus_Failed
			record.ErrorMessage = deliveryErr.Error()
			zap.S().Errorw("deliver photonpay webhook", "error", deliveryErr, "webhook_record_id", record.ID)
		} else {
			record.StatusCode = result.StatusCode
			record.ResponseBody = result.ResponseBody
			record.RequestHeaders = result.RequestHeaders
			record.ResponseHeaders = result.ResponseHeaders
			if result.StatusCode >= 200 && result.StatusCode < 300 && photonPayWebhookAcknowledged(result.ResponseBody) {
				deliveredAt := time.Now().UTC()
				record.Status = enums.WebhookDeliveryStatus_Succeeded
				record.DeliveredAt = &deliveredAt
			} else {
				record.Status = enums.WebhookDeliveryStatus_Failed
				record.ErrorMessage = "unexpected PhotonPay webhook response"
			}
		}
		if err := u.webhookRecordRepo.Save(ctx, record); err != nil {
			zap.S().Errorw("save photonpay webhook record", "error", err)
		}
	}
}

func photonPayWebhookAcknowledged(body string) bool {
	var response struct {
		Roger bool `json:"roger"`
	}
	return json.Unmarshal([]byte(body), &response) == nil && response.Roger
}

type photonPayEventPayload struct {
	MemberID                   string `json:"memberId"`
	MatrixAccount              string `json:"matrixAccount"`
	CreatedAt                  string `json:"createdAt"`
	UpdatedAt                  string `json:"updatedAt"`
	CardID                     string `json:"cardId"`
	CardType                   string `json:"cardType"`
	TransactionID              string `json:"transactionId"`
	OriginTransactionID        string `json:"originTransactionId,omitempty"`
	RequestID                  string `json:"requestId,omitempty"`
	TransactionType            string `json:"transactionType"`
	Status                     string `json:"status"`
	Code                       string `json:"code"`
	Message                    string `json:"msg"`
	MCC                        string `json:"mcc,omitempty"`
	AuthCode                   string `json:"authCode,omitempty"`
	TransactionAmount          string `json:"transactionAmount"`
	TransactionCurrency        string `json:"transactionCurrency"`
	TxnPrincipalChangeAccount  string `json:"txnPrincipalChangeAccount"`
	TxnPrincipalChangeAmount   string `json:"txnPrincipalChangeAmount"`
	TxnPrincipalChangeCurrency string `json:"txnPrincipalChangeCurrency"`
	MerchantName               string `json:"merchantName,omitempty"`
	MerchantLocation           string `json:"merchantLocation,omitempty"`
	TransactionStatus          string `json:"transactionStatus"`
	TransactionCountry         string `json:"transactionCountry,omitempty"`
	TransactionHappenedAt      string `json:"transactionHappenedAt"`
	SettleAmount               string `json:"settleAmount,omitempty"`
	SettleCurrency             string `json:"settleCurrency,omitempty"`
}

func (u *PhotonPayUIUsecase) photonPayWebhookPayload(ctx context.Context, transaction *model.CardTransaction) ([]byte, error) {
	card, err := u.cardRepo.FindCardByID(ctx, transaction.CardID)
	if err != nil {
		zap.S().Errorw("find photonpay webhook card", "error", err)
		return nil, ErrDatabaseOperation
	}
	payload := photonPayEventPayload{
		MemberID:                   photon.MemberID,
		CreatedAt:                  transaction.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                  transaction.UpdatedAt.UTC().Format(time.RFC3339),
		CardID:                     strconv.FormatInt(card.ID, 10),
		CardType:                   string(photon.CardTypeFromGeneric(card.CardType)),
		TransactionID:              strconv.FormatInt(transaction.ID, 10),
		RequestID:                  transaction.RequestID,
		TransactionType:            string(photon.WebhookEventFromGenericTransactionType(transaction.Type)),
		Status:                     string(photon.TransactionStatusFromGeneric(transaction.Status)),
		Code:                       photon.WebhookSuccessCode,
		Message:                    photon.WebhookSuccessMessage,
		MCC:                        transaction.MerchantMCC,
		AuthCode:                   transaction.AuthorizationCode,
		TransactionAmount:          transaction.TxAmount.String(),
		TransactionCurrency:        string(transaction.TxCurrency),
		TxnPrincipalChangeAccount:  photon.WebhookBalanceAccountCard,
		TxnPrincipalChangeAmount:   transaction.TxAmount.String(),
		TxnPrincipalChangeCurrency: string(transaction.TxCurrency),
		MerchantName:               transaction.MerchantName,
		MerchantLocation:           transaction.MerchantCountry,
		TransactionStatus:          string(photon.TransactionStatusFromGeneric(transaction.Status)),
		TransactionCountry:         transaction.MerchantCountry,
		TransactionHappenedAt:      transaction.CreatedAt.UTC().Format(time.RFC3339),
	}
	if transaction.OriginCardTransactionID != 0 {
		payload.OriginTransactionID = strconv.FormatInt(transaction.OriginCardTransactionID, 10)
	}
	if transaction.Type == enums.CardTransactionType_CLEAR {
		payload.SettleAmount = transaction.TxAmount.String()
		payload.SettleCurrency = string(transaction.TxCurrency)
	}
	return json.Marshal(payload)
}

func photonPayWebhookCategory(transaction *model.CardTransaction) photon.WebhookNotificationCategory {
	if transaction.Status == enums.TransactionStatus_SUCCEED {
		return photon.WebhookNotificationCategoryIssuingSettlement
	}
	return photon.WebhookNotificationCategoryIssuing
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
