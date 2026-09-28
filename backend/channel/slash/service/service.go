package service

import (
	"context"
	"time"

	"generic-mock/channel/slash/biz"
	slash "generic-mock/channel/slash/enums"
	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/channel/slash/pkg/idconv"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do/v2"
	"github.com/shopspring/decimal"
)

type SlashUIService struct {
	usecase        *biz.SlashUIUsecase
	webhookUsecase *biz.SlashWebhookUsecase
}

func NewSlashUIService(injector do.Injector) (*SlashUIService, error) {
	return &SlashUIService{
		usecase:        do.MustInvoke[*biz.SlashUIUsecase](injector),
		webhookUsecase: do.MustInvoke[*biz.SlashWebhookUsecase](injector),
	}, nil
}

func slashWebhookDispatchRequest(
	accountID model.ID,
	event slash.WebhookEvent,
	resourceID model.ID,
) *biz.DispatchWebhookRequest {
	resourceIDString := idconv.ToUUID(resourceID)
	return &biz.DispatchWebhookRequest{
		AccountID: accountID,
		Event:     event,
		EntityID:  resourceIDString,
		EventID:   resourceIDString,
	}
}

type ListRequest struct {
	AccountID  *model.ID `form:"account_id" binding:"omitempty,gt=0"`
	PageNumber *int      `form:"page_number" binding:"omitempty,min=1"`
	PageSize   *int      `form:"page_size" binding:"omitempty,min=1"`
}

