package service

import (
	"context"
	"time"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
)

type OpenAPIService struct {
	usecase *biz.OpenAPIUsecase
}

func NewOpenAPIService(injector *do.Injector) (*OpenAPIService, error) {
	return &OpenAPIService{
		usecase: do.MustInvokeNamed[*biz.OpenAPIUsecase](injector, "slash.openapi-usecase"),
	}, nil
}

type OpenAPIListRequest struct {
	PageNumber int `form:"page_number"`
	PageSize   int `form:"page_size"`
}

type OpenAPIMetadata struct {
	Count int `json:"count"`
}

type OpenAPICard struct {
	ID            string `json:"id"`
	CardProductID string `json:"card_product_id"`
	CardholderID  string `json:"cardholder_id"`
	Status        string `json:"status"`
	PAN           string `json:"pan"`
	CVV           string `json:"cvv"`
	Expiration    string `json:"expiration"`
	Currency      string `json:"currency"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type OpenAPIListCardsRequest struct {
	OpenAPIListRequest
	FilterStatus string `form:"filter_status"`
}

type OpenAPIListCardsData struct {
	Items    []OpenAPICard   `json:"items"`
	Metadata OpenAPIMetadata `json:"metadata"`
}

func (s *OpenAPIService) ListCards(ctx context.Context, req *OpenAPIListCardsRequest) (*OpenAPIListCardsData, error) {
	offset, limit := openAPIPagination(req.PageNumber, req.PageSize)
	items, err := s.usecase.ListCards(ctx, &biz.OpenAPIListCardsRequest{
		Offset: offset,
		Limit:  limit,
		Status: enums.CardStatus(req.FilterStatus),
	})
	if err != nil {
		return nil, err
	}

	return &OpenAPIListCardsData{
		Items: types.BulkConvertSlice(items, func(item *model.Card) OpenAPICard {
			return *openAPICard(item)
		}),
		Metadata: OpenAPIMetadata{
			Count: len(items),
		},
	}, nil
}

type OpenAPICreateCardRequest struct {
	CardholderID  string `json:"cardholder_id" binding:"required"`
	CardProductID string `json:"card_product_id" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
}

func (s *OpenAPIService) CreateCard(ctx context.Context, req *OpenAPICreateCardRequest) (*OpenAPICard, error) {
	item, err := s.usecase.CreateCard(ctx, &biz.OpenAPICreateCardRequest{
		CardHolderID:  model.ID(req.CardholderID),
		CardProductID: model.ID(req.CardProductID),
		Currency:      enums.Currency(req.Currency),
	})
	if err != nil {
		return nil, err
	}

	return openAPICard(item), nil
}

type OpenAPIIDRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (s *OpenAPIService) GetCard(ctx context.Context, req *OpenAPIIDRequest) (*OpenAPICard, error) {
	item, err := s.usecase.GetCard(ctx, model.ID(req.ID))
	if err != nil {
		return nil, err
	}

	return openAPICard(item), nil
}

type OpenAPIUpdateCardRequest struct {
	OpenAPIIDRequest
	Status string `json:"status" binding:"required"`
}

func (s *OpenAPIService) UpdateCard(ctx context.Context, req *OpenAPIUpdateCardRequest) (*OpenAPICard, error) {
	item, err := s.usecase.UpdateCard(ctx, &biz.OpenAPIUpdateCardRequest{
		ID:     model.ID(req.ID),
		Status: enums.CardStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	return openAPICard(item), nil
}

type OpenAPICardProduct struct {
	ID        string `json:"id"`
	Prefix    string `json:"prefix"`
	IsDefault bool   `json:"is_default"`
}

type OpenAPIListCardProductsRequest struct{}

type OpenAPIListCardProductsData struct {
	Items []OpenAPICardProduct `json:"items"`
}

func (s *OpenAPIService) ListCardProducts(ctx context.Context, _ *OpenAPIListCardProductsRequest) (*OpenAPIListCardProductsData, error) {
	items, err := s.usecase.ListCardProducts(ctx)
	if err != nil {
		return nil, err
	}

	return &OpenAPIListCardProductsData{
		Items: types.BulkConvertSlice(items, func(item *model.CardProduct) OpenAPICardProduct {
			return OpenAPICardProduct{
				ID:        item.ID,
				Prefix:    item.Prefix,
				IsDefault: item.IsDefault,
			}
		}),
	}, nil
}

type OpenAPITransaction struct {
	ID           string `json:"id"`
	CardID       string `json:"card_id"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	MerchantName string `json:"merchant_name"`
	AuthorizedAt string `json:"authorized_at"`
	CreatedAt    string `json:"created_at"`
}

type OpenAPIListTransactionsRequest struct {
	OpenAPIListRequest
	FilterCardID string `form:"filter_card_id"`
}

type OpenAPIListTransactionsData struct {
	Items    []OpenAPITransaction `json:"items"`
	Metadata OpenAPIMetadata      `json:"metadata"`
}

func (s *OpenAPIService) ListTransactions(ctx context.Context, req *OpenAPIListTransactionsRequest) (*OpenAPIListTransactionsData, error) {
	offset, limit := openAPIPagination(req.PageNumber, req.PageSize)
	items, err := s.usecase.ListTransactions(ctx, &biz.OpenAPIListTransactionsRequest{
		Offset: offset,
		Limit:  limit,
		CardID: model.ID(req.FilterCardID),
	})
	if err != nil {
		return nil, err
	}

	return &OpenAPIListTransactionsData{
		Items: types.BulkConvertSlice(items, func(item *model.CardTransaction) OpenAPITransaction {
			return *openAPITransaction(item)
		}),
		Metadata: OpenAPIMetadata{
			Count: len(items),
		},
	}, nil
}

func (s *OpenAPIService) GetTransaction(ctx context.Context, req *OpenAPIIDRequest) (*OpenAPITransaction, error) {
	item, err := s.usecase.GetTransaction(ctx, model.ID(req.ID))
	if err != nil {
		return nil, err
	}

	return openAPITransaction(item), nil
}

func openAPICard(item *model.Card) *OpenAPICard {
	return &OpenAPICard{
		ID:            item.ID,
		CardProductID: item.CardProductID,
		CardholderID:  item.CardHolderID,
		Status:        string(item.Status),
		PAN:           item.CardNumber,
		CVV:           item.Cvv,
		Expiration:    item.ExpireAt.Format("01/06"),
		Currency:      string(item.CardCurrency),
		CreatedAt:     item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func openAPITransaction(item *model.CardTransaction) *OpenAPITransaction {
	return &OpenAPITransaction{
		ID:           item.ID,
		CardID:       item.CardID,
		Status:       string(item.Status),
		Type:         string(item.Type),
		Amount:       item.TxAmount.String(),
		Currency:     string(item.TxCurrency),
		MerchantName: item.MerchantName,
		AuthorizedAt: item.OccurredAt.UTC().Format(time.RFC3339),
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func openAPIPagination(pageNumber int, pageSize int) (int, int) {
	page, size := types.NormalizePagination(pageNumber, pageSize)

	return (page - 1) * size, size
}
