package service

import (
	"context"
	"time"

	"generic-mock/channel/slash/biz"
	slash "generic-mock/channel/slash/enums"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
)

type SlashUIService struct {
	usecase *biz.SlashUIUsecase
}

func NewSlashUIService(injector *do.Injector) (*SlashUIService, error) {
	return &SlashUIService{
		usecase: do.MustInvoke[*biz.SlashUIUsecase](injector),
	}, nil
}

type ListRequest struct {
	PageNumber int `form:"page_number"`
	PageSize   int `form:"page_size"`
}

type AccountData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
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

func (s *SlashUIService) ListAccounts(ctx context.Context, _ *struct{}) (*[]AccountData, error) {
	items, err := s.usecase.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]AccountData, 0, len(items))
	for _, item := range items {
		result = append(result, *slashAccountData(item))
	}
	return &result, nil
}

type WebhookData struct {
	ID        string    `json:"id"`
	Event     string    `json:"event"`
	TargetURL string    `json:"target_url"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateWebhookRequest struct {
	Event     string `json:"event" binding:"required"`
	TargetURL string `json:"target_url" binding:"required,url"`
	Enabled   bool   `json:"enabled"`
}

func (s *SlashUIService) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*WebhookData, error) {
	item, err := s.usecase.CreateWebhook(ctx, &biz.CreateWebhookRequest{Event: req.Event, TargetURL: req.TargetURL, Enabled: req.Enabled})
	if err != nil {
		return nil, err
	}
	return webhookData(item), nil
}

func (s *SlashUIService) ListWebhooks(ctx context.Context, _ *struct{}) (*[]WebhookData, error) {
	items, err := s.usecase.ListWebhooks(ctx)
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
	ID        string `uri:"id" binding:"required"`
	TargetURL string `json:"target_url" binding:"required,url"`
	Enabled   bool   `json:"enabled"`
}

func (s *SlashUIService) UpdateWebhook(ctx context.Context, req *UpdateWebhookRequest) (*WebhookData, error) {
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateWebhook(ctx, &biz.UpdateWebhookRequest{ID: id, TargetURL: req.TargetURL, Enabled: req.Enabled})
	if err != nil {
		return nil, err
	}
	return webhookData(item), nil
}

func (s *SlashUIService) DeleteWebhook(ctx context.Context, req *IDRequest) (*struct{}, error) {
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
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
	ID        string `json:"id"`
	Prefix    string `json:"prefix"`
	IsDefault bool   `json:"is_default"`
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
				ID:        slashIDString(item.Product.ID),
				Prefix:    item.Product.Prefix,
				IsDefault: item.Product.IsDefault,
			}
		}),
	}, nil
}

type VirtualAccountData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	Balance   string    `json:"balance"`
	Spend     string    `json:"spend"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *SlashUIService) ListVirtualAccounts(ctx context.Context, _ *struct{}) (*[]VirtualAccountData, error) {
	items, err := s.usecase.ListVirtualAccounts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]VirtualAccountData, 0, len(items))
	for _, item := range items {
		result = append(result, VirtualAccountData{
			ID: slashIDString(item.ID), Name: item.Name, Currency: string(item.Wallet.Currency),
			Balance: item.Wallet.Amount.String(), Spend: item.Wallet.Out.String(), CreatedAt: item.CreatedAt,
		})
	}
	return &result, nil
}

type CardHolderRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Mobile    string `json:"phone_number" binding:"required"`
}

