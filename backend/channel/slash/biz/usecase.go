package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type SlashTransaction interface {
	InTx(context.Context, func(context.Context) error) error
}

type SlashAccountRepository interface {
	Create(context.Context, *model.Account) error
	Save(context.Context, *model.Account) error
	Exist(context.Context, model.ID) (bool, error)
	Find(context.Context, model.ID) (*model.Account, error)
	Count(context.Context) (int64, error)
	List(context.Context, *ListAccountsRequest) ([]*model.Account, error)
}

type SlashCardHolderRepository interface {
	Create(context.Context, *model.CardHolder) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardHolder, error)
	Count(context.Context, *ListCardHoldersRequest) (int64, error)
	List(context.Context, *ListCardHoldersRequest) ([]*model.CardHolder, error)
	Save(context.Context, *model.CardHolder) error
}

type SlashCardRepository interface {
	Create(context.Context, *model.Card) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.Card, error)
	Count(context.Context, *ListCardsRequest) (int64, error)
	List(context.Context, *ListCardsRequest) ([]*model.Card, error)
	Save(context.Context, *model.Card) error
	ExistByAccountID(context.Context, *ResourceRequest) (bool, error)
	FindByAccountID(context.Context, *ResourceRequest) (*model.Card, error)
}

type SlashCardProductRepository interface {
	ExistByID(context.Context, model.ID) (bool, error)
	FindByIDForUpdate(context.Context, model.ID) (*model.CardProduct, error)
	ExistDefault(context.Context) (bool, error)
	FindDefaultForUpdate(context.Context) (*model.CardProduct, error)
	List(context.Context) ([]*model.CardProduct, error)
	Save(context.Context, *model.CardProduct) error
	ExistByAccountID(context.Context, *ResourceRequest) (bool, error)
	FindByAccountIDForUpdate(context.Context, *ResourceRequest) (*model.CardProduct, error)
	ListByAccountID(context.Context, model.ID) ([]*model.CardProduct, error)
}

type SlashAuthorizationRepository interface {
	Create(context.Context, *model.Authorization) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.Authorization, error)
	Count(context.Context, *ListAuthorizationsRequest) (int64, error)
	List(context.Context, *ListAuthorizationsRequest) ([]*model.Authorization, error)
	Save(context.Context, *model.Authorization) error
}

type SlashCardTransactionRepository interface {
	Create(context.Context, *model.CardTransaction) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardTransaction, error)
	Count(context.Context, *ListCardTransactionsRequest) (int64, error)
	List(context.Context, *ListCardTransactionsRequest) ([]*model.CardTransaction, error)
	ExistByAccountID(context.Context, *ResourceRequest) (bool, error)
	FindByAccountID(context.Context, *ResourceRequest) (*model.CardTransaction, error)
}

type SlashVirtualAccountRepository interface {
	List(context.Context) ([]*model.VirtualAccount, error)
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.VirtualAccount, error)
	ListByAccountID(context.Context, model.ID) ([]*model.VirtualAccount, error)
	ExistByAccountID(context.Context, *ResourceRequest) (bool, error)
	FindByAccountID(context.Context, *ResourceRequest) (*model.VirtualAccount, error)
}

type SlashWalletRepository interface {
	Create(context.Context, *model.Wallet) error
	FindByIDForUpdate(context.Context, model.ID) (*model.Wallet, error)
	Save(context.Context, *model.Wallet) error
	FindByAccountIDForUpdate(context.Context, *ResourceRequest) (*model.Wallet, error)
}

type SlashWebhookConfigRepository interface {
	Create(context.Context, *model.WebhookConfig) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.WebhookConfig, error)
	List(context.Context) ([]*model.WebhookConfig, error)
	Save(context.Context, *model.WebhookConfig) error
	Delete(context.Context, model.ID) error
}

type SlashUIUsecase struct {
	transaction               SlashTransaction
	cardHolderRepository      SlashCardHolderRepository
	cardRepository            SlashCardRepository
	cardProductRepository     SlashCardProductRepository
	authorizationRepository   SlashAuthorizationRepository
	cardTransactionRepository SlashCardTransactionRepository
	webhookConfigRepository   SlashWebhookConfigRepository
	virtualAccountRepository  SlashVirtualAccountRepository
	accountRepository         SlashAccountRepository
	walletRepository          SlashWalletRepository
}

