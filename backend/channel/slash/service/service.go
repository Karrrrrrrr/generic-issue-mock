package service

import (
	"context"
	"time"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
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
				ID:        item.Product.ID,
				Prefix:    item.Product.Prefix,
				IsDefault: item.Product.IsDefault,
			}
		}),
	}, nil
}

type CardHolderRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Mobile    string `json:"phone_number" binding:"required"`
}

type CardHolderData struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Mobile    string    `json:"phone_number"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	CardStatus string `json:"card_status" binding:"required"`
}

type CardData struct {
	ID            string    `json:"id"`
	CardHolderID  string    `json:"cardholder_id"`
	CardProductID string    `json:"card_product_id"`
	CardNumber    string    `json:"card_number"`
	Last4         string    `json:"last4"`
	CardBin       string    `json:"card_bin"`
	CardScheme    string    `json:"card_scheme"`
	CardCurrency  string    `json:"card_currency"`
	FormFactor    string    `json:"form_factor"`
	CardStatus    string    `json:"card_status"`
	Status        string    `json:"status"`
	ExpiresAt     time.Time `json:"expires_at"`
	Cvv           string    `json:"cvv"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (s *SlashUIService) CreateCard(ctx context.Context, req *CreateCardRequest) (*CardData, error) {
	item, err := s.usecase.CreateCard(ctx, &biz.CreateCardRequest{
		CardHolderID:  model.ID(req.CardHolderID),
		CardProductID: model.ID(req.CardProductID),
		Currency:      enums.Currency(req.CardCurrency),
	})
	if err != nil {
		return nil, err
	}
	return cardData(item), nil
}

type ListCardsRequest struct {
	ListRequest
	ID         string `form:"id"`
	CardNumber string `form:"card_number"`
	CardStatus string `form:"card_status"`
}