type CardHolderData struct {
	ID        string                 `json:"id"`
	FirstName string                 `json:"first_name"`
	LastName  string                 `json:"last_name"`
	Email     string                 `json:"email"`
	Mobile    string                 `json:"phone_number"`
	Status    slash.CardHolderStatus `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func (s *SlashUIService) CreateCardHolder(ctx context.Context, req *CardHolderRequest) (*CardHolderData, error) {
	item, err := s.usecase.CreateCardHolder(ctx, &biz.CreateCardHolderRequest{
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
	offset, limit := pagination(req.PageNumber, req.PageSize)
	items, total, err := s.usecase.ListCardHolders(ctx, &biz.ListCardHoldersRequest{
		Offset: offset,
		Limit:  limit,
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
	CardHolderID  string `json:"cardholder_id" binding:"required"`
	CardProductID string `json:"card_product_id"`
	CardCurrency  string `json:"card_currency" binding:"required"`
}

type UpdateCardStatusRequest struct {
	CardStatus slash.CardStatus `json:"card_status" binding:"required"`
}

type CardData struct {
	ID            string            `json:"id"`
	CardHolderID  string            `json:"cardholder_id"`
	CardProductID string            `json:"card_product_id"`
	CardNumber    string            `json:"card_number"`
	Last4         string            `json:"last4"`
	CardBin       string            `json:"card_bin"`
	CardScheme    common.CardScheme `json:"card_scheme"`
	CardCurrency  string            `json:"card_currency"`
	FormFactor    string            `json:"form_factor"`
	CardStatus    slash.CardStatus  `json:"card_status"`
	Status        slash.CardStatus  `json:"status"`
	ExpiresAt     time.Time         `json:"expires_at"`
	Cvv           string            `json:"cvv"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

func (s *SlashUIService) CreateCard(ctx context.Context, req *CreateCardRequest) (*CardData, error) {
	cardHolderID, err := slashID(req.CardHolderID)
	if err != nil {
		return nil, err
	}
	cardProductID, err := slashID(req.CardProductID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.CreateCard(ctx, &biz.CreateCardRequest{
		CardHolderID:  cardHolderID,
		CardProductID: cardProductID,
		Currency:      common.Currency(req.CardCurrency),
	})
	if err != nil {
		return nil, err
	}
	return cardData(item), nil
}

type ListCardsRequest struct {
	ListRequest
	ID         string           `form:"id"`
	CardNumber string           `form:"card_number"`
	CardStatus slash.CardStatus `form:"card_status"`
}

func (s *SlashUIService) ListCards(ctx context.Context, req *ListCardsRequest) (*ListResponse[*CardData], error) {
	offset, limit := pagination(req.PageNumber, req.PageSize)
	items, total, err := s.usecase.ListCards(ctx, &biz.ListCardsRequest{
		Offset:     offset,
		Limit:      limit,
		IDContains: req.ID,
		CardNumber: req.CardNumber,
		Status:     slash.CardStatusToGeneric(req.CardStatus),
	})
	if err != nil {
		return nil, err
	}
	return &ListResponse[*CardData]{
		TotalItems: total,
		Data:       types.BulkConvertSlice(items, cardData),
	}, nil
}

type IDRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (s *SlashUIService) GetCard(ctx context.Context, req *IDRequest) (*CardData, error) {
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
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
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateCardStatus(ctx, &biz.UpdateCardStatusRequest{
		ID:     id,
		Status: slash.CardStatusToGeneric(req.CardStatus),
	})
	if err != nil {
		return nil, err
	}
	return cardData(item), nil
}

type SimulateAuthorizationRequest struct {
	CardID               string  `json:"card_id" binding:"required"`
	TransactionAmount    float64 `json:"transaction_amount" binding:"required"`
	TransactionCurrency  string  `json:"transaction_currency" binding:"required"`
	MerchantName         string  `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string  `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string  `json:"merchant_country"`
	MerchantCity         string  `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

type SimulateAuthorizationData struct {
	Approved      bool                    `json:"approved"`
	Status        slash.TransactionStatus `json:"status"`
	Authorization AuthorizationData       `json:"authorization"`
	Transaction   TransactionData         `json:"transaction"`
}

type SimulateRefundRequest struct {
	CardID               string  `json:"card_id" binding:"required"`
	Amount               float64 `json:"amount" binding:"required,gt=0"`
	Currency             string  `json:"currency" binding:"required"`
	MerchantName         string  `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string  `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string  `json:"merchant_country" binding:"required"`
	MerchantCity         string  `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

func (s *SlashUIService) SimulateRefund(ctx context.Context, req *SimulateRefundRequest) (*TransactionData, error) {
	cardID, err := slashID(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.SimulateRefund(ctx, &biz.SimulateRefundRequest{
		CardID:          cardID,
		Amount:          decimal.NewFromFloat(req.Amount),
		Currency:        common.Currency(req.Currency),
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}
	return transactionData(item), nil
}

func (s *SlashUIService) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*SimulateAuthorizationData, error) {
	cardID, err := slashID(req.CardID)
	if err != nil {
		return nil, err
	}
	result, err := s.usecase.SimulateAuthorization(ctx, &biz.SimulateAuthorizationRequest{
		CardID:          cardID,
		Amount:          decimal.NewFromFloat(req.TransactionAmount),
		Currency:        common.Currency(req.TransactionCurrency),
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}
	return &SimulateAuthorizationData{
		Approved:      true,
		Status:        slash.TransactionStatusFromGeneric(result.Authorization.Status),
		Authorization: *authorizationData(result.Authorization),
		Transaction:   *transactionData(result.CardTransaction),
	}, nil
}

type ListAuthorizationsRequest struct {
	ListRequest
	ID     string                  `form:"id"`
	CardID string                  `form:"card_id"`
	Status slash.TransactionStatus `form:"status"`
}

type AuthorizationData struct {
	ID                   string                  `json:"id"`
	CardID               string                  `json:"card_id"`
	Status               slash.TransactionStatus `json:"status"`
	AuthorizedAmount     string                  `json:"authorized_amount"`
	Currency             string                  `json:"currency"`
	MerchantName         string                  `json:"merchant_name"`
	MerchantCategoryCode string                  `json:"merchant_category_code"`
	AuthorizationCode    string                  `json:"authorization_code"`
	AuthorizedAt         time.Time               `json:"authorized_at"`
	CreatedAt            time.Time               `json:"created_at"`
}

func (s *SlashUIService) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) (*ListResponse[*AuthorizationData], error) {
	offset, limit := pagination(req.PageNumber, req.PageSize)
	id, err := slashOptionalID(req.ID)
	if err != nil {
		return nil, err
	}
	cardID, err := slashOptionalID(req.CardID)
	if err != nil {
		return nil, err
	}
	items, total, err := s.usecase.ListAuthorizations(ctx, &biz.ListAuthorizationsRequest{
		Offset: offset,
		Limit:  limit,
		ID:     id,
		CardID: cardID,
		Status: slash.TransactionStatusToGeneric(req.Status),
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
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetAuthorization(ctx, id)
	if err != nil {
		return nil, err
	}
	return authorizationData(item), nil
}

type ListTransactionsRequest struct {
	ListRequest
	ID              string                  `form:"id"`
	CardID          string                  `form:"card_id"`
	AuthorizationID string                  `form:"authorization_id"`
	TransactionType slash.TransactionType   `form:"transaction_type"`
	Status          slash.TransactionStatus `form:"status"`
}

type TransactionData struct {
	ID                   string                  `json:"id"`
	CardID               string                  `json:"card_id"`
	AuthorizationID      string                  `json:"authorization_id"`
	TransactionType      slash.TransactionType   `json:"transaction_type"`
	Status               slash.TransactionStatus `json:"status"`
	Amount               string                  `json:"amount"`
	Currency             string                  `json:"currency"`
	MerchantName         string                  `json:"merchant_name"`
	MerchantCountry      string                  `json:"merchant_country"`
	MerchantCategoryCode string                  `json:"merchant_category_code"`
	AuthorizationCode    string                  `json:"authorization_code"`
	TransactedAt         time.Time               `json:"transacted_at"`
	CreatedAt            time.Time               `json:"created_at"`
}

func (s *SlashUIService) ListTransactions(ctx context.Context, req *ListTransactionsRequest) (*ListResponse[*TransactionData], error) {
	offset, limit := pagination(req.PageNumber, req.PageSize)
	id, err := slashOptionalID(req.ID)
	if err != nil {
		return nil, err
	}
	cardID, err := slashOptionalID(req.CardID)
	if err != nil {
		return nil, err
	}
	authorizationID, err := slashOptionalID(req.AuthorizationID)
	if err != nil {
		return nil, err
	}
	items, total, err := s.usecase.ListCardTransactions(ctx, &biz.ListCardTransactionsRequest{
		Offset:          offset,
		Limit:           limit,
		ID:              id,
		CardID:          cardID,
		AuthorizationID: authorizationID,
		Type:            slash.TransactionTypeToGeneric(req.TransactionType),
		Status:          slash.TransactionStatusToGeneric(req.Status),
	})
	if err != nil {
		return nil, err
	}
	return &ListResponse[*TransactionData]{
		TotalItems: total,
		Data:       types.BulkConvertSlice(items, transactionData),
	}, nil
}

func (s *SlashUIService) GetTransaction(ctx context.Context, req *IDRequest) (*TransactionData, error) {
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCardTransaction(ctx, id)
	if err != nil {
		return nil, err
	}
	return transactionData(item), nil
}

type ApplyTransactionStepRequest struct {
	IDRequest
	Amount float64 `json:"amount"`
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
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.ApplyTransactionStep(ctx, &biz.ApplyTransactionStepRequest{
		CardTransactionID: id,
		Type:              transactionType,
		Amount:            decimal.NewFromFloat(req.Amount),
	})
	if err != nil {
		return nil, err
	}
	return transactionData(item), nil
}

func cardHolderData(item *model.CardHolder) *CardHolderData {
	return &CardHolderData{
		ID:        slashIDString(item.ID),
		FirstName: item.FirstName,
		LastName:  item.LastName,
		Email:     item.Email,
		Mobile:    item.Mobile,
		Status:    slash.CardHolderStatusFromGeneric(item.Status),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func webhookData(item *model.WebhookConfig) *WebhookData {
	return &WebhookData{
		ID:        slashIDString(item.ID),
		Event:     item.Event,
		TargetURL: item.TargetURL,
		Enabled:   item.Enabled,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func slashAccountData(item *model.Account) *AccountData {
	return &AccountData{ID: slashIDString(item.ID), Name: item.Name, CreatedAt: item.CreatedAt}
}

func cardData(item *model.Card) *CardData {
	return &CardData{
		ID:            slashIDString(item.ID),
		CardHolderID:  slashIDString(item.CardHolderID),
		CardProductID: slashIDString(item.CardProductID),
		CardNumber:    item.CardNumber,
		Last4:         last4(item.CardNumber),
		CardBin:       item.CardBin,
		CardScheme:    item.CardScheme,
		CardCurrency:  string(item.CardCurrency),
		FormFactor:    string(item.FormType),
		CardStatus:    slash.CardStatusFromGeneric(item.Status),
		Status:        slash.CardStatusFromGeneric(item.Status),
		ExpiresAt:     item.ExpireAt,
		Cvv:           item.Cvv,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func authorizationData(item *model.Authorization) *AuthorizationData {
	return &AuthorizationData{
		ID:                   slashIDString(item.ID),
		CardID:               slashIDString(item.CardID),
		Status:               slash.TransactionStatusFromGeneric(item.Status),
		AuthorizedAmount:     item.Amount.String(),
		Currency:             string(item.Currency),
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		AuthorizationCode:    item.AuthorizationCode,
		AuthorizedAt:         item.OccurredAt,
		CreatedAt:            item.CreatedAt,
	}
}

func transactionData(item *model.CardTransaction) *TransactionData {
	return &TransactionData{
		ID:                   slashIDString(item.ID),
		CardID:               slashIDString(item.CardID),
		AuthorizationID:      slashIDString(item.AuthorizationID),
		TransactionType:      slash.TransactionTypeFromGeneric(item.Type),
		Status:               slash.TransactionStatusFromGeneric(item.Status),
		Amount:               item.TxAmount.String(),
		Currency:             string(item.TxCurrency),
		MerchantName:         item.MerchantName,
		MerchantCountry:      item.MerchantCountry,
		MerchantCategoryCode: item.MerchantMCC,
		AuthorizationCode:    item.AuthorizationCode,
		TransactedAt:         item.OccurredAt,
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
