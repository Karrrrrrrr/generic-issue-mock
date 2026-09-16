package service

import (
	"context"
	"time"

	"generic-mock/channel/photonpay/biz"
	photon "generic-mock/channel/photonpay/enums"
	common "generic-mock/enums"
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
	ID        string                  `json:"id"`
	FirstName string                  `json:"first_name"`
	LastName  string                  `json:"last_name"`
	Email     string                  `json:"email"`
	Mobile    string                  `json:"phone_number"`
	Status    photon.CardHolderStatus `json:"status"`
	CreatedAt time.Time               `json:"created_at"`
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
	ID           string            `json:"id"`
	CardHolderID string            `json:"cardholder_id"`
	CardNumber   string            `json:"card_number"`
	CardBin      string            `json:"card_bin"`
	CardCurrency string            `json:"card_currency"`
	CardStatus   photon.CardStatus `json:"card_status"`
	Cvv          string            `json:"cvv"`
	ExpiresAt    time.Time         `json:"expires_at"`
	CreatedAt    time.Time         `json:"created_at"`
}

func (s *PhotonPayUIService) CreateCard(ctx context.Context, req *UICreateCardRequest) (*UICardData, error) {
	cardHolderID, err := photonPayID(req.CardHolderID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.OpenCard(ctx, &biz.UIOpenCardRequest{
		CardHolderID: cardHolderID,
		Currency:     common.Currency(req.CardCurrency),
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

func (s *PhotonPayUIService) ListAuthorizations(ctx context.Context, req *UIListRequest) (*UIListResponse[*UIAuthorizationData], error) {
	items, err := s.usecase.ListAuthorizations(ctx, photonPayUIListRequest(req))
	if err != nil {
		return nil, err
	}
	return &UIListResponse[*UIAuthorizationData]{TotalItems: len(items), Data: types.BulkConvertSlice(items, photonPayUIAuthorizationData)}, nil
}

type UIUpdateCardStatusRequest struct {
	ID         string            `uri:"id" binding:"required"`
	CardStatus photon.CardStatus `json:"card_status" binding:"required"`
}

func (s *PhotonPayUIService) UpdateCardStatus(ctx context.Context, req *UIUpdateCardStatusRequest) (*UICardData, error) {
	cardID, err := photonPayID(req.ID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.ChangeCardStatus(ctx, &biz.UIChangeCardStatusRequest{
		CardID: cardID,
		Status: photon.CardStatusToGeneric(req.CardStatus),
	})
	if err != nil {
		return nil, err
	}

	return photonPayUICardData(card), nil
}

type UITransactionData struct {
	ID                   string                   `json:"id"`
	CardID               string                   `json:"card_id"`
	AuthorizationID      string                   `json:"authorization_id"`
	TransactionType      photon.TransactionType   `json:"transaction_type"`
	Status               photon.TransactionStatus `json:"status"`
	Amount               string                   `json:"amount"`
	Currency             string                   `json:"currency"`
	MerchantName         string                   `json:"merchant_name"`
	MerchantCategoryCode string                   `json:"merchant_category_code"`
	TransactedAt         time.Time                `json:"transacted_at"`
}

type UIAuthorizationData struct {
	ID                   string                     `json:"id"`
	CardID               string                     `json:"card_id"`
	Status               photon.AuthorizationStatus `json:"status"`
	AuthorizedAmount     string                     `json:"authorized_amount"`
	Currency             string                     `json:"currency"`
	MerchantName         string                     `json:"merchant_name"`
	MerchantCategoryCode string                     `json:"merchant_category_code"`
	AuthorizationCode    string                     `json:"authorization_code"`
	AuthorizedAt         time.Time                  `json:"authorized_at"`
}

type UISimulateAuthorizationRequest struct {
	CardID               string  `json:"card_id" binding:"required"`
	TransactionAmount    float64 `json:"transaction_amount" binding:"required"`
	TransactionCurrency  string  `json:"transaction_currency" binding:"required"`
	MerchantName         string  `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string  `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string  `json:"merchant_country"`
	MerchantCity         string  `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

type UISimulateAuthorizationData struct {
	Approved      bool                     `json:"approved"`
	Status        photon.TransactionStatus `json:"status"`
	Authorization *UIAuthorizationData     `json:"authorization"`
	Transaction   *UITransactionData       `json:"transaction"`
}

type UISimulateRefundRequest struct {
	CardID               string  `json:"card_id" binding:"required"`
	Amount               float64 `json:"amount" binding:"required,gt=0"`
	Currency             string  `json:"currency" binding:"required"`
	MerchantName         string  `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string  `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string  `json:"merchant_country" binding:"required"`
	MerchantCity         string  `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

func (s *PhotonPayUIService) SimulateRefund(ctx context.Context, req *UISimulateRefundRequest) (*UITransactionData, error) {
	cardID, err := photonPayID(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.SimulateRefund(ctx, &biz.UISimulateRefundRequest{
		CardID:          cardID,
		Amount:          decimal.NewFromFloat(req.Amount),
		Currency:        common.Currency(req.Currency),
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}
	return photonPayUITransactionData(item), nil
}

func (s *PhotonPayUIService) SimulateAuthorization(ctx context.Context, req *UISimulateAuthorizationRequest) (*UISimulateAuthorizationData, error) {
	cardID, err := photonPayID(req.CardID)
	if err != nil {
		return nil, err
	}
	result, err := s.usecase.SimulateAuthorization(ctx, &biz.UISimulateAuthorizationRequest{
		CardID:          cardID,
		Amount:          decimal.NewFromFloat(req.TransactionAmount),
		Currency:        common.Currency(req.TransactionCurrency),
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}

	return &UISimulateAuthorizationData{
		Approved:      true,
		Status:        photon.TransactionStatusFromGeneric(result.Authorization.Status),
		Authorization: photonPayUIAuthorizationData(result.Authorization),
		Transaction:   photonPayUITransactionData(result.CardTransaction),
	}, nil
}

type UIApplyTransactionStepRequest struct {
	ID     string  `uri:"id" binding:"required"`
	Amount float64 `json:"amount"`
}

func (s *PhotonPayUIService) ClearTransaction(ctx context.Context, req *UIApplyTransactionStepRequest) (*UITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_CLEAR)
}

func (s *PhotonPayUIService) ReverseTransaction(ctx context.Context, req *UIApplyTransactionStepRequest) (*UITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_VOID)
}

func (s *PhotonPayUIService) RefundTransaction(ctx context.Context, req *UIApplyTransactionStepRequest) (*UITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_REFUND)
}

func (s *PhotonPayUIService) applyTransactionStep(ctx context.Context, req *UIApplyTransactionStepRequest, transactionType common.CardTransactionType) (*UITransactionData, error) {
	transactionID, err := photonPayID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.ApplyTransactionStep(ctx, &biz.UIApplyTransactionStepRequest{
		CardTransactionID: transactionID,
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
		ID:        photonPayIDString(item.ID),
		FirstName: item.FirstName,
		LastName:  item.LastName,
		Email:     item.Email,
		Mobile:    item.Mobile,
		Status:    photon.CardHolderStatusFromGeneric(item.Status),
		CreatedAt: item.CreatedAt,
	}
}

func photonPayUICardData(item *model.Card) *UICardData {
	return &UICardData{
		ID:           photonPayIDString(item.ID),
		CardHolderID: photonPayIDString(item.CardHolderID),
		CardNumber:   item.CardNumber,
		CardBin:      item.CardBin,
		CardCurrency: string(item.CardCurrency),
		CardStatus:   photon.CardStatusFromGeneric(item.Status),
		Cvv:          item.Cvv,
		ExpiresAt:    item.ExpireAt,
		CreatedAt:    item.CreatedAt,
	}
}

func photonPayUIAuthorizationData(item *model.Authorization) *UIAuthorizationData {
	return &UIAuthorizationData{
		ID:                   photonPayIDString(item.ID),
		CardID:               photonPayIDString(item.CardID),
		Status:               photon.AuthorizationStatusFromGeneric(item.Status),
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
		ID:                   photonPayIDString(item.ID),
		CardID:               photonPayIDString(item.CardID),
		AuthorizationID:      photonPayIDString(item.AuthorizationID),
		TransactionType:      photon.TransactionTypeFromGeneric(item.Type),
		Status:               photon.TransactionStatusFromGeneric(item.Status),
		Amount:               item.TxAmount.String(),
		Currency:             string(item.TxCurrency),
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		TransactedAt:         item.OccurredAt,
	}
}
