package service

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"generic-mock/channel/pingpong/biz"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
	"generic-mock/model"
)

type SimulateAuthorizationRequest struct {
	CardID          string          `json:"card_id" binding:"required"`
	Amount          Number          `json:"amount"`
	Currency        common.Currency `json:"currency" binding:"required,oneof=USD"`
	RequestID       string          `json:"request_id" binding:"required"`
	MerchantName    string          `json:"merchant_name" binding:"required"`
	MerchantCountry *string         `json:"merchant_country" binding:"omitempty,min=1"`
	MerchantMCC     *string         `json:"merchant_category_code" binding:"omitempty,min=1"`
}

type AuthorizationData struct {
	Settled            Number                       `json:"settled"`
	Reversed           Number                       `json:"reversed"`
	Refunded           Number                       `json:"refunded"`
	ID                 string                       `json:"id"`
	AccountID          string                       `json:"account_id"`
	AccountName        string                       `json:"account_name"`
	CardID             string                       `json:"card_id"`
	Amount             Number                       `json:"amount"`
	Remaining          Number                       `json:"remaining"`
	Currency           common.Currency              `json:"currency"`
	MerchantName       string                       `json:"merchant_name"`
	Status             common.CardTransactionStatus `json:"status"`
	NotificationStatus common.NotificationStatus    `json:"notification_status"`
	CreatedAt          time.Time                    `json:"created_at"`
}

type UIListAuthorizationsRequest struct {
	UIListRequest
	UIListTimeRange
	ID           *string                       `form:"id"`
	Status       *common.CardTransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
	MerchantName *string                       `form:"merchant_name" binding:"omitempty,min=1"`
	CardID       *string                       `form:"card_id"`
}

type UIStageRequest struct {
	ID        string                     `uri:"id" binding:"required"`
	Stage     common.CardTransactionType `json:"stage" binding:"required,oneof=clear void refund"`
	Amount    Number                     `json:"amount"`
	RequestID string                     `json:"request_id" binding:"required"`
}

func toAuthorizationData(item *model.Authorization) AuthorizationData {
	amounts := biz.CalculateAuthorizationAmounts(item)
	return AuthorizationData{
		Settled:            Number{amounts.Settled},
		Reversed:           Number{amounts.Reversed},
		Refunded:           Number{amounts.Refunded},
		ID:                 idconv.ToString(item.ID),
		AccountID:          idconv.ToString(item.AccountID),
		AccountName:        item.Account.GetName(),
		CardID:             idconv.ToString(item.CardID),
		Amount:             Number{item.Amount},
		Remaining:          Number{amounts.Remaining},
		Currency:           item.Currency,
		MerchantName:       item.MerchantName,
		Status:             item.Status,
		NotificationStatus: common.NotificationStatus_ContractPending,
		CreatedAt:          item.CreatedAt,
	}
}

func (s *PingPongUIService) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*AuthorizationData, error) {
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.uc.SimulateAuthorization(ctx, &biz.SimulateAuthorizationRequest{
		CardID:          cardID,
		Amount:          req.Amount.Decimal,
		Currency:        req.Currency,
		RequestID:       req.RequestID,
		MerchantName:    &req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
	})
	if err != nil {
		return nil, err
	}
	result := toAuthorizationData(item)
	return &result, nil
}

func (s *PingPongUIService) ListAuthorizations(ctx context.Context, req *UIListAuthorizationsRequest) (*UIPage[AuthorizationData], error) {
	if err := req.UIListTimeRange.Validate(); err != nil {
		return nil, err
	}
	id, err := idconv.FromOptionalString(req.ID)
	if err != nil {
		return nil, err
	}
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromOptionalString(req.CardID)
	if err != nil {
		return nil, err
	}
	page, limit, err := req.PageRequest.resolvePagination()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListAuthorizations(ctx, &biz.ListAuthorizationsRequest{
		AccountID:    accountID,
		ID:           id,
		Status:       req.Status,
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
		CardID:       cardID,
		Offset:       (page - 1) * limit,
		Limit:        limit,
	})
	if err != nil {
		return nil, err
	}
	result := &UIPage[AuthorizationData]{
		Items: make([]AuthorizationData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toAuthorizationData(item))
	}
	return result, nil
}

func (s *PingPongUIService) ApplyAuthorizationStage(ctx context.Context, req *UIStageRequest) (*Empty, error) {
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	if err := s.uc.Stage(ctx, &biz.AuthorizationStageRequest{
		ID:        id,
		Kind:      req.Stage,
		Amount:    req.Amount.Decimal,
		RequestID: req.RequestID,
	}); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

type UIAuthorizationDetailRequest struct {
	ID        string `uri:"id" binding:"required"`
	AccountID string `form:"account_id" binding:"required"`
}

type UIAuthorizationTransactionData struct {
	ID              string                       `json:"id"`
	TransactionType common.CardTransactionType   `json:"transaction_type"`
	Status          common.CardTransactionStatus `json:"status"`
	Amount          string                       `json:"amount"`
	Currency        common.Currency              `json:"currency"`
	CreatedAt       time.Time                    `json:"created_at"`
}

type UIAuthorizationDetailData struct {
	AuthorizationData
	CardNumber        string                           `json:"card_number"`
	AuthorizationCode string                           `json:"authorization_code"`
	MerchantCountry   string                           `json:"merchant_country"`
	MerchantMCC       string                           `json:"merchant_mcc"`
	RawPayload        json.RawMessage                  `json:"raw_payload"`
	Transactions      []UIAuthorizationTransactionData `json:"transactions"`
}

func (s *PingPongUIService) GetAuthorization(ctx context.Context, req *UIAuthorizationDetailRequest) (*UIAuthorizationDetailData, error) {
	accountID, err := idconv.FromString(req.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	detail, err := s.uc.GetAuthorization(ctx, &biz.UIAuthorizationDetailRequest{
		AccountID: accountID,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}
	item := detail.Authorization
	result := &UIAuthorizationDetailData{
		AuthorizationData: toAuthorizationData(item),
		CardNumber:        detail.Card.CardNumber,
		AuthorizationCode: item.AuthorizationCode,
		MerchantCountry:   item.MerchantCountry,
		MerchantMCC:       item.MerchantMCC,
		RawPayload:        json.RawMessage(item.RawPayload),
		Transactions:      make([]UIAuthorizationTransactionData, 0, len(item.CardTransactions)),
	}
	sort.Slice(item.CardTransactions, func(left, right int) bool {
		return item.CardTransactions[left].ID > item.CardTransactions[right].ID
	})
	for _, transaction := range item.CardTransactions {
		result.Transactions = append(result.Transactions, UIAuthorizationTransactionData{
			ID:              idconv.ToString(transaction.ID),
			TransactionType: transaction.Type,
			Status:          transaction.Status,
			Amount:          transaction.TxAmount.String(),
			Currency:        transaction.TxCurrency,
			CreatedAt:       transaction.CreatedAt,
		})
	}
	return result, nil
}
