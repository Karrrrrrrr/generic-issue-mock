package service

import (
	"context"

	"generic-mock/channel/paynda/pkg/idconv"
)

type ListUICardProductsRequest struct{}

type UICardProductData struct {
	ID     string `json:"id"`
	Prefix string `json:"prefix"`
}

type UICardProductsData struct {
	Items []UICardProductData `json:"items"`
}

func (s *PayndaUIService) ListCardProducts(ctx context.Context, _ *ListUICardProductsRequest) (*UICardProductsData, error) {
	products, err := s.usecase.ListCardProducts(ctx)
	if err != nil {
		return nil, err
	}
	result := &UICardProductsData{
		Items: make([]UICardProductData, 0, len(products)),
	}
	for _, product := range products {
		result.Items = append(result.Items, UICardProductData{
			ID:     idconv.ToString(product.ID),
			Prefix: product.Prefix,
		})
	}
	return result, nil
}
