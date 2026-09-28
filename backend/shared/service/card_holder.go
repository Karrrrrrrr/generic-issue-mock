package service

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"
)

type ListCardHoldersRequest struct {
	PageRequest
	ID        *model.ID `form:"id" binding:"omitempty,gt=0"`
	AccountID *model.ID `form:"account_id" binding:"omitempty,gt=0"`
}

type CardHolderData struct {
	ID           model.ID               `json:"id"`
	AccountID    model.ID               `json:"account_id"`
	AccountName  string                 `json:"account_name"`
	FirstName    string                 `json:"first_name"`
	LastName     string                 `json:"last_name"`
	Email        string                 `json:"email"`
	Mobile       string                 `json:"mobile"`
	MobilePrefix string                 `json:"mobile_prefix"`
	Status       enums.CardHolderStatus `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
}

func (s *Service) ListCardHolders(ctx context.Context, req *ListCardHoldersRequest) (*Page[CardHolderData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListCardHolders(ctx, &biz.ListUICardHoldersRequest{
		UIPageRequest: page,
		ID:            req.ID,
		AccountID:     req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	result := &Page[CardHolderData]{
		Items: make([]CardHolderData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toCardHolderData(item))
	}
	return result, nil
}

func toCardHolderData(item *model.CardHolder) CardHolderData {
	return CardHolderData{
		ID:           item.ID,
		AccountID:    item.AccountID,
		AccountName:  item.Account.GetName(),
		FirstName:    item.FirstName,
		LastName:     item.LastName,
		Email:        item.Email,
		Mobile:       item.Mobile,
		MobilePrefix: item.MobilePrefix,
		Status:       item.Status,
		CreatedAt:    item.CreatedAt,
	}
}
