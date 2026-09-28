package service

import (
	"context"

	"generic-mock/model"
)

type ListUICardProductsRequest struct{}

type UICardProductData struct {
	ID     model.ID `json:"id"`
	Prefix string   `json:"prefix"`
}

type UICardProductsData struct {
	Items []UICardProductData `json:"items"`
}

func (s *PhotonPayUIService) ListCardProducts(ctx context.Context, _ *ListUICardProductsRequest) (*UICardProductsData, error) {
	products, err := s.usecase.ListCardProducts(ctx)
	if err != nil {
		return nil, err
	}
	result := &UICardProductsData{
		Items: make([]UICardProductData, 0, len(products)),
	}
	for _, product := range products {
		result.Items = append(result.Items, UICardProductData{
			ID:     product.ID,
			Prefix: product.Prefix,
		})
	}
	return result, nil
}
