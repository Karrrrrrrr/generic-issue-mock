package service

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"generic-mock/channel/pingpong/biz"
	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
)

type SimulateAuthorizationRequest struct {
	CardID          model.ID        `json:"card_id" binding:"required,gt=0"`
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
	ID                 model.ID                     `json:"id"`
	AccountID          model.ID                     `json:"account_id"`
	AccountName        string                       `json:"account_name"`
	CardID             model.ID                     `json:"card_id"`
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
	ID           *model.ID                     `form:"id" binding:"omitempty,gt=0"`
	Status       *common.CardTransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
	MerchantName *string                       `form:"merchant_name" binding:"omitempty,min=1"`
	CardID       *model.ID                     `form:"card_id" binding:"omitempty,gt=0"`
}

type UIStageRequest struct {
	ID        model.ID                   `uri:"id" binding:"required,gt=0"`
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
		ID:                 item.ID,
		AccountID:          item.AccountID,
		AccountName:        item.Account.GetName(),
		CardID:             item.CardID,
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
	cardID := req.CardID
	if cardID <= 0 {
		return nil, pingerrors.ErrInvalid
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
	id := req.ID
	if id != nil && *id <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	cardID := req.CardID
	if cardID != nil && *cardID <= 0 {
		return nil, pingerrors.ErrInvalid
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
	id := req.ID
	if id <= 0 {
		return nil, pingerrors.ErrInvalid
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
	ID        model.ID `uri:"id" binding:"required,gt=0"`
	AccountID model.ID `form:"account_id" binding:"required,gt=0"`
}

type UIAuthorizationTransactionData struct {
	ID              model.ID                     `json:"id"`
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
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	id := req.ID
	if id <= 0 {
		return nil, pingerrors.ErrInvalid
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
			ID:              transaction.ID,
			TransactionType: transaction.Type,
			Status:          transaction.Status,
			Amount:          transaction.TxAmount.String(),
			Currency:        transaction.TxCurrency,
			CreatedAt:       transaction.CreatedAt,
		})
	}
	return result, nil
}
