package service

import (
	"context"

	ping "generic-mock/channel/pingpong/enums"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
)

type ProductData struct {
	CardNetwork        ping.Network    `json:"card_network"`
	CardProductCode    string          `json:"card_product_code"`
	ValidMonth         int             `json:"valid_month"`
	AllowWithdrawal    bool            `json:"allow_withdrawal"`
	CardApplicationFee string          `json:"card_application_fee"`
	BillingCurrency    common.Currency `json:"billing_currency"`
	BinRange           string          `json:"bin_range"`
	Share              bool            `json:"share"`
	NameCN             string          `json:"name_cn"`
	NameEN             string          `json:"name_en"`
}

type ProductsData struct {
	ProductList []ProductData `json:"product_list"`
}

func (s *PingPongOpenAPIService) Products(ctx context.Context, req *OpenAPIRequest) (*ProductsData, error) {
	if _, err := s.resolveAccountID(ctx, req); err != nil {
		return nil, err
	}
	items, err := s.uc.Products(ctx)
	if err != nil {
		return nil, err
	}
	result := &ProductsData{ProductList: make([]ProductData, 0, len(items))}
	for _, item := range items {
		result.ProductList = append(result.ProductList, ProductData{
			CardNetwork:        ping.Visa,
			CardProductCode:    idconv.ToString(item.ID),
			ValidMonth:         24,
			AllowWithdrawal:    true,
			CardApplicationFee: "0",
			BillingCurrency:    common.Currency_USD,
			BinRange:           item.Prefix,
			Share:              false,
			NameCN:             "预算组独立卡",
			NameEN:             "Budget-funded card",
		})
	}
	return result, nil
}
