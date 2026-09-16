package service

import (
	"context"

	"generic-mock/channel/paynda/biz"
	common "generic-mock/enums"
	"generic-mock/model"

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
	Offset int `form:"offset"`
	Limit  int `form:"limit"`
}

type PayndaUICardsData struct {
	Records []*model.Card `json:"records"`
}

type PayndaUICardHoldersData struct {
	Records []*model.CardHolder `json:"records"`
}

func (s *PayndaUIService) ListCards(ctx context.Context, req *PayndaUIListRequest) (*PayndaUICardsData, error) {
	items, err := s.usecase.ListCards(ctx, &biz.PayndaListRequest{
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, err
	}

	return &PayndaUICardsData{
		Records: items,
	}, nil
}

func (s *PayndaUIService) ListCardHolders(ctx context.Context, req *PayndaUIListRequest) (*PayndaUICardHoldersData, error) {
	items, err := s.usecase.ListCardHolders(ctx, &biz.PayndaListRequest{
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, err
	}

	return &PayndaUICardHoldersData{
		Records: items,
	}, nil
}

type PayndaUISimulateAuthorizationRequest struct {
	CardID          string          `json:"cardId" binding:"required"`
	Amount          decimal.Decimal `json:"amount" binding:"required"`
	Currency        common.Currency `json:"currency" binding:"required"`
	MerchantName    string          `json:"merchantName"`
	MerchantCountry string          `json:"merchantCountry"`
	MerchantMCC     string          `json:"merchantMcc"`
}

type PayndaUISimulateAuthorizationData struct {
	Authorization   *model.Authorization   `json:"authorization"`
	CardTransaction *model.CardTransaction `json:"cardTransaction"`
}

func (s *PayndaUIService) SimulateAuthorization(
	ctx context.Context,
	req *PayndaUISimulateAuthorizationRequest,
) (*PayndaUISimulateAuthorizationData, error) {
	result, err := s.usecase.SimulateAuthorization(ctx, &biz.PayndaSimulateAuthorizationRequest{
		CardID:          req.CardID,
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
		Authorization:   result.Authorization,
		CardTransaction: result.CardTransaction,
	}, nil
}