func NewSlashUIUsecase(injector *do.Injector) (*SlashUIUsecase, error) {
	return &SlashUIUsecase{
		transaction:               do.MustInvoke[SlashTransaction](injector),
		cardHolderRepository:      do.MustInvoke[SlashCardHolderRepository](injector),
		cardRepository:            do.MustInvoke[SlashCardRepository](injector),
		cardProductRepository:     do.MustInvoke[SlashCardProductRepository](injector),
		authorizationRepository:   do.MustInvoke[SlashAuthorizationRepository](injector),
		cardTransactionRepository: do.MustInvoke[SlashCardTransactionRepository](injector),
		webhookConfigRepository:   do.MustInvoke[SlashWebhookConfigRepository](injector),
		virtualAccountRepository:  do.MustInvoke[SlashVirtualAccountRepository](injector),
		accountRepository:         do.MustInvoke[SlashAccountRepository](injector),
		walletRepository:          do.MustInvoke[SlashWalletRepository](injector),
	}, nil
}

type CreateAccountRequest struct {
	Name string
}

func (u *SlashUIUsecase) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*model.Account, error) {
	var item *model.Account
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		item = &model.Account{Channel: enums.Channel_Slash, Name: req.Name}
		if err := u.accountRepository.Create(txCtx, item); err != nil {
			zap.S().Errorw("create slash UI account", "error", err)
			return ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: item.ID,
			Channel:   enums.Channel_Slash,
			Type:      enums.WalletType_Account,
			Currency:  enums.Currency_USD,
		}
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create slash UI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		item.WalletID = wallet.ID
		if err := u.accountRepository.Save(txCtx, item); err != nil {
			zap.S().Errorw("attach slash UI account wallet", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

type ListAccountsRequest struct {
	Offset int
	Limit  int
}

func (u *SlashUIUsecase) ListAccounts(
	ctx context.Context,
	req *ListAccountsRequest,
) ([]*model.Account, int64, error) {
	items, err := u.accountRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list slash UI accounts", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	total, err := u.accountRepository.Count(ctx)
	if err != nil {
		zap.S().Errorw("count slash UI accounts", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

type UpdateAccountRequest struct {
	ID   model.ID
	Name string
}

func (u *SlashUIUsecase) UpdateAccount(
	ctx context.Context,
	req *UpdateAccountRequest,
) (*model.Account, error) {
	exists, err := u.accountRepository.Exist(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check slash UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item, err := u.accountRepository.Find(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find slash UI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.Name = req.Name
	if err := u.accountRepository.Save(ctx, item); err != nil {
		zap.S().Errorw("update slash UI account", "error", err)
		return nil, ErrDatabaseOperation
	}

	return item, nil
}

type ListCardHoldersRequest struct {
	AccountID model.ID
	Offset    int
	Limit     int
}

type ListCardsRequest struct {
	AccountID  model.ID
	Offset     int
	Limit      int
	IDContains string
	CardNumber string
	Status     enums.CardStatus
}

type ListAuthorizationsRequest struct {
	AccountID model.ID
	Offset    int
	Limit     int
	ID        model.ID
	CardID    model.ID
	Status    enums.CardTransactionStatus
}

type ListCardTransactionsRequest struct {
	AccountID       model.ID
	Offset          int
	Limit           int
	ID              model.ID
	CardID          model.ID
	AuthorizationID model.ID
	Type            enums.CardTransactionType
	Status          enums.CardTransactionStatus
}

type ResourceRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CreateCardHolderRequest struct {
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (u *SlashUIUsecase) CreateCardHolder(ctx context.Context, req *CreateCardHolderRequest) (*model.CardHolder, error) {
	holder := &model.CardHolder{
		Channel:      enums.Channel_Slash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Mobile:       req.Mobile,
		Status:       enums.CardHolderStatus_Normal,
		ReviewStatus: enums.CardHolderReviewStatus_Approved,
		Shared:       true,
	}
	if err := u.cardHolderRepository.Create(ctx, holder); err != nil {
		zap.S().Errorw("create slash card holder", "error", err)
		return nil, ErrDatabaseOperation
	}

	return holder, nil
}

func (u *SlashUIUsecase) ListCardHolders(ctx context.Context, req *ListCardHoldersRequest) ([]*model.CardHolder, int64, error) {
	items, err := u.cardHolderRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list slash card holders", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.cardHolderRepository.Count(ctx, req)
	if err != nil {
		zap.S().Errorw("count slash card holders", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

type CreateWebhookRequest struct {
	Event     string
	TargetURL string
	Enabled   bool
}

func (u *SlashUIUsecase) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*model.WebhookConfig, error) {
	item := &model.WebhookConfig{
		Channel:   enums.Channel_Slash,
		Event:     req.Event,
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	}
	if err := u.webhookConfigRepository.Create(ctx, item); err != nil {
		zap.S().Errorw("create slash webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *SlashUIUsecase) ListWebhooks(ctx context.Context) ([]*model.WebhookConfig, error) {
	items, err := u.webhookConfigRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list slash webhooks", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

type UpdateWebhookRequest struct {
	ID        model.ID
	TargetURL string
	Enabled   bool
}

func (u *SlashUIUsecase) UpdateWebhook(ctx context.Context, req *UpdateWebhookRequest) (*model.WebhookConfig, error) {
	exists, err := u.webhookConfigRepository.ExistByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check slash webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := u.webhookConfigRepository.FindByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find slash webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.TargetURL = req.TargetURL
	item.Enabled = req.Enabled
	if err := u.webhookConfigRepository.Save(ctx, item); err != nil {
		zap.S().Errorw("update slash webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *SlashUIUsecase) DeleteWebhook(ctx context.Context, id model.ID) error {
	exists, err := u.webhookConfigRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash webhook", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	if err := u.webhookConfigRepository.Delete(ctx, id); err != nil {
		zap.S().Errorw("delete slash webhook", "error", err)
		return ErrDatabaseOperation
	}
	return nil
}

type CreateCardRequest struct {
	CardHolderID  model.ID
	CardProductID model.ID
	Currency      enums.Currency
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

		product, err := u.getCardProductForUpdate(txCtx, req.CardProductID)
		if err != nil {
			return err
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
	items, err := u.cardRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list slash cards", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.cardRepository.Count(ctx, req)
	if err != nil {
		zap.S().Errorw("count slash cards", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

type CardProductInfo struct {
	Product *model.CardProduct
}

func (u *SlashUIUsecase) ListCardProducts(ctx context.Context) ([]*CardProductInfo, error) {
	products, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list slash card products", "error", err)
		return nil, ErrDatabaseOperation
	}
	items := make([]*CardProductInfo, 0, len(products))
	for _, product := range products {
		items = append(items, &CardProductInfo{
			Product: product,
		})
	}

	return items, nil
}

func (u *SlashUIUsecase) ListVirtualAccounts(ctx context.Context) ([]*model.VirtualAccount, error) {
	items, err := u.virtualAccountRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list slash virtual accounts", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
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

type UpdateCardStatusRequest struct {
	ID     model.ID
	Status enums.CardStatus
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

type SimulateAuthorizationRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

type SimulateAuthorizationResult struct {
	Authorization   *model.Authorization
	CardTransaction *model.CardTransaction
}

type SimulateRefundRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

func (u *SlashUIUsecase) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*SimulateAuthorizationResult, error) {
	var result *SimulateAuthorizationResult
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.GetCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}
		walletID := card.WalletID
		if card.VirtualAccountID != nil {
			virtualAccount, err := u.virtualAccountRepository.FindByID(txCtx, *card.VirtualAccountID)
			if err != nil {
				zap.S().Errorw("find slash UI authorization virtual account", "error", err)
				return ErrDatabaseOperation
			}
			walletID = virtualAccount.WalletID
		}
		if walletID == 0 {
			return ErrResourceNotFound
		}
		wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, walletID)
		if err != nil {
			zap.S().Errorw("lock slash UI authorization wallet", "error", err)
			return ErrDatabaseOperation
		}
		if wallet.Amount.LessThan(req.Amount) {
			return ErrInvalidOperation
		}
		now := time.Now().UTC()
		authorization := &model.Authorization{
			Channel:           enums.Channel_Slash,
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
		if err := u.authorizationRepository.Create(txCtx, authorization); err != nil {
			zap.S().Errorw("create slash authorization", "error", err)
			return ErrDatabaseOperation
		}
		transaction := &model.CardTransaction{
			Channel:           enums.Channel_Slash,
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
		if err := u.cardTransactionRepository.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create slash authorization transaction", "error", err)
			return ErrDatabaseOperation
		}
		result = &SimulateAuthorizationResult{
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

// SimulateRefund creates a posted refund directly for an active card. It is not
// a follow-up operation on a prior card transaction.
func (u *SlashUIUsecase) SimulateRefund(ctx context.Context, req *SimulateRefundRequest) (*model.CardTransaction, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var transaction *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.GetCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}

		transaction = &model.CardTransaction{
			Channel:           enums.Channel_Slash,
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
		if err := u.cardTransactionRepository.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create slash simulated refund", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (u *SlashUIUsecase) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) ([]*model.Authorization, int64, error) {
	items, err := u.authorizationRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list slash authorizations", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.authorizationRepository.Count(ctx, req)
	if err != nil {
		zap.S().Errorw("count slash authorizations", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *SlashUIUsecase) GetAuthorization(ctx context.Context, id model.ID) (*model.Authorization, error) {
	if err := u.requireAuthorization(ctx, id); err != nil {
		return nil, err
	}
	item, err := u.authorizationRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash authorization", "error", err)
		return nil, ErrDatabaseOperation
	}

	return item, nil
}

func (u *SlashUIUsecase) ListCardTransactions(ctx context.Context, req *ListCardTransactionsRequest) ([]*model.CardTransaction, int64, error) {
	items, err := u.cardTransactionRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list slash card transactions", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.cardTransactionRepository.Count(ctx, req)
	if err != nil {
		zap.S().Errorw("count slash card transactions", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *SlashUIUsecase) GetCardTransaction(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	if err := u.requireCardTransaction(ctx, id); err != nil {
		return nil, err
	}
	item, err := u.cardTransactionRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash card transaction", "error", err)
		return nil, ErrDatabaseOperation
	}

	return item, nil
}

type ApplyTransactionStepRequest struct {
	CardTransactionID model.ID
	Type              enums.CardTransactionType
	Amount            decimal.Decimal
}

func (u *SlashUIUsecase) ApplyTransactionStep(ctx context.Context, req *ApplyTransactionStepRequest) (*model.CardTransaction, error) {
	var next *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		origin, err := u.GetCardTransaction(txCtx, req.CardTransactionID)
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
			Channel:                 enums.Channel_Slash,
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
		if err := u.cardTransactionRepository.Create(txCtx, next); err != nil {
			zap.S().Errorw("create slash card transaction step", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return next, nil
}

func (u *SlashUIUsecase) requireCardHolder(ctx context.Context, id model.ID) error {
	exists, err := u.cardHolderRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card holder", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	return nil
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

func (u *SlashUIUsecase) requireAuthorization(ctx context.Context, id model.ID) error {
	exists, err := u.authorizationRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash authorization", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	return nil
}

func (u *SlashUIUsecase) requireCardTransaction(ctx context.Context, id model.ID) error {
	exists, err := u.cardTransactionRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card transaction", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	return nil
}

func (u *SlashUIUsecase) getCardProductForUpdate(ctx context.Context, id model.ID) (*model.CardProduct, error) {
	if id == 0 {
		exists, err := u.cardProductRepository.ExistDefault(ctx)
		if err != nil {
			zap.S().Errorw("check slash default card product", "error", err)
			return nil, ErrDatabaseOperation
		}
		if !exists {
			return nil, ErrResourceNotFound
		}

		product, err := u.cardProductRepository.FindDefaultForUpdate(ctx)
		if err != nil {
			zap.S().Errorw("lock slash default card product", "error", err)
			return nil, ErrDatabaseOperation
		}

		return product, nil
	}

	if err := u.requireCardProduct(ctx, id); err != nil {
		return nil, err
	}
	product, err := u.cardProductRepository.FindByIDForUpdate(ctx, id)
	if err != nil {
		zap.S().Errorw("lock slash card product", "error", err)
		return nil, ErrDatabaseOperation
	}

	return product, nil
}

func (u *SlashUIUsecase) requireCardProduct(ctx context.Context, id model.ID) error {
	exists, err := u.cardProductRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card product", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}
