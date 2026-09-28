package service

import (
	"context"

	"generic-mock/channel/paynda/biz"
	payndaerrors "generic-mock/channel/paynda/errors"
	"generic-mock/channel/paynda/pkg/idconv"
	"generic-mock/model"
	timeTypes "generic-mock/pkg/types/time"
)

type OpenAPIBalanceAccountRequest struct {
	AppID            string  `header:"appId" binding:"required"`
	BalanceAccountID *string `uri:"balanceAccountId"`
	Name             *string `json:"name"`
}

type OpenAPIBalanceAccountData struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	MerchantID string             `json:"merchantId"`
	CreateTime timeTypes.DateTime `json:"createTime"`
	UpdateTime timeTypes.DateTime `json:"updateTime"`
}

type OpenAPIBalanceAccountsData struct {
	Records []*OpenAPIBalanceAccountData `json:"records"`
	Total   int64                        `json:"total"`
	Current int64                        `json:"current"`
	Size    int64                        `json:"size"`
}

func balanceAccountData(item *model.Account) *OpenAPIBalanceAccountData {
	return &OpenAPIBalanceAccountData{
		ID:         idconv.ToString(item.ID),
		MerchantID: idconv.ToString(item.ID),
		Name:       item.Name,
		CreateTime: timeTypes.DateTime(item.CreatedAt.UTC()),
		UpdateTime: timeTypes.DateTime(item.UpdatedAt.UTC()),
	}
}

func (service *PayndaOpenAPIService) GetBalanceAccount(ctx context.Context, req *OpenAPIBalanceAccountRequest) (*OpenAPIBalanceAccountData, error) {
	selector := req.AppID
	if req.BalanceAccountID != nil {
		selector = *req.BalanceAccountID
	}
	id, err := idconv.FromAccountString(selector)
	if err != nil {
		return nil, err
	}
	item, err := service.usecase.GetAccountWallet(ctx, id)
	if err != nil {
		return nil, err
	}
	return balanceAccountData(item.Account), nil
}

func (service *PayndaOpenAPIService) ListBalanceAccounts(ctx context.Context, req *OpenAPIBalanceAccountRequest) (*OpenAPIBalanceAccountsData, error) {
	item, err := service.GetBalanceAccount(ctx, req)
	if err != nil {
		return nil, err
	}
	return &OpenAPIBalanceAccountsData{
		Records: []*OpenAPIBalanceAccountData{item},
		Total:   1,
		Current: 1,
		Size:    100,
	}, nil
}

func (service *PayndaOpenAPIService) CreateBalanceAccount(ctx context.Context, req *OpenAPIBalanceAccountRequest) (*OpenAPIBalanceAccountData, error) {
	if req.Name == nil || *req.Name == "" {
		return nil, payndaerrors.ErrInvalidOperation
	}
	if _, err := idconv.FromAccountString(req.AppID); err != nil {
		return nil, err
	}
	item, err := service.usecase.CreateAccount(ctx, &biz.OpenAPICreateAccountRequest{Name: *req.Name})
	if err != nil {
		return nil, err
	}
	return balanceAccountData(item), nil
}

func (service *PayndaOpenAPIService) UpdateBalanceAccount(ctx context.Context, req *OpenAPIBalanceAccountRequest) (*OpenAPIBalanceAccountData, error) {
	if req.Name == nil || req.BalanceAccountID == nil {
		return nil, payndaerrors.ErrInvalidOperation
	}
	id, err := idconv.FromAccountString(*req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	item, err := service.usecase.UpdateAccount(ctx, &biz.OpenAPIUpdateAccountRequest{
		ID:   id,
		Name: *req.Name,
	})
	if err != nil {
		return nil, err
	}
	return balanceAccountData(item), nil
}

func (service *PayndaOpenAPIService) DeleteBalanceAccount(ctx context.Context, req *OpenAPIBalanceAccountRequest) (*struct{}, error) {
	_, err := service.GetBalanceAccount(ctx, req)
	return &struct{}{}, err
}
