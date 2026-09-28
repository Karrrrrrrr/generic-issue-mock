package service

import (
	"context"

	"generic-mock/channel/pingpong/pkg/idconv"
)

type UIProductData struct {
	ID     string `json:"id"`
	Prefix string `json:"prefix"`
}

func (s *PingPongUIService) ListCardProducts(ctx context.Context, req *Empty) (*UIPage[UIProductData], error) {
	items, err := s.uc.Products(ctx)
	if err != nil {
		return nil, err
	}
	result := &UIPage[UIProductData]{
		Items: make([]UIProductData, 0, len(items)),
		Total: int64(len(items)),
	}
	for _, item := range items {
		result.Items = append(result.Items, UIProductData{
			ID:     idconv.ToString(item.ID),
			Prefix: item.Prefix,
		})
	}
	return result, nil
}