type AccountData struct {
	WalletID  model.ID        `json:"wallet_id"`
	ID        model.ID        `json:"id"`
	Name      string          `json:"name"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
}

type AuthorizationConfigData struct {
	AccountName   string    `json:"account_name"`
	AccountID     model.ID  `json:"account_id"`
	TargetURL     string    `json:"target_url"`
	Enabled       bool      `json:"enabled"`
	TimeoutMillis int       `json:"timeout_millis"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type GetAuthorizationConfigRequest struct {
	AccountID model.ID `form:"account_id" binding:"required,gt=0"`
}

func (s *SlashUIService) GetAuthorizationConfig(
	ctx context.Context,
	req *GetAuthorizationConfigRequest,
) (*AuthorizationConfigData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.GetAuthorizationConfig(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return authorizationConfigData(item), nil
}

type UpdateAuthorizationConfigRequest struct {
	AccountID     model.ID `json:"account_id" binding:"required,gt=0"`
	TargetURL     string   `json:"target_url" binding:"required,url"`
	Enabled       bool     `json:"enabled"`
	TimeoutMillis int      `json:"timeout_millis" binding:"required,min=1"`
}

func (s *SlashUIService) UpdateAuthorizationConfig(
	ctx context.Context,
	req *UpdateAuthorizationConfigRequest,
) (*AuthorizationConfigData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.UpdateAuthorizationConfig(ctx, &biz.UpdateAuthorizationConfigRequest{
		AccountID:     accountID,
		TargetURL:     req.TargetURL,
		Enabled:       req.Enabled,
		TimeoutMillis: req.TimeoutMillis,
	})
	if err != nil {
		return nil, err
	}
	return authorizationConfigData(item), nil
}

type CreateAccountRequest struct {
	Name string `json:"name" binding:"required"`
}

func (s *SlashUIService) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*AccountData, error) {
	item, err := s.usecase.CreateAccount(ctx, &biz.CreateAccountRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return slashAccountData(item), nil
}

type ListAccountsResponse struct {
	TotalItems int            `json:"total_items"`
	Data       []*AccountData `json:"data"`
}

func (s *SlashUIService) ListAccounts(
	ctx context.Context,
	req *ListRequest,
) (*ListAccountsResponse, error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	page, size := types.NormalizePagination(types.Value(req.PageNumber), types.Value(req.PageSize))
	items, total, err := s.usecase.ListAccounts(ctx, &biz.ListAccountsRequest{
		AccountID: accountID,
		Offset:    (page - 1) * size,
		Limit:     size,
	})
	if err != nil {
		return nil, err
	}

	return &ListAccountsResponse{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, slashAccountData),
	}, nil
}

type UpdateAccountRequest struct {
	ID   model.ID `uri:"id" binding:"required,gt=0"`
	Name string   `json:"name" binding:"required"`
}

func (s *SlashUIService) UpdateAccount(
	ctx context.Context,
	req *UpdateAccountRequest,
) (*AccountData, error) {
	id := req.ID
	if id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}

	item, err := s.usecase.UpdateAccount(ctx, &biz.UpdateAccountRequest{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}

	return slashAccountData(item), nil
}

type WebhookData struct {
	AccountName string             `json:"account_name"`
	ID          model.ID           `json:"id"`
	AccountID   model.ID           `json:"account_id"`
	Event       slash.WebhookEvent `json:"event"`
	TargetURL   string             `json:"target_url"`
	Enabled     bool               `json:"enabled"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type CreateWebhookRequest struct {
	AccountID model.ID           `json:"account_id" binding:"required,gt=0"`
	Event     slash.WebhookEvent `json:"event" binding:"required"`
	TargetURL string             `json:"target_url" binding:"required,url"`
	Enabled   bool               `json:"enabled"`
}

func (s *SlashUIService) ListWebhookEvents(context.Context, *struct{}) (*[]slash.WebhookEvent, error) {
	events := slash.WebhookEvents()
	return &events, nil
}

func (s *SlashUIService) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*WebhookData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}

	item, err := s.usecase.CreateWebhook(ctx, &biz.CreateWebhookRequest{
		AccountID: accountID,
		Event:     req.Event,
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return webhookData(item), nil
}

func (s *SlashUIService) ListWebhooks(ctx context.Context, req *struct {
	AccountID *model.ID `form:"account_id" binding:"omitempty,gt=0"`
}) (*[]WebhookData, error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	items, err := s.usecase.ListWebhooks(ctx, &biz.ListWebhooksRequest{
		AccountID: accountID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]WebhookData, 0, len(items))
	for _, item := range items {
		result = append(result, *webhookData(item))
	}
	return &result, nil
}

type UpdateWebhookRequest struct {
	ID        model.ID `uri:"id" binding:"required,gt=0"`
	TargetURL string   `json:"target_url" binding:"required,url"`
	Enabled   bool     `json:"enabled"`
}

func (s *SlashUIService) UpdateWebhook(ctx context.Context, req *UpdateWebhookRequest) (*WebhookData, error) {
	id := req.ID
	if id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.UpdateWebhook(ctx, &biz.UpdateWebhookRequest{
		ID:        id,
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return webhookData(item), nil
}

func (s *SlashUIService) DeleteWebhook(ctx context.Context, req *IDRequest) (*struct{}, error) {
	id := req.ID
	if id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	if err := s.usecase.DeleteWebhook(ctx, id); err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

type ListResponse[T any] struct {
	TotalItems int64 `json:"total_items"`
	Data       []T   `json:"data"`
}

type CardProductData struct {
	ID     model.ID `json:"id"`
	Prefix string   `json:"prefix"`
}

type ListCardProductsData struct {
	Items []CardProductData `json:"items"`
}

type ListCardProductsRequest struct{}

func (s *SlashUIService) ListCardProducts(ctx context.Context, _ *ListCardProductsRequest) (*ListCardProductsData, error) {
	items, err := s.usecase.ListCardProducts(ctx)
	if err != nil {
		return nil, err
	}

	return &ListCardProductsData{
		Items: types.BulkConvertSlice(items, func(item *biz.CardProductInfo) CardProductData {
			return CardProductData{
				ID:     item.Product.ID,
				Prefix: item.Product.Prefix,
			}
		}),
	}, nil
}

type VirtualAccountData struct {
	AccountID     model.ID  `json:"account_id"`
	AccountName   string    `json:"account_name"`
	ID            model.ID  `json:"id"`
	Name          string    `json:"name"`
	Currency      string    `json:"currency"`
	FundingSource string    `json:"funding_source"`
	Balance       string    `json:"balance"`
	Spend         string    `json:"spend"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *SlashUIService) ListVirtualAccounts(ctx context.Context, _ *struct{}) (*[]VirtualAccountData, error) {
	items, err := s.usecase.ListVirtualAccounts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]VirtualAccountData, 0, len(items))
	for _, item := range items {
		result = append(result, VirtualAccountData{
			ID:            item.ID,
			Name:          item.Name,
			Currency:      string(item.Wallet.Currency),
			FundingSource: "账户资金",
			Balance:       item.Wallet.Available.String(),
			Spend:         item.Wallet.Out.String(),
			CreatedAt:     item.CreatedAt,
		})
	}
	return &result, nil
}

