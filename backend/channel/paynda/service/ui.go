package service

import (
	"context"
	"time"

	"generic-mock/channel/paynda/biz"
	paynda "generic-mock/channel/paynda/enums"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
)

type PayndaUIService struct {
	usecase *biz.PayndaUIUsecase
}

func NewPayndaUIService(injector *do.Injector) (*PayndaUIService, error) {
	return &PayndaUIService{
		usecase: do.MustInvoke[*biz.PayndaUIUsecase](injector),
	}, nil
}

type PayndaUIListRequest struct {
	PageNumber int `form:"page_number"`
	PageSize   int `form:"page_size"`
}

type PayndaUIListResponse[T any] struct {
	TotalItems int `json:"total_items"`
	Data       []T `json:"data"`
}

type PayndaUIWebhookData struct {
	ID        string              `json:"id"`
	Event     paynda.WebhookEvent `json:"event"`
	TargetURL string              `json:"target_url"`
	Enabled   bool                `json:"enabled"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}
type PayndaUICreateWebhookRequest struct {
	Event     paynda.WebhookEvent `json:"event" binding:"required"`
	TargetURL string              `json:"target_url" binding:"required,url"`
	Enabled   bool                `json:"enabled"`
}
type PayndaUIUpdateWebhookRequest struct {
	ID        string `uri:"id" binding:"required"`
	TargetURL string `json:"target_url" binding:"required,url"`
	Enabled   bool   `json:"enabled"`
}

func (s *PayndaUIService) CreateWebhook(ctx context.Context, req *PayndaUICreateWebhookRequest) (*PayndaUIWebhookData, error) {
	item, err := s.usecase.CreateWebhook(ctx, &biz.PayndaUICreateWebhookRequest{Event: req.Event, TargetURL: req.TargetURL, Enabled: req.Enabled})
	if err != nil {
		return nil, err
	}
	return payndaUIWebhookData(item), nil
}
func (s *PayndaUIService) ListWebhooks(ctx context.Context, _ *struct{}) (*[]PayndaUIWebhookData, error) {
	items, err := s.usecase.ListWebhooks(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]PayndaUIWebhookData, 0, len(items))
	for _, item := range items {
		result = append(result, *payndaUIWebhookData(item))
	}
	return &result, nil
}

func (s *PayndaUIService) ListWebhookEvents(context.Context, *struct{}) (*[]paynda.WebhookEvent, error) {
	events := paynda.WebhookEvents()
	return &events, nil
}
func (s *PayndaUIService) UpdateWebhook(ctx context.Context, req *PayndaUIUpdateWebhookRequest) (*PayndaUIWebhookData, error) {
	id, err := payndaID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateWebhook(ctx, &biz.PayndaUIUpdateWebhookRequest{ID: id, TargetURL: req.TargetURL, Enabled: req.Enabled})
	if err != nil {
		return nil, err
	}
	return payndaUIWebhookData(item), nil
}
func (s *PayndaUIService) DeleteWebhook(ctx context.Context, req *struct {
	ID string `uri:"id" binding:"required"`
}) (*struct{}, error) {
	id, err := payndaID(req.ID)
	if err != nil {
		return nil, err
	}
	if err := s.usecase.DeleteWebhook(ctx, id); err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

type PayndaUICardHolderRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Mobile    string `json:"phone_number" binding:"required"`
}

type PayndaUICardHolderData struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Mobile    string    `json:"phone_number"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
type PayndaUICreateCardRequest struct {
	CardHolderID string `json:"cardholder_id" binding:"required"`
	CardCurrency string `json:"card_currency" binding:"required"`
}
type PayndaUICardData struct {
	ID           string    `json:"id"`
	CardHolderID string    `json:"cardholder_id"`
	CardNumber   string    `json:"card_number"`
	CardBin      string    `json:"card_bin"`
	CardCurrency string    `json:"card_currency"`
	CardStatus   string    `json:"card_status"`
	Cvv          string    `json:"cvv"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
type PayndaUIUpdateCardStatusRequest struct {
	ID         string            `uri:"id" binding:"required"`
	CardStatus paynda.CardStatus `json:"card_status" binding:"required"`
}

func (s *PayndaUIService) CreateCardHolder(ctx context.Context, req *PayndaUICardHolderRequest) (*PayndaUICardHolderData, error) {
	item, err := s.usecase.CreateCardHolder(ctx, &biz.PayndaUICreateCardHolderRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Mobile:    req.Mobile,
	})
	if err != nil {
		return nil, err
	}
	return payndaUICardHolderData(item), nil
}

func (s *PayndaUIService) ListCards(ctx context.Context, req *PayndaUIListRequest) (*PayndaUIListResponse[*PayndaUICardData], error) {
	items, err := s.usecase.ListCards(ctx, payndaUIListRequest(req))
	if err != nil {
		return nil, err
	}
	return &PayndaUIListResponse[*PayndaUICardData]{TotalItems: len(items), Data: types.BulkConvertSlice(items, payndaUICardData)}, nil
}

func (s *PayndaUIService) ListAuthorizations(ctx context.Context, req *PayndaUIListRequest) (*PayndaUIListResponse[*PayndaUIAuthorizationData], error) {
	items, err := s.usecase.ListAuthorizations(ctx, payndaUIListRequest(req))
	if err != nil {
		return nil, err
	}
	return &PayndaUIListResponse[*PayndaUIAuthorizationData]{TotalItems: len(items), Data: types.BulkConvertSlice(items, payndaUIAuthorizationData)}, nil
}

func (s *PayndaUIService) ListCardHolders(ctx context.Context, req *PayndaUIListRequest) (*PayndaUIListResponse[*PayndaUICardHolderData], error) {
	items, err := s.usecase.ListCardHolders(ctx, payndaUIListRequest(req))
	if err != nil {
		return nil, err
	}
	return &PayndaUIListResponse[*PayndaUICardHolderData]{TotalItems: len(items), Data: types.BulkConvertSlice(items, payndaUICardHolderData)}, nil
}

func (s *PayndaUIService) CreateCard(ctx context.Context, req *PayndaUICreateCardRequest) (*PayndaUICardData, error) {
	holderID, err := payndaID(req.CardHolderID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.CreateCard(ctx, &biz.PayndaUICreateCardRequest{CardHolderID: holderID, Currency: common.Currency(req.CardCurrency)})
	if err != nil {
		return nil, err
	}
	return payndaUICardData(item), nil
}

func (s *PayndaUIService) UpdateCardStatus(ctx context.Context, req *PayndaUIUpdateCardStatusRequest) (*PayndaUICardData, error) {
	id, err := payndaID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateCardStatus(ctx, &biz.PayndaUIUpdateCardStatusRequest{CardID: id, Status: paynda.CardStatusToGeneric(req.CardStatus)})
	if err != nil {
		return nil, err
	}
	return payndaUICardData(item), nil
}

type PayndaUISimulateAuthorizationRequest struct {
	CardID          string          `json:"card_id" binding:"required"`
	Amount          decimal.Decimal `json:"transaction_amount" binding:"required"`
	Currency        common.Currency `json:"transaction_currency" binding:"required"`
	MerchantName    string          `json:"merchant_name" binding:"required"`
	MerchantCountry string          `json:"merchant_country" binding:"required"`
	MerchantCity    string          `json:"merchant_city"` // Invalid: generic model has no merchant city field.
	MerchantMCC     string          `json:"merchant_category_code"`
}

type PayndaUISimulateAuthorizationData struct {
	Authorization   *PayndaUIAuthorizationData `json:"authorization"`
	CardTransaction *PayndaUITransactionData   `json:"transaction"`
}

type PayndaUISimulateRefundRequest struct {
	CardID               string          `json:"card_id" binding:"required"`
	Amount               decimal.Decimal `json:"amount" binding:"required"`
	Currency             common.Currency `json:"currency" binding:"required"`
	MerchantName         string          `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string          `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string          `json:"merchant_country" binding:"required"`
	MerchantCity         string          `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

func (s *PayndaUIService) SimulateRefund(ctx context.Context, req *PayndaUISimulateRefundRequest) (*PayndaUITransactionData, error) {
	cardID, err := payndaID(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.SimulateRefund(ctx, &biz.PayndaSimulateRefundRequest{
		CardID:          cardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}
	return payndaUITransactionData(item), nil
}

type PayndaUIAuthorizationData struct {
	ID                   string    `json:"id"`
	CardID               string    `json:"card_id"`
	Status               string    `json:"status"`
	AuthorizedAmount     string    `json:"authorized_amount"`
	Currency             string    `json:"currency"`
	MerchantName         string    `json:"merchant_name"`
	MerchantCategoryCode string    `json:"merchant_category_code"`
	AuthorizedAt         time.Time `json:"authorized_at"`
}
type PayndaUITransactionData struct {
	ID                   string    `json:"id"`
	CardID               string    `json:"card_id"`
	AuthorizationID      string    `json:"authorization_id"`
	TransactionType      string    `json:"transaction_type"`
	Status               string    `json:"status"`
	Amount               string    `json:"amount"`
	Currency             string    `json:"currency"`
	MerchantName         string    `json:"merchant_name"`
	MerchantCategoryCode string    `json:"merchant_category_code"`
	TransactedAt         time.Time `json:"transacted_at"`
}

func (s *PayndaUIService) SimulateAuthorization(
	ctx context.Context,
	req *PayndaUISimulateAuthorizationRequest,
) (*PayndaUISimulateAuthorizationData, error) {
	cardID, err := payndaID(req.CardID)
	if err != nil {
		return nil, err
	}
	result, err := s.usecase.SimulateAuthorization(ctx, &biz.PayndaSimulateAuthorizationRequest{
		CardID:          cardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
	})
	if err != nil {
		return nil, err
	}

	return &PayndaUISimulateAuthorizationData{
		Authorization:   payndaUIAuthorizationData(result.Authorization),
		CardTransaction: payndaUITransactionData(result.CardTransaction),
	}, nil
}

type PayndaUIApplyTransactionStepRequest struct {
	ID     string          `uri:"id" binding:"required"`
	Amount decimal.Decimal `json:"amount"`
}

func (s *PayndaUIService) ClearTransaction(ctx context.Context, req *PayndaUIApplyTransactionStepRequest) (*PayndaUITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_CLEAR)
}
func (s *PayndaUIService) ReverseTransaction(ctx context.Context, req *PayndaUIApplyTransactionStepRequest) (*PayndaUITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_VOID)
}
func (s *PayndaUIService) RefundTransaction(ctx context.Context, req *PayndaUIApplyTransactionStepRequest) (*PayndaUITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_REFUND)
}
func (s *PayndaUIService) ListTransactions(ctx context.Context, req *PayndaUIListRequest) (*PayndaUIListResponse[*PayndaUITransactionData], error) {
	items, err := s.usecase.ListTransactions(ctx, payndaUIListRequest(req))
	if err != nil {
		return nil, err
	}
	return &PayndaUIListResponse[*PayndaUITransactionData]{TotalItems: len(items), Data: types.BulkConvertSlice(items, payndaUITransactionData)}, nil
}
func (s *PayndaUIService) applyTransactionStep(
	ctx context.Context,
	req *PayndaUIApplyTransactionStepRequest,
	kind common.CardTransactionType,
) (*PayndaUITransactionData, error) {
	id, err := payndaID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.ApplyTransactionStep(ctx, &biz.PayndaUIApplyTransactionStepRequest{CardTransactionID: id, Type: kind, Amount: req.Amount})
	if err != nil {
		return nil, err
	}
	return payndaUITransactionData(item), nil
}
func payndaUIListRequest(req *PayndaUIListRequest) *biz.PayndaListRequest {
	page, size := types.NormalizePagination(req.PageNumber, req.PageSize)
	return &biz.PayndaListRequest{Offset: (page - 1) * size, Limit: size}
}
func payndaUICardHolderData(item *model.CardHolder) *PayndaUICardHolderData {
	return &PayndaUICardHolderData{
		ID:        payndaIDString(item.ID),
		FirstName: item.FirstName,
		LastName:  item.LastName,
		Email:     item.Email,
		Mobile:    item.Mobile,
		Status:    string(item.Status),
		CreatedAt: item.CreatedAt,
	}
}
func payndaUICardData(item *model.Card) *PayndaUICardData {
	return &PayndaUICardData{
		ID:           payndaIDString(item.ID),
		CardHolderID: payndaIDString(item.CardHolderID),
		CardNumber:   item.CardNumber,
		CardBin:      item.CardBin,
		CardCurrency: string(item.CardCurrency),
		CardStatus:   string(paynda.CardStatusFromGeneric(item.Status)),
		Cvv:          item.Cvv,
		ExpiresAt:    item.ExpireAt,
		CreatedAt:    item.CreatedAt,
	}
}
func payndaUIAuthorizationData(item *model.Authorization) *PayndaUIAuthorizationData {
	return &PayndaUIAuthorizationData{
		ID:                   payndaIDString(item.ID),
		CardID:               payndaIDString(item.CardID),
		Status:               string(item.Status),
		AuthorizedAmount:     item.Amount.String(),
		Currency:             string(item.Currency),
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		AuthorizedAt:         item.OccurredAt,
	}
}
func payndaUIWebhookData(item *model.WebhookConfig) *PayndaUIWebhookData {
	return &PayndaUIWebhookData{ID: payndaIDString(item.ID), Event: paynda.WebhookEvent(item.Event), TargetURL: item.TargetURL, Enabled: item.Enabled, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
func payndaUITransactionData(item *model.CardTransaction) *PayndaUITransactionData {
	return &PayndaUITransactionData{
		ID:                   payndaIDString(item.ID),
		CardID:               payndaIDString(item.CardID),
		AuthorizationID:      payndaIDString(item.AuthorizationID),
		TransactionType:      string(paynda.TransactionTypeFromGeneric(item.Type)),
		Status:               string(item.Status),
		Amount:               item.TxAmount.String(),
		Currency:             string(item.TxCurrency),
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		TransactedAt:         item.OccurredAt,
	}
}
