package service

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"
)

type ListCardProductsRequest struct {
	PageRequest
	ID *model.ID `form:"id" binding:"omitempty,gt=0"`
}

type CardProductData struct {
	ID     model.ID `json:"id"`
	Prefix string   `json:"prefix"`
}

func (s *Service) ListCardProducts(ctx context.Context, req *ListCardProductsRequest) (*Page[CardProductData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListCardProducts(ctx, &biz.ListUICardProductsRequest{
		UIPageRequest: page,
		ID:            req.ID,
	})
	if err != nil {
		return nil, err
	}
	result := &Page[CardProductData]{
		Items: make([]CardProductData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toCardProductData(item))
	}
	return result, nil
}

func toCardProductData(item *model.CardProduct) CardProductData {
	return CardProductData{
		ID:     item.ID,
		Prefix: item.Prefix,
	}
}