type CardHolderRequest struct {
	AccountID model.ID `json:"account_id" binding:"required,gt=0"`
	FirstName string   `json:"first_name" binding:"required"`
	LastName  string   `json:"last_name" binding:"required"`
	Email     string   `json:"email" binding:"required"`
	Mobile    string   `json:"phone_number" binding:"required"`
}

type CardHolderData struct {
	AccountName string                  `json:"account_name"`
	AccountID   model.ID                `json:"account_id"`
	ID          model.ID                `json:"id"`
	FirstName   string                  `json:"first_name"`
	LastName    string                  `json:"last_name"`
	Email       string                  `json:"email"`
	Mobile      string                  `json:"phone_number"`
	Status      common.CardHolderStatus `json:"status"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

func (s *SlashUIService) CreateCardHolder(ctx context.Context, req *CardHolderRequest) (*CardHolderData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.CreateCardHolder(ctx, &biz.CreateCardHolderRequest{
		AccountID: accountID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Mobile:    req.Mobile,
	})
	if err != nil {
		return nil, err
	}
	return cardHolderData(item), nil
}

func (s *SlashUIService) ListCardHolders(ctx context.Context, req *ListRequest) (*ListResponse[*CardHolderData], error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	offset, limit := pagination(types.Value(req.PageNumber), types.Value(req.PageSize))
	items, total, err := s.usecase.ListCardHolders(ctx, &biz.ListCardHoldersRequest{
		AccountID: accountID,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	return &ListResponse[*CardHolderData]{
		TotalItems: total,
		Data:       types.BulkConvertSlice(items, cardHolderData),
	}, nil
}

type CreateCardRequest struct {
	AccountID     model.ID `json:"account_id" binding:"required,gt=0"`
	CardHolderID  model.ID `json:"cardholder_id" binding:"required,gt=0"`
	CardProductID model.ID `json:"card_product_id" binding:"required,gt=0"`
	CardCurrency  string   `json:"card_currency" binding:"required"`
}

type UpdateCardStatusRequest struct {
	AccountID  model.ID          `json:"account_id" binding:"required,gt=0"`
	CardStatus common.CardStatus `json:"card_status" binding:"required,oneof=active frozen deleted"`
}

type CardData struct {
	AccountName   string            `json:"account_name"`
	AccountID     model.ID          `json:"account_id"`
	WalletID      model.ID          `json:"wallet_id"`
	ID            model.ID          `json:"id"`
	CardHolderID  model.ID          `json:"cardholder_id"`
	CardProductID model.ID          `json:"card_product_id"`
	CardNumber    string            `json:"card_number"`
	Last4         string            `json:"last4"`
	CardBin       string            `json:"card_bin"`
	CardScheme    common.CardScheme `json:"card_scheme"`
	CardCurrency  string            `json:"card_currency"`
	FormFactor    string            `json:"form_factor"`
	CardStatus    common.CardStatus `json:"card_status"`
	Status        common.CardStatus `json:"status"`
	ExpiresAt     time.Time         `json:"expires_at"`
	Cvv           string            `json:"cvv"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	FundingSource string            `json:"funding_source"`
	Balance       string            `json:"balance"`
}

func (s *SlashUIService) CreateCard(ctx context.Context, req *CreateCardRequest) (*CardData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	cardHolderID := req.CardHolderID
	if cardHolderID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	cardProductID := req.CardProductID
	if cardProductID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.CreateCard(ctx, &biz.CreateCardRequest{
		AccountID:     accountID,
		CardHolderID:  cardHolderID,
		CardProductID: cardProductID,
		Currency:      common.Currency(req.CardCurrency),
	})
	if err != nil {
		return nil, err
	}
	s.webhookUsecase.Dispatch(ctx, slashWebhookDispatchRequest(
		item.AccountID,
		slash.WebhookEventCardCreate,
		item.ID,
	))
	return cardData(item), nil
}