func (s *SlashUIService) ListCards(ctx context.Context, req *ListCardsRequest) (*ListResponse[*CardData], error) {
	offset, limit := pagination(req.PageNumber, req.PageSize)
	items, total, err := s.usecase.ListCards(ctx, &biz.ListCardsRequest{
		Offset:     offset,
		Limit:      limit,
		IDContains: req.ID,
		CardNumber: req.CardNumber,
		Status:     enums.CardStatus(req.CardStatus),
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
	item, err := s.usecase.GetCard(ctx, model.ID(req.ID))
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
	item, err := s.usecase.UpdateCardStatus(ctx, &biz.UpdateCardStatusRequest{
		ID:     model.ID(req.ID),
		Status: enums.CardStatus(req.CardStatus),
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
}

type SimulateAuthorizationData struct {
	Approved      bool              `json:"approved"`
	Status        string            `json:"status"`
	Authorization AuthorizationData `json:"authorization"`
	Transaction   TransactionData   `json:"transaction"`
}

func (s *SlashUIService) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*SimulateAuthorizationData, error) {
	result, err := s.usecase.SimulateAuthorization(ctx, &biz.SimulateAuthorizationRequest{
		CardID:          model.ID(req.CardID),
		Amount:          decimal.NewFromFloat(req.TransactionAmount),
		Currency:        enums.Currency(req.TransactionCurrency),
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}
	return &SimulateAuthorizationData{
		Approved:      true,
		Status:        string(result.Authorization.Status),
		Authorization: *authorizationData(result.Authorization),
		Transaction:   *transactionData(result.CardTransaction),
	}, nil
}

type ListAuthorizationsRequest struct {
	ListRequest
	ID     string `form:"id"`
	CardID string `form:"card_id"`
	Status string `form:"status"`
}

type AuthorizationData struct {
	ID                   string    `json:"id"`
	CardID               string    `json:"card_id"`
	Status               string    `json:"status"`
	AuthorizedAmount     string    `json:"authorized_amount"`
	Currency             string    `json:"currency"`
	MerchantName         string    `json:"merchant_name"`
	MerchantCategoryCode string    `json:"merchant_category_code"`
	AuthorizationCode    string    `json:"authorization_code"`
	AuthorizedAt         time.Time `json:"authorized_at"`
	CreatedAt            time.Time `json:"created_at"`
}

func (s *SlashUIService) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) (*ListResponse[*AuthorizationData], error) {
	offset, limit := pagination(req.PageNumber, req.PageSize)
	items, total, err := s.usecase.ListAuthorizations(ctx, &biz.ListAuthorizationsRequest{
		Offset: offset,
		Limit:  limit,
		ID:     model.ID(req.ID),
		CardID: model.ID(req.CardID),
		Status: enums.CardTransactionStatus(req.Status),
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
	item, err := s.usecase.GetAuthorization(ctx, model.ID(req.ID))
	if err != nil {
		return nil, err
	}
	return authorizationData(item), nil
}

type ListTransactionsRequest struct {
	ListRequest
	ID              string `form:"id"`
	CardID          string `form:"card_id"`
	AuthorizationID string `form:"authorization_id"`
	TransactionType string `form:"transaction_type"`
	Status          string `form:"status"`
}

type TransactionData struct {
	ID                   string    `json:"id"`
	CardID               string    `json:"card_id"`
	AuthorizationID      string    `json:"authorization_id"`
	TransactionType      string    `json:"transaction_type"`
	Status               string    `json:"status"`
	Amount               string    `json:"amount"`
	Currency             string    `json:"currency"`
	MerchantName         string    `json:"merchant_name"`
	MerchantCountry      string    `json:"merchant_country"`
	MerchantCategoryCode string    `json:"merchant_category_code"`
	AuthorizationCode    string    `json:"authorization_code"`
	TransactedAt         time.Time `json:"transacted_at"`
	CreatedAt            time.Time `json:"created_at"`
}

func (s *SlashUIService) ListTransactions(ctx context.Context, req *ListTransactionsRequest) (*ListResponse[*TransactionData], error) {
	offset, limit := pagination(req.PageNumber, req.PageSize)
	items, total, err := s.usecase.ListCardTransactions(ctx, &biz.ListCardTransactionsRequest{
		Offset:          offset,
		Limit:           limit,
		ID:              model.ID(req.ID),
		CardID:          model.ID(req.CardID),
		AuthorizationID: model.ID(req.AuthorizationID),
		Type:            enums.CardTransactionType(req.TransactionType),
		Status:          enums.CardTransactionStatus(req.Status),
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
	item, err := s.usecase.GetCardTransaction(ctx, model.ID(req.ID))
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
	return s.applyTransactionStep(ctx, req, enums.CardTransactionType_CLEAR)
}

func (s *SlashUIService) ReverseTransaction(ctx context.Context, req *ApplyTransactionStepRequest) (*TransactionData, error) {
	return s.applyTransactionStep(ctx, req, enums.CardTransactionType_VOID)
}

func (s *SlashUIService) RefundTransaction(ctx context.Context, req *ApplyTransactionStepRequest) (*TransactionData, error) {
	return s.applyTransactionStep(ctx, req, enums.CardTransactionType_REFUND)
}

func (s *SlashUIService) applyTransactionStep(ctx context.Context, req *ApplyTransactionStepRequest, transactionType enums.CardTransactionType) (*TransactionData, error) {
	item, err := s.usecase.ApplyTransactionStep(ctx, &biz.ApplyTransactionStepRequest{
		CardTransactionID: model.ID(req.ID),
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
		ID:        item.ID,
		FirstName: item.FirstName,
		LastName:  item.LastName,
		Email:     item.Email,
		Mobile:    item.Mobile,
		Status:    string(item.Status),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func cardData(item *model.Card) *CardData {
	return &CardData{
		ID:            item.ID,
		CardHolderID:  item.CardHolderID,
		CardProductID: item.CardProductID,
		CardNumber:    item.CardNumber,
		Last4:         last4(item.CardNumber),
		CardBin:       item.CardBin,
		CardScheme:    item.CardScheme,
		CardCurrency:  string(item.CardCurrency),
		FormFactor:    string(item.FormType),
		CardStatus:    string(item.Status),
		Status:        string(item.Status),
		ExpiresAt:     item.ExpireAt,
		Cvv:           item.Cvv,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func authorizationData(item *model.Authorization) *AuthorizationData {
	return &AuthorizationData{
		ID:                   item.ID,
		CardID:               item.CardID,
		Status:               string(item.Status),
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
		ID:                   item.ID,
		CardID:               item.CardID,
		AuthorizationID:      item.AuthorizationID,
		TransactionType:      string(item.Type),
		Status:               string(item.Status),
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
