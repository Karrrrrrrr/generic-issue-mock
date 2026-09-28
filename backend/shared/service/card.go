package service

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
)

type ListCardsRequest struct {
	PageRequest
	TimeRange
	ID         *model.ID         `form:"id" binding:"omitempty,gt=0"`
	AccountID  *model.ID         `form:"account_id" binding:"omitempty,gt=0"`
	Status     *enums.CardStatus `form:"status"`
	CardNumber *string           `form:"card_number"`
}

type CardData struct {
	ID               model.ID           `json:"id"`
	AccountID        model.ID           `json:"account_id"`
	AccountName      string             `json:"account_name"`
	CardProductID    model.ID           `json:"card_product_id"`
	CardHolderID     model.ID           `json:"cardholder_id"`
	VirtualAccountID *model.ID          `json:"virtual_account_id"`
	WalletID         model.ID           `json:"wallet_id"`
	CardType         enums.CardType     `json:"card_type"`
	CardNumber       string             `json:"card_number"`
	CardBin          string             `json:"card_bin"`
	CVV              string             `json:"cvv"`
	Currency         enums.Currency     `json:"currency"`
	Scheme           enums.CardScheme   `json:"scheme"`
	FormType         enums.CardFormType `json:"form_type"`
	Status           enums.CardStatus   `json:"status"`
	Available        decimal.Decimal    `json:"available"`
	PendingOut       decimal.Decimal    `json:"pending_out"`
	ExpiresAt        time.Time          `json:"expires_at"`
	CreatedAt        time.Time          `json:"created_at"`
}

func (s *Service) ListCards(ctx context.Context, req *ListCardsRequest) (*Page[CardData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListCards(ctx, &biz.ListUICardsRequest{
		UIPageRequest: page,
		ID:            req.ID,
		AccountID:     req.AccountID,
		Status:        req.Status,
		CardNumber:    req.CardNumber,
		UITimeRange: biz.UITimeRange{
			CreatedFrom: req.CreatedFrom,
			CreatedTo:   req.CreatedTo,
		},
	})
	if err != nil {
		return nil, err
	}
	result := &Page[CardData]{
		Items: make([]CardData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toCardData(item))
	}
	return result, nil
}

func toCardData(item *model.Card) CardData {
	available := decimal.Zero
	pending := decimal.Zero
	if item.Wallet != nil {
		available = item.Wallet.Available
		pending = item.Wallet.PendingOut
	}
	return CardData{
		ID:               item.ID,
		AccountID:        item.AccountID,
		AccountName:      item.Account.GetName(),
		CardProductID:    item.CardProductID,
		CardHolderID:     item.CardHolderID,
		VirtualAccountID: item.VirtualAccountID,
		WalletID:         item.WalletID,
		CardType:         item.CardType,
		CardNumber:       item.CardNumber,
		CardBin:          item.CardBin,
		CVV:              item.Cvv,
		Currency:         item.CardCurrency,
		Scheme:           item.CardScheme,
		FormType:         item.FormType,
		Status:           item.Status,
		Available:        available,
		PendingOut:       pending,
		ExpiresAt:        item.ExpireAt,
		CreatedAt:        item.CreatedAt,
	}
}

type GetCardRequest struct {
	ID        model.ID `form:"id" binding:"required,gt=0"`
	AccountID model.ID `form:"account_id" binding:"required,gt=0"`
}

func (s *Service) GetCard(ctx context.Context, req *GetCardRequest) (*CardData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.GetCard(ctx, &biz.GetUICardRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	result := toCardData(item)
	return &result, nil
}

type UpdateCardStatusRequest struct {
	ID        model.ID         `json:"id" binding:"required,gt=0"`
	AccountID model.ID         `json:"account_id" binding:"required,gt=0"`
	Status    enums.CardStatus `json:"status" binding:"required,oneof=active frozen deleted"`
}

func (s *Service) UpdateCardStatus(ctx context.Context, req *UpdateCardStatusRequest) (*CardData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.UpdateCardStatus(ctx, &biz.UpdateUICardStatusRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Status:    req.Status,
	})
	if err != nil {
		return nil, err
	}
	result := toCardData(item)
	return &result, nil
}
