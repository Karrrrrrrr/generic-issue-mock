package service

import (
	"context"
	"encoding/json"

	"generic-mock/channel/slash/biz"
	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/channel/slash/pkg/idconv"
	"generic-mock/model"

	"github.com/shopspring/decimal"
)

type OpenAPIVirtualAccountMutationRequest struct {
	OpenAPIAccountRequest
	CommissionDetails *json.RawMessage `json:"commissionDetails"` // Invalid: commission schedules are not simulated.
	ID                *string          `uri:"id"`
	AccountID         *string          `json:"accountId"`
	Name              string           `json:"name" binding:"required"`
}

func virtualAccountData(item *model.VirtualAccount) *OpenAPIVirtualAccountData {
	result := &OpenAPIVirtualAccountData{
		VirtualAccount: OpenAPIVirtualAccountDetails{
			ID:          idconv.ToUUID(item.ID),
			Name:        item.Name,
			AccountType: "primary",
		},
	}
	if item.Wallet != nil {
		result.Balance.AmountCents = item.Wallet.Available.Mul(decimal.NewFromInt(100)).IntPart()
		result.Spend.AmountCents = item.Wallet.Out.Mul(decimal.NewFromInt(100)).IntPart()
	}
	return result
}

func (service *SlashOpenAPIService) GetVirtualAccount(ctx context.Context, req *OpenAPIIDRequest) (*OpenAPIVirtualAccountData, error) {
	accountID, err := idconv.FromAccountUUID(req.APIKey)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromUUID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := service.usecase.GetVirtualAccount(ctx, &biz.ResourceRequest{
		AccountID: &accountID,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}
	return virtualAccountData(item), nil
}

func (service *SlashOpenAPIService) CreateVirtualAccount(ctx context.Context, req *OpenAPIVirtualAccountMutationRequest) (*OpenAPIVirtualAccountData, error) {
	accountID, err := idconv.FromAccountUUID(req.APIKey)
	if err != nil {
		return nil, err
	}
	if req.AccountID != nil {
		selected, err := idconv.FromUUID(*req.AccountID)
		if err != nil {
			return nil, err
		}
		if selected != accountID {
			return nil, slasherrors.ErrInvalidOperation
		}
	}
	item, err := service.usecase.CreateVirtualAccount(ctx, &biz.OpenAPICreateVirtualAccountRequest{
		AccountID: accountID,
		Name:      req.Name,
	})
	if err != nil {
		return nil, err
	}
	return virtualAccountData(item), nil
}

func (service *SlashOpenAPIService) UpdateVirtualAccount(ctx context.Context, req *OpenAPIVirtualAccountMutationRequest) (*OpenAPIVirtualAccountData, error) {
	accountID, err := idconv.FromAccountUUID(req.APIKey)
	if err != nil {
		return nil, err
	}
	if req.ID == nil {
		return nil, slasherrors.ErrInvalidOperation
	}
	id, err := idconv.FromUUID(*req.ID)
	if err != nil {
		return nil, err
	}
	item, err := service.usecase.UpdateVirtualAccount(ctx, &biz.OpenAPIUpdateVirtualAccountRequest{
		AccountID: accountID,
		ID:        id,
		Name:      req.Name,
	})
	if err != nil {
		return nil, err
	}
	return virtualAccountData(item), nil
}
