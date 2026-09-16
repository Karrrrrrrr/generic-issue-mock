package service

import (
	"context"
	"time"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
)

type UIListRequest struct {
	PageNumber int `form:"page_number"`
	PageSize   int `form:"page_size"`
}

type UIListResponse[T any] struct {
	TotalItems int `json:"total_items"`
	Data       []T `json:"data"`
}

type UICardHolderRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Mobile    string `json:"phone_number" binding:"required"`
}

type UICardHolderData struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Mobile    string    `json:"phone_number"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type UIService struct {
	usecase *biz.UIUsecase
}

func NewUIService(injector *do.Injector) (*UIService, error) {
	return &UIService{
		usecase: do.MustInvokeNamed[*biz.UIUsecase](injector, "photonpay.ui-usecase"),
	}, nil
}

func (s *UIService) UICreateCardHolder(ctx context.Context, req *UICardHolderRequest) (*UICardHolderData, error) {
	holder, err := s.usecase.CreateCardHolder(ctx, &biz.UICreateCardHolderRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Mobile:    req.Mobile,
	})
	if err != nil {
		return nil, err
	}

	return uiCardHolderData(holder), nil
}

func (s *UIService) UIListCardHolders(ctx context.Context, req *UIListRequest) (*UIListResponse[UICardHolderData], error) {
	holders, err := s.usecase.ListCardHolders(ctx, uiListRequest(req))
	if err != nil {
		return nil, err
	}

	return &UIListResponse[UICardHolderData]{
		TotalItems: len(holders),
		Data: types.BulkConvertSlice(holders, func(item *model.CardHolder) UICardHolderData {
			return *uiCardHolderData(item)
		}),
	}, nil
}

type UICreateCardRequest struct {
	CardHolderID string `json:"cardholder_id" binding:"required"`
	CardCurrency string `json:"card_currency" binding:"required"`
	RequestID    string `json:"request_id" binding:"required"`
}

type UICardData struct {
	ID           string    `json:"id"`
	CardHolderID string    `json:"cardholder_id"`
	CardNumber   string    `json:"card_number"`
	CardBin      string    `json:"card_bin"`
	CardCurrency string    `json:"card_currency"`
	CardStatus   string    `json:"card_status"`
	Cvv          string    `json:"cvv"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func (s *UIService) UICreateCard(ctx context.Context, req *UICreateCardRequest) (*UICardData, error) {
	card, err := s.usecase.OpenCard(ctx, &biz.UIOpenCardRequest{
		CardHolderID: model.ID(req.CardHolderID),
		Currency:     enums.Currency(req.CardCurrency),
		RequestID:    req.RequestID,
	})
	if err != nil {
		return nil, err
	}

	return uiCardData(card), nil
}

func (s *UIService) UIListCards(ctx context.Context, req *UIListRequest) (*UIListResponse[UICardData], error) {
	cards, err := s.usecase.ListCards(ctx, uiListRequest(req))
	if err != nil {
		return nil, err
	}

	return &UIListResponse[UICardData]{
		TotalItems: len(cards),
		Data: types.BulkConvertSlice(cards, func(item *model.Card) UICardData {
			return *uiCardData(item)
		}),
	}, nil
}

type UIUpdateCardStatusRequest struct {
	ID         string `uri:"id" binding:"required"`
	CardStatus string `json:"card_status" binding:"required"`
}

func (s *UIService) UIUpdateCardStatus(ctx context.Context, req *UIUpdateCardStatusRequest) (*UICardData, error) {
	card, err := s.usecase.ChangeCardStatus(ctx, &biz.UIChangeCardStatusRequest{
		CardID: model.ID(req.ID),
		Status: enums.CardStatus(req.CardStatus),
	})
	if err != nil {
		return nil, err
	}

	return uiCardData(card), nil
}

type UITransactionData struct {
	ID                   string    `json:"id"`
	CardID               string    `json:"card_id"`
	AuthorizationID      string    `json:"authorization_id"`
	TransactionType      string    `json:"transaction_type"`
	Status               string    `json:"status"`
	Amount               string    `json:"amount"`
	Currency             string    `json:"currency"`
	MerchantName         string    `json:"merchant_name"`
	MerchantCategoryCode string    `json:"merchant_category_code"`
	TransactedAt         time.Time `json:"transacted_at"`
}

func (s *UIService) UIListTransactions(ctx context.Context, req *UIListRequest) (*UIListResponse[UITransactionData], error) {
	transactions, err := s.usecase.ListTransactions(ctx, uiListRequest(req))
	if err != nil {
		return nil, err
	}

	return &UIListResponse[UITransactionData]{
		TotalItems: len(transactions),
		Data: types.BulkConvertSlice(transactions, func(item *model.CardTransaction) UITransactionData {
			return UITransactionData{
				ID:                   item.ID,
				CardID:               item.CardID,
				AuthorizationID:      item.AuthorizationID,
				TransactionType:      string(item.Type),
				Status:               string(item.Status),
				Amount:               item.TxAmount.String(),
				Currency:             string(item.TxCurrency),
				MerchantName:         item.MerchantName,
				MerchantCategoryCode: item.MerchantMCC,
				TransactedAt:         item.OccurredAt,
			}
		}),
	}, nil
}

func uiListRequest(req *UIListRequest) *biz.ListRequest {
	page, size := types.NormalizePagination(req.PageNumber, req.PageSize)

	return &biz.ListRequest{
		Offset: (page - 1) * size,
		Limit:  size,
	}
}

func uiCardHolderData(item *model.CardHolder) *UICardHolderData {
	return &UICardHolderData{
		ID:        item.ID,
		FirstName: item.FirstName,
		LastName:  item.LastName,
		Email:     item.Email,
		Mobile:    item.Mobile,
		Status:    string(item.Status),
		CreatedAt: item.CreatedAt,
	}
}

func uiCardData(item *model.Card) *UICardData {
	return &UICardData{
		ID:           item.ID,
		CardHolderID: item.CardHolderID,
		CardNumber:   item.CardNumber,
		CardBin:      item.CardBin,
		CardCurrency: string(item.CardCurrency),
		CardStatus:   string(item.Status),
		Cvv:          item.Cvv,
		ExpiresAt:    item.ExpireAt,
		CreatedAt:    item.CreatedAt,
	}
}
