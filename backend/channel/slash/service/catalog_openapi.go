package service

import (
	"context"

	"generic-mock/channel/slash/pkg/idconv"
)

type OpenAPIMerchant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"` // Invalid: merchant directory URLs are not persisted.
}

type OpenAPIMerchantCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (service *SlashOpenAPIService) GetMerchant(ctx context.Context, req *OpenAPIAccountPathRequest) (*OpenAPIMerchant, error) {
	item, err := service.protocolAccount(ctx, req)
	if err != nil {
		return nil, err
	}
	return &OpenAPIMerchant{
		ID:   idconv.ToUUID(item.ID),
		Name: item.Name,
	}, nil
}

func (service *SlashOpenAPIService) ListMerchants(ctx context.Context, req *OpenAPIAccountPathRequest) (*OpenAPIItems[*OpenAPIMerchant], error) {
	item, err := service.GetMerchant(ctx, req)
	if err != nil {
		return nil, err
	}
	return &OpenAPIItems[*OpenAPIMerchant]{
		Items:    []*OpenAPIMerchant{item},
		Metadata: OpenAPIMetadata{Count: 1},
	}, nil
}

func (service *SlashOpenAPIService) ListMerchantCategories(ctx context.Context, req *OpenAPIAccountPathRequest) (*OpenAPIItems[*OpenAPIMerchantCategory], error) {
	if _, err := service.protocolAccount(ctx, req); err != nil {
		return nil, err
	}
	return &OpenAPIItems[*OpenAPIMerchantCategory]{Items: []*OpenAPIMerchantCategory{}}, nil
}

type OpenAPITransactionAggregation struct {
	Count     int   `json:"count"`
	TotalIn   int64 `json:"totalIn"`
	TotalOut  int64 `json:"totalOut"`
	NetChange int64 `json:"netChange"`
}

func (service *SlashOpenAPIService) GetTransactionAggregations(ctx context.Context, req *OpenAPIListTransactionsRequest) (*OpenAPITransactionAggregation, error) {
	items, err := service.ListTransactions(ctx, req)
	if err != nil {
		return nil, err
	}
	return &OpenAPITransactionAggregation{Count: len(items.Items)}, nil
}

type OpenAPIFee struct {
	ID             string `json:"id"`
	DateCharged    string `json:"dateCharged"`
	FeeAmountCents int64  `json:"feeAmountCents"`
}

func (service *SlashOpenAPIService) GetTransactionFees(ctx context.Context, req *OpenAPIIDRequest) (*OpenAPIItems[*OpenAPIFee], error) {
	if _, err := service.GetTransaction(ctx, req); err != nil {
		return nil, err
	}
	return &OpenAPIItems[*OpenAPIFee]{Items: []*OpenAPIFee{}}, nil
}
