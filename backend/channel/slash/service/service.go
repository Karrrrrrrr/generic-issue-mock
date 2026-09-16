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

type Service struct {
	usecase *biz.Usecase
}

func NewService(injector *do.Injector) (*Service, error) {
	return &Service{
		usecase: do.MustInvokeNamed[*biz.Usecase](injector, "slash.usecase"),
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

type CardHolderRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Mobile    string `json:"phone_number" binding:"required"`
}

type CardHolderData struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Mobile    string `json:"phone_number"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Service) CreateCardHolder(ctx context.Context, req *CardHolderRequest) (*CardHolderData, error) {
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

func (s *Service) ListCardHolders(ctx context.Context, req *ListRequest) (*ListResponse[CardHolderData], error) {
	offset, limit := pagination(req.PageNumber, req.PageSize)
	items, total, err := s.usecase.ListCardHolders(ctx, &biz.ListCardHoldersRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	return &ListResponse[CardHolderData]{
		TotalItems: total,
		Data: types.BulkConvertSlice(items, func(item *model.CardHolder) CardHolderData {
			return *cardHolderData(item)
		}),
	}, nil
}

type CreateCardRequest struct {
	CardHolderID string `json:"cardholder_id" binding:"required"`
	CardCurrency string `json:"card_currency" binding:"required"`
}

type UpdateCardStatusRequest struct {
	CardStatus string `json:"card_status" binding:"required"`
}

type CardData struct {
	ID           string `json:"id"`
	CardHolderID string `json:"cardholder_id"`
	CardNumber   string `json:"card_number"`
	Last4        string `json:"last4"`
	CardBin      string `json:"card_bin"`
	CardScheme   string `json:"card_scheme"`
	CardCurrency string `json:"card_currency"`
	FormFactor   string `json:"form_factor"`
	CardStatus   string `json:"card_status"`
	Status       string `json:"status"`
	ExpiryMonth  string `json:"expiry_month"`
	ExpiryYear   string `json:"expiry_year"`
	Cvv          string `json:"cvv"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func (s *Service) CreateCard(ctx context.Context, req *CreateCardRequest) (*CardData, error) {
	item, err := s.usecase.CreateCard(ctx, &biz.CreateCardRequest{
		CardHolderID: model.ID(req.CardHolderID),
		Currency:     enums.Currency(req.CardCurrency),
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

func (s *Service) ListCards(ctx context.Context, req *ListCardsRequest) (*ListResponse[CardData], error) {
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
	return &ListResponse[CardData]{
		TotalItems: total,
		Data: types.BulkConvertSlice(items, func(item *model.Card) CardData {
			return *cardData(item)
		}),
	}, nil
}

type IDRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (s *Service) GetCard(ctx context.Context, req *IDRequest) (*CardData, error) {
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

func (s *Service) UpdateCardStatus(ctx context.Context, req *UpdateCardRequest) (*CardData, error) {
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

func (s *Service) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*SimulateAuthorizationData, error) {
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
	ID                   string `json:"id"`
	CardID               string `json:"card_id"`
	Status               string `json:"status"`
	AuthorizedAmount     string `json:"authorized_amount"`
	Currency             string `json:"currency"`
	MerchantName         string `json:"merchant_name"`
	MerchantCategoryCode string `json:"merchant_category_code"`
	AuthorizationCode    string `json:"authorization_code"`
	AuthorizedAt         string `json:"authorized_at"`
	CreatedAt            string `json:"created_at"`
}

func (s *Service) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) (*ListResponse[AuthorizationData], error) {
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
	return &ListResponse[AuthorizationData]{
		TotalItems: total,
		Data: types.BulkConvertSlice(items, func(item *model.Authorization) AuthorizationData {
			return *authorizationData(item)
		}),
	}, nil
}

func (s *Service) GetAuthorization(ctx context.Context, req *IDRequest) (*AuthorizationData, error) {
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
	ID                   string `json:"id"`
	CardID               string `json:"card_id"`
	AuthorizationID      string `json:"authorization_id"`
	TransactionType      string `json:"transaction_type"`
	Status               string `json:"status"`
	Amount               string `json:"amount"`
	Currency             string `json:"currency"`
	MerchantName         string `json:"merchant_name"`
	MerchantCountry      string `json:"merchant_country"`
	MerchantCategoryCode string `json:"merchant_category_code"`
	AuthorizationCode    string `json:"authorization_code"`
	TransactedAt         string `json:"transacted_at"`
	CreatedAt            string `json:"created_at"`
}

func (s *Service) ListTransactions(ctx context.Context, req *ListTransactionsRequest) (*ListResponse[TransactionData], error) {
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
	return &ListResponse[TransactionData]{
		TotalItems: total,
		Data: types.BulkConvertSlice(items, func(item *model.CardTransaction) TransactionData {
			return *transactionData(item)
		}),
	}, nil
}

func (s *Service) GetTransaction(ctx context.Context, req *IDRequest) (*TransactionData, error) {
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

func (s *Service) ClearTransaction(ctx context.Context, req *ApplyTransactionStepRequest) (*TransactionData, error) {
	return s.applyTransactionStep(ctx, req, enums.CardTransactionType_CLEAR)
}

func (s *Service) ReverseTransaction(ctx context.Context, req *ApplyTransactionStepRequest) (*TransactionData, error) {
	return s.applyTransactionStep(ctx, req, enums.CardTransactionType_VOID)
}

func (s *Service) RefundTransaction(ctx context.Context, req *ApplyTransactionStepRequest) (*TransactionData, error) {
	return s.applyTransactionStep(ctx, req, enums.CardTransactionType_REFUND)
}

func (s *Service) applyTransactionStep(ctx context.Context, req *ApplyTransactionStepRequest, transactionType enums.CardTransactionType) (*TransactionData, error) {
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
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func cardData(item *model.Card) *CardData {
	expireTime, _ := time.Parse("01/06", item.ExpireTime)
	return &CardData{
		ID:           item.ID,
		CardHolderID: item.CardHolderID,
		CardNumber:   item.CardNumber,
		Last4:        last4(item.CardNumber),
		CardBin:      item.CardBin,
		CardScheme:   item.CardScheme,
		CardCurrency: string(item.CardCurrency),
		FormFactor:   string(item.FormType),
		CardStatus:   string(item.Status),
		Status:       string(item.Status),
		ExpiryMonth:  expireTime.Format("01"),
		ExpiryYear:   expireTime.Format("2006"),
		Cvv:          item.Cvv,
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    item.UpdatedAt.UTC().Format(time.RFC3339),
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
		AuthorizedAt:         item.OccurredAt.UTC().Format(time.RFC3339),
		CreatedAt:            item.CreatedAt.UTC().Format(time.RFC3339),
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
		TransactedAt:         item.OccurredAt.UTC().Format(time.RFC3339),
		CreatedAt:            item.CreatedAt.UTC().Format(time.RFC3339),
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
