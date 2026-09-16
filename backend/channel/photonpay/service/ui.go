package service

import (
	"context"
	"time"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
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

type PhotonPayUIService struct {
	usecase *biz.PhotonPayUIUsecase
}

func NewPhotonPayUIService(injector *do.Injector) (*PhotonPayUIService, error) {
	return &PhotonPayUIService{
		usecase: do.MustInvoke[*biz.PhotonPayUIUsecase](injector),
	}, nil
}

func (s *PhotonPayUIService) CreateCardHolder(ctx context.Context, req *UICardHolderRequest) (*UICardHolderData, error) {
	holder, err := s.usecase.CreateCardHolder(ctx, &biz.UICreateCardHolderRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Mobile:    req.Mobile,
	})
	if err != nil {
		return nil, err
	}

	return photonPayUICardHolderData(holder), nil
}

func (s *PhotonPayUIService) ListCardHolders(ctx context.Context, req *UIListRequest) (*UIListResponse[*UICardHolderData], error) {
	holders, err := s.usecase.ListCardHolders(ctx, photonPayUIListRequest(req))
	if err != nil {
		return nil, err
	}

	return &UIListResponse[*UICardHolderData]{
		TotalItems: len(holders),
		Data:       types.BulkConvertSlice(holders, photonPayUICardHolderData),
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

func (s *PhotonPayUIService) CreateCard(ctx context.Context, req *UICreateCardRequest) (*UICardData, error) {
	card, err := s.usecase.OpenCard(ctx, &biz.UIOpenCardRequest{
		CardHolderID: model.ID(req.CardHolderID),
		Currency:     enums.Currency(req.CardCurrency),
		RequestID:    req.RequestID,
	})
	if err != nil {
		return nil, err
	}

	return photonPayUICardData(card), nil
}

func (s *PhotonPayUIService) ListCards(ctx context.Context, req *UIListRequest) (*UIListResponse[*UICardData], error) {
	cards, err := s.usecase.ListCards(ctx, photonPayUIListRequest(req))
	if err != nil {
		return nil, err
	}

	return &UIListResponse[*UICardData]{
		TotalItems: len(cards),
		Data:       types.BulkConvertSlice(cards, photonPayUICardData),
	}, nil
}

type UIUpdateCardStatusRequest struct {
	ID         string `uri:"id" binding:"required"`
	CardStatus string `json:"card_status" binding:"required"`
}

func (s *PhotonPayUIService) UpdateCardStatus(ctx context.Context, req *UIUpdateCardStatusRequest) (*UICardData, error) {
	card, err := s.usecase.ChangeCardStatus(ctx, &biz.UIChangeCardStatusRequest{
		CardID: model.ID(req.ID),
		Status: enums.CardStatus(req.CardStatus),
	})
	if err != nil {
		return nil, err
	}

	return photonPayUICardData(card), nil
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

type UIAuthorizationData struct {
	ID                   string    `json:"id"`
	CardID               string    `json:"card_id"`
	Status               string    `json:"status"`
	AuthorizedAmount     string    `json:"authorized_amount"`
	Currency             string    `json:"currency"`
	MerchantName         string    `json:"merchant_name"`
	MerchantCategoryCode string    `json:"merchant_category_code"`
	AuthorizationCode    string    `json:"authorization_code"`
	AuthorizedAt         time.Time `json:"authorized_at"`
}

type UISimulateAuthorizationRequest struct {
	CardID               string  `json:"card_id" binding:"required"`
	TransactionAmount    float64 `json:"transaction_amount" binding:"required"`
	TransactionCurrency  string  `json:"transaction_currency" binding:"required"`
	MerchantName         string  `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string  `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string  `json:"merchant_country"`
}

type UISimulateAuthorizationData struct {
	Approved      bool                 `json:"approved"`
	Status        string               `json:"status"`
	Authorization *UIAuthorizationData `json:"authorization"`
	Transaction   *UITransactionData   `json:"transaction"`
}

func (s *PhotonPayUIService) SimulateAuthorization(ctx context.Context, req *UISimulateAuthorizationRequest) (*UISimulateAuthorizationData, error) {
	result, err := s.usecase.SimulateAuthorization(ctx, &biz.UISimulateAuthorizationRequest{
		CardID:          model.ID(req.CardID),
		Amount:          decimal.NewFromFloat(req.TransactionAmount),
		Currency:        enums.Currency(req.TransactionCurrency),
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}

	return &UISimulateAuthorizationData{
		Approved:      true,
		Status:        string(result.Authorization.Status),
		Authorization: photonPayUIAuthorizationData(result.Authorization),
		Transaction:   photonPayUITransactionData(result.CardTransaction),
	}, nil
}

type UIApplyTransactionStepRequest struct {
	ID     string  `uri:"id" binding:"required"`
	Amount float64 `json:"amount"`
}

func (s *PhotonPayUIService) ClearTransaction(ctx context.Context, req *UIApplyTransactionStepRequest) (*UITransactionData, error) {
	return s.applyTransactionStep(ctx, req, enums.CardTransactionType_CLEAR)
}

func (s *PhotonPayUIService) ReverseTransaction(ctx context.Context, req *UIApplyTransactionStepRequest) (*UITransactionData, error) {
	return s.applyTransactionStep(ctx, req, enums.CardTransactionType_VOID)
}

func (s *PhotonPayUIService) RefundTransaction(ctx context.Context, req *UIApplyTransactionStepRequest) (*UITransactionData, error) {
	return s.applyTransactionStep(ctx, req, enums.CardTransactionType_REFUND)
}

func (s *PhotonPayUIService) applyTransactionStep(ctx context.Context, req *UIApplyTransactionStepRequest, transactionType enums.CardTransactionType) (*UITransactionData, error) {
	item, err := s.usecase.ApplyTransactionStep(ctx, &biz.UIApplyTransactionStepRequest{
		CardTransactionID: model.ID(req.ID),
		Type:              transactionType,
		Amount:            decimal.NewFromFloat(req.Amount),
	})
	if err != nil {
		return nil, err
	}

	return photonPayUITransactionData(item), nil
}

func (s *PhotonPayUIService) ListTransactions(ctx context.Context, req *UIListRequest) (*UIListResponse[*UITransactionData], error) {
	transactions, err := s.usecase.ListTransactions(ctx, photonPayUIListRequest(req))
	if err != nil {
		return nil, err
	}

	return &UIListResponse[*UITransactionData]{
		TotalItems: len(transactions),
		Data:       types.BulkConvertSlice(transactions, photonPayUITransactionData),
	}, nil
}

func photonPayUIListRequest(req *UIListRequest) *biz.ListRequest {
	page, size := types.NormalizePagination(req.PageNumber, req.PageSize)

	return &biz.ListRequest{
		Offset: (page - 1) * size,
		Limit:  size,
	}
}

func photonPayUICardHolderData(item *model.CardHolder) *UICardHolderData {
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

func photonPayUICardData(item *model.Card) *UICardData {
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

func photonPayUIAuthorizationData(item *model.Authorization) *UIAuthorizationData {
	return &UIAuthorizationData{
		ID:                   item.ID,
		CardID:               item.CardID,
		Status:               string(item.Status),
		AuthorizedAmount:     item.Amount.String(),
		Currency:             string(item.Currency),
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		AuthorizationCode:    item.AuthorizationCode,
		AuthorizedAt:         item.OccurredAt,
	}
}

func photonPayUITransactionData(item *model.CardTransaction) *UITransactionData {
	return &UITransactionData{
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
}