type IDRequest struct {
	ID model.ID `uri:"id" binding:"required,gt=0"`
}

func (s *SlashUIService) GetCard(ctx context.Context, req *IDRequest) (*CardData, error) {
	id := req.ID
	if id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.GetCard(ctx, id)
	if err != nil {
		return nil, err
	}
	return cardData(item), nil
}

type UpdateCardRequest struct {
	IDRequest
	UpdateCardStatusRequest
}

func (s *SlashUIService) UpdateCardStatus(ctx context.Context, req *UpdateCardRequest) (*CardData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	id := req.ID
	if id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.UpdateCardStatus(ctx, &biz.UpdateCardStatusRequest{
		AccountID: accountID,
		ID:        id,
		Status:    req.CardStatus,
	})
	if err != nil {
		return nil, err
	}
	event := slash.WebhookEventCardUpdate
	if req.CardStatus == common.CardStatus_Deleted {
		event = slash.WebhookEventCardDelete
	}
	s.webhookUsecase.Dispatch(ctx, slashWebhookDispatchRequest(
		item.AccountID,
		event,
		item.ID,
	))
	return cardData(item), nil
}

type SimulateAuthorizationRequest struct {
	CardID               model.ID `json:"card_id" binding:"required,gt=0"`
	TransactionAmount    float64  `json:"transaction_amount" binding:"required"`
	TransactionCurrency  string   `json:"transaction_currency" binding:"required"`
	MerchantName         string   `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string   `json:"merchant_category_code" binding:"required"`
	MerchantCountry      *string  `json:"merchant_country" binding:"omitempty,min=1"`
	MerchantCity         string   `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

type SimulateAuthorizationData struct {
	Approved      bool                         `json:"approved"`
	Status        common.CardTransactionStatus `json:"status"`
	Authorization AuthorizationData            `json:"authorization"`
	Transaction   TransactionData              `json:"transaction"`
}

type SimulateRefundRequest struct {
	AuthorizationID      *model.ID        `json:"authorization_id" binding:"omitempty,gt=0"`
	CardID               *model.ID        `json:"card_id" binding:"omitempty,gt=0"`
	Amount               float64          `json:"amount" binding:"required,gt=0"`
	Currency             *common.Currency `json:"currency"`
	MerchantName         string           `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string           `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string           `json:"merchant_country" binding:"required"`
	MerchantCity         string           `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

func (s *SlashUIService) SimulateRefund(ctx context.Context, req *SimulateRefundRequest) (*TransactionData, error) {
	authorizationID := req.AuthorizationID
	if authorizationID != nil && *authorizationID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	cardID := req.CardID
	if cardID != nil && *cardID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.SimulateRefund(ctx, &biz.SimulateRefundRequest{
		Notificator:     s,
		AuthorizationID: authorizationID,
		CardID:          cardID,
		Amount:          decimal.NewFromFloat(req.Amount),
		Currency:        req.Currency,
		MerchantName:    &req.MerchantName,
		MerchantCountry: &req.MerchantCountry,
		MerchantMCC:     &req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}
	return transactionData(item), nil
}

func (s *SlashUIService) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*SimulateAuthorizationData, error) {
	cardID := req.CardID
	if cardID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	result, err := s.usecase.SimulateAuthorization(ctx, &biz.SimulateAuthorizationRequest{
		Notificator:     s,
		CardID:          cardID,
		Amount:          decimal.NewFromFloat(req.TransactionAmount),
		Currency:        common.Currency(req.TransactionCurrency),
		MerchantName:    &req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     &req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}
	return &SimulateAuthorizationData{
		Approved:      true,
		Status:        result.Authorization.Status,
		Authorization: *authorizationData(result.Authorization),
		Transaction:   *transactionData(result.CardTransaction),
	}, nil
}

type ListAuthorizationsRequest struct {
	ListRequest
	ID     *model.ID                     `form:"id" binding:"omitempty,gt=0"`
	CardID *model.ID                     `form:"card_id" binding:"omitempty,gt=0"`
	Status *common.CardTransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
}

type AuthorizationData struct {
	AccountID            model.ID                     `json:"account_id"`
	AccountName          string                       `json:"account_name"`
	ID                   model.ID                     `json:"id"`
	CardID               model.ID                     `json:"card_id"`
	Status               common.CardTransactionStatus `json:"status"`
	AuthorizedAmount     string                       `json:"authorized_amount"`
	Currency             string                       `json:"currency"`
	MerchantName         string                       `json:"merchant_name"`
	MerchantCategoryCode string                       `json:"merchant_category_code"`
	AuthorizationCode    string                       `json:"authorization_code"`
	AuthorizedAt         time.Time                    `json:"authorized_at"`
	CreatedAt            time.Time                    `json:"created_at"`
}

func (s *SlashUIService) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) (*ListResponse[*AuthorizationData], error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	offset, limit := pagination(types.Value(req.PageNumber), types.Value(req.PageSize))
	id := req.ID
	if id != nil && *id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	cardID := req.CardID
	if cardID != nil && *cardID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	items, total, err := s.usecase.ListAuthorizations(ctx, &biz.ListAuthorizationsRequest{
		AccountID: accountID,
		Offset:    offset,
		Limit:     limit,
		ID:        id,
		CardID:    cardID,
		Status:    req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &ListResponse[*AuthorizationData]{
		TotalItems: total,
		Data:       types.BulkConvertSlice(items, authorizationData),
	}, nil
}

func (s *SlashUIService) GetAuthorization(ctx context.Context, req *IDRequest) (*AuthorizationData, error) {
	id := req.ID
	if id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.GetAuthorization(ctx, id)
	if err != nil {
		return nil, err
	}
	return authorizationData(item), nil
}

type TransactionData struct {
	AccountName          string                       `json:"account_name"`
	AccountID            model.ID                     `json:"account_id"`
	ID                   model.ID                     `json:"id"`
	CardID               model.ID                     `json:"card_id"`
	AuthorizationID      model.ID                     `json:"authorization_id"`
	TransactionType      common.CardTransactionType   `json:"transaction_type"`
	Status               common.CardTransactionStatus `json:"status"`
	Amount               string                       `json:"amount"`
	Currency             string                       `json:"currency"`
	MerchantName         string                       `json:"merchant_name"`
	MerchantCountry      string                       `json:"merchant_country"`
	MerchantCategoryCode string                       `json:"merchant_category_code"`
	AuthorizationCode    string                       `json:"authorization_code"`
	TransactedAt         time.Time                    `json:"transacted_at"`
	CreatedAt            time.Time                    `json:"created_at"`
}

func (s *SlashUIService) GetTransaction(ctx context.Context, req *IDRequest) (*TransactionData, error) {
	id := req.ID
	if id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.GetCardTransaction(ctx, id)
	if err != nil {
		return nil, err
	}
	return transactionData(item), nil
}

type ApplyTransactionStepRequest struct {
	IDRequest
	Amount *decimal.Decimal `json:"amount"`
}

func (s *SlashUIService) ClearTransaction(ctx context.Context, req *ApplyTransactionStepRequest) (*TransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_CLEAR)
}

func (s *SlashUIService) ReverseTransaction(ctx context.Context, req *ApplyTransactionStepRequest) (*TransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_VOID)
}

func (s *SlashUIService) RefundTransaction(ctx context.Context, req *ApplyTransactionStepRequest) (*TransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_REFUND)
}

func (s *SlashUIService) applyTransactionStep(ctx context.Context, req *ApplyTransactionStepRequest, transactionType common.CardTransactionType) (*TransactionData, error) {
	id := req.ID
	if id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.ApplyTransactionStep(ctx, &biz.ApplyTransactionStepRequest{
		Notificator:       s,
		CardTransactionID: id,
		Type:              transactionType,
		Amount:            req.Amount,
	})
	if err != nil {
		return nil, err
	}
	return transactionData(item), nil
}

func cardHolderData(item *model.CardHolder) *CardHolderData {
	return &CardHolderData{
		AccountID:   item.AccountID,
		AccountName: uiAccountName(item.Account),
		ID:          item.ID,
		FirstName:   item.FirstName,
		LastName:    item.LastName,
		Email:       item.Email,
		Mobile:      item.Mobile,
		Status:      item.Status,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func webhookData(item *model.WebhookConfig) *WebhookData {
	return &WebhookData{
		ID:          item.ID,
		AccountID:   item.AccountID,
		AccountName: uiAccountName(item.Account),
		Event:       slash.WebhookEvent(item.Event),
		TargetURL:   item.TargetURL,
		Enabled:     item.Enabled,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func slashAccountData(item *model.Account) *AccountData {
	balance := decimal.Zero
	if item.Wallet != nil {
		balance = item.Wallet.Available
	}
	return &AccountData{
		WalletID:  item.WalletID,
		ID:        item.ID,
		Name:      item.Name,
		Balance:   balance,
		CreatedAt: item.CreatedAt,
	}
}

func authorizationConfigData(item *model.AuthorizationConfig) *AuthorizationConfigData {
	return &AuthorizationConfigData{
		AccountID:     item.AccountID,
		AccountName:   uiAccountName(item.Account),
		TargetURL:     item.TargetURL,
		Enabled:       item.Enabled,
		TimeoutMillis: item.TimeoutMillis,
		UpdatedAt:     item.UpdatedAt,
	}
}

func cardData(item *model.Card) *CardData {
	balance := decimal.Zero
	fundingSource := "卡资金"
	if item.Wallet != nil {
		balance = item.Wallet.Available
		if item.Wallet.Type == common.WalletType_VirtualAccount {
			fundingSource = "虚拟账户共享资金"
		}
	}

	return &CardData{
		AccountID:     item.AccountID,
		AccountName:   uiAccountName(item.Account),
		WalletID:      item.WalletID,
		ID:            item.ID,
		CardHolderID:  item.CardHolderID,
		CardProductID: item.CardProductID,
		CardNumber:    item.CardNumber,
		Last4:         last4(item.CardNumber),
		CardBin:       item.CardBin,
		CardScheme:    item.CardScheme,
		CardCurrency:  string(item.CardCurrency),
		FormFactor:    string(item.FormType),
		CardStatus:    item.Status,
		Status:        item.Status,
		ExpiresAt:     item.ExpireAt,
		Cvv:           item.Cvv,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
		FundingSource: fundingSource,
		Balance:       balance.String(),
	}
}

func authorizationData(item *model.Authorization) *AuthorizationData {
	return &AuthorizationData{
		AccountID:            item.AccountID,
		AccountName:          uiAccountName(item.Account),
		ID:                   item.ID,
		CardID:               item.CardID,
		Status:               item.Status,
		AuthorizedAmount:     item.Amount.String(),
		Currency:             string(item.Currency),
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		AuthorizationCode:    item.AuthorizationCode,
		AuthorizedAt:         item.CreatedAt,
		CreatedAt:            item.CreatedAt,
	}
}

func transactionData(item *model.CardTransaction) *TransactionData {
	return &TransactionData{
		AccountID:            item.AccountID,
		AccountName:          uiAccountName(item.Account),
		ID:                   item.ID,
		CardID:               item.CardID,
		AuthorizationID:      item.AuthorizationID,
		TransactionType:      item.Type,
		Status:               item.Status,
		Amount:               item.TxAmount.String(),
		Currency:             string(item.TxCurrency),
		MerchantName:         item.MerchantName,
		MerchantCountry:      item.MerchantCountry,
		MerchantCategoryCode: item.MerchantMCC,
		AuthorizationCode:    item.AuthorizationCode,
		TransactedAt:         item.CreatedAt,
		CreatedAt:            item.CreatedAt,
	}
}

func pagination(pageNumber int, pageSize int) (int, int) {
	page, size := types.NormalizePagination(pageNumber, pageSize)
	return (page - 1) * size, size
}

func last4(value string) string {
	if len(value) < 4 {
		return value
	}
	return value[len(value)-4:]
}

type ManagementListRequest struct {
	AccountID *model.ID `form:"account_id" binding:"omitempty,gt=0"`
}

type ManagementAccountRequest struct {
	AccountID model.ID `form:"account_id" json:"account_id" binding:"required,gt=0"`
}
type FundsData struct {
	AccountName string           `json:"account_name"`
	AccountID   model.ID         `json:"account_id"`
	ID          model.ID         `json:"id"`
	Currency    common.Currency  `json:"currency"`
	Kind        slash.WalletKind `json:"kind"`
	Amount      string           `json:"amount"`
}

func (s *SlashUIService) ListFunds(ctx context.Context, req *ManagementListRequest) (*[]FundsData, error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	items, err := s.usecase.ListFunds(ctx, accountID)
	if err != nil {
		return nil, err
	}
	result := make([]FundsData, 0, len(items))
	for _, item := range items {
		var kind slash.WalletKind
		switch item.Type {
		case common.WalletType_Account:
			kind = slash.WalletKindAccount
		case common.WalletType_VirtualAccount:
			kind = slash.WalletKindVirtualAccount
		case common.WalletType_Card:
			kind = slash.WalletKindCard
		}
		result = append(result, FundsData{
			AccountID:   item.AccountID,
			AccountName: uiAccountName(item.Account),
			ID:          item.ID,
			Currency:    item.Currency,
			Kind:        kind,
			Amount:      item.Available.String(),
		})
	}
	return &result, nil
}

type MoveFundsRequest struct {
	CardID *model.ID `json:"card_id" binding:"omitempty,gt=0"`
	ManagementAccountRequest
	SourceID *model.ID       `json:"source_id" binding:"omitempty,gt=0"`
	TargetID *model.ID       `json:"target_id" binding:"omitempty,gt=0"`
	Amount   decimal.Decimal `json:"amount"`
}

func (s *SlashUIService) MoveFunds(ctx context.Context, req *MoveFundsRequest) (*struct{}, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	if err := s.usecase.MoveFunds(ctx, &biz.MoveFundsRequest{
		CardID:    req.CardID,
		AccountID: accountID,
		SourceID:  req.SourceID,
		TargetID:  req.TargetID,
		Amount:    req.Amount,
	}); err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

type ManagedVirtualAccountData struct {
	AccountName string   `json:"account_name"`
	AccountID   model.ID `json:"account_id"`
	ID          model.ID `json:"id"`
	Name        string   `json:"name"`
	WalletID    model.ID `json:"wallet_id"`
}

func (s *SlashUIService) ListManagedVirtualAccounts(ctx context.Context, req *ManagementListRequest) (*[]ManagedVirtualAccountData, error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	items, err := s.usecase.ListManagedVirtualAccounts(ctx, accountID)
	if err != nil {
		return nil, err
	}
	result := make([]ManagedVirtualAccountData, 0, len(items))
	for _, item := range items {
		result = append(result, ManagedVirtualAccountData{
			AccountID:   item.AccountID,
			AccountName: uiAccountName(item.Account),
			ID:          item.ID,
			Name:        item.Name,
			WalletID:    item.WalletID,
		})
	}
	return &result, nil
}

type CreateManagedVirtualAccountRequest struct {
	ManagementAccountRequest
	Name     string          `json:"name" binding:"required"`
	Currency common.Currency `json:"currency" binding:"required,oneof=USD GBP JPY CNY"`
}

func (s *SlashUIService) CreateManagedVirtualAccount(ctx context.Context, req *CreateManagedVirtualAccountRequest) (*ManagedVirtualAccountData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.CreateManagedVirtualAccount(ctx, &biz.CreateManagedVirtualAccountRequest{
		AccountID: accountID,
		Name:      req.Name,
		Currency:  req.Currency,
	})
	if err != nil {
		return nil, err
	}
	return &ManagedVirtualAccountData{
		AccountID:   item.AccountID,
		AccountName: uiAccountName(item.Account),
		ID:          item.ID,
		Name:        item.Name,
		WalletID:    item.WalletID,
	}, nil
}

func uiAccountName(account *model.Account) string {
	if account == nil {
		return ""
	}
	return account.Name
}
