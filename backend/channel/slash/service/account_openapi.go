package service

import (
	"context"
	"time"

	"generic-mock/channel/slash/biz"
	slash "generic-mock/channel/slash/enums"
	"generic-mock/channel/slash/pkg/idconv"
	"generic-mock/model"

	"github.com/shopspring/decimal"
)

type OpenAPIAccountPathRequest struct {
	OpenAPIAccountRequest
	ID *string `uri:"id"`
}

type OpenAPIAccountData struct {
	ID            string              `json:"id"`
	Status        slash.AccountStatus `json:"status"`
	Name          string              `json:"name"`
	AccountNumber string              `json:"accountNumber"`
	RoutingNumber string              `json:"routingNumber"` // Invalid: bank routing is not simulated.
	CreatedAt     time.Time           `json:"createdAt"`
	Type          slash.AccountType   `json:"type"`
	Balances      []string            `json:"balances"`
}

type OpenAPIItems[T any] struct {
	Items    []T             `json:"items"`
	Metadata OpenAPIMetadata `json:"metadata"`
}

type OpenAPIBalanceData struct {
	AccountID string            `json:"accountId"`
	Type      slash.BalanceType `json:"type"`
	Available OpenAPIAmount     `json:"available"`
	Posted    OpenAPIAmount     `json:"posted"`
	Timestamp time.Time         `json:"timestamp"`
}

type OpenAPIAccountBalancesData struct {
	Balances []OpenAPIBalanceData `json:"balances"`
}

func (service *SlashOpenAPIService) protocolAccount(ctx context.Context, req *OpenAPIAccountPathRequest) (*model.Account, error) {
	accountID, err := idconv.FromAccountUUID(req.APIKey)
	if err != nil {
		return nil, err
	}
	if req.ID != nil {
		id, err := idconv.FromUUID(*req.ID)
		if err != nil {
			return nil, err
		}
		if id != accountID {
			return nil, biz.ErrResourceNotFound
		}
	}
	return service.usecase.GetAccount(ctx, accountID)
}

func (service *SlashOpenAPIService) GetAccount(ctx context.Context, req *OpenAPIAccountPathRequest) (*OpenAPIAccountData, error) {
	item, err := service.protocolAccount(ctx, req)
	if err != nil {
		return nil, err
	}
	return &OpenAPIAccountData{
		ID:            idconv.ToUUID(item.ID),
		Name:          item.Name,
		AccountNumber: idconv.ToUUID(item.ID),
		Status:        slash.AccountStatusOpen,
		Type:          slash.AccountTypeDebit,
		CreatedAt:     item.CreatedAt,
		Balances:      []string{idconv.ToUUID(item.WalletID)},
	}, nil
}

func (service *SlashOpenAPIService) ListAccounts(ctx context.Context, req *OpenAPIAccountPathRequest) (*OpenAPIItems[*OpenAPIAccountData], error) {
	item, err := service.GetAccount(ctx, req)
	if err != nil {
		return nil, err
	}
	return &OpenAPIItems[*OpenAPIAccountData]{
		Items:    []*OpenAPIAccountData{item},
		Metadata: OpenAPIMetadata{Count: 1},
	}, nil
}

func (service *SlashOpenAPIService) ListAccountBalances(ctx context.Context, req *OpenAPIAccountPathRequest) (*OpenAPIAccountBalancesData, error) {
	item, err := service.protocolAccount(ctx, req)
	if err != nil {
		return nil, err
	}
	amount := OpenAPIAmount{AmountCents: item.Wallet.Amount.Mul(decimal.NewFromInt(100)).IntPart()}
	return &OpenAPIAccountBalancesData{Balances: []OpenAPIBalanceData{{
		AccountID: idconv.ToUUID(item.ID),
		Type:      slash.BalanceTypeCash,
		Available: amount,
		Posted:    amount,
		Timestamp: time.Now().UTC(),
	}}}, nil
}

type OpenAPILegalEntity struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Structure string `json:"structure"` // Invalid: legal structures are not persisted.
}

func (service *SlashOpenAPIService) ListLegalEntities(ctx context.Context, req *OpenAPIAccountPathRequest) (*OpenAPIItems[*OpenAPILegalEntity], error) {
	item, err := service.GetAccount(ctx, req)
	if err != nil {
		return nil, err
	}
	return &OpenAPIItems[*OpenAPILegalEntity]{
		Items: []*OpenAPILegalEntity{{
			ID:   item.ID,
			Name: item.Name,
		}},
		Metadata: OpenAPIMetadata{Count: 1},
	}, nil
}
