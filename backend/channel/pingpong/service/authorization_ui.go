package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	ping "generic-mock/channel/pingpong/enums"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
	"generic-mock/model"
)

type SimulateAuthorizationRequest struct {
	CardID       string          `json:"card_id" binding:"required"`
	Amount       Number          `json:"amount"`
	Currency     common.Currency `json:"currency" binding:"required,oneof=USD"`
	RequestID    string          `json:"request_id" binding:"required"`
	MerchantName string          `json:"merchant_name" binding:"required"`
}

type AuthorizationData struct {
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
	CardID *string `form:"card_id"`
}

type UIStageRequest struct {
	ID        string     `uri:"id" binding:"required"`
	Stage     ping.Stage `json:"stage" binding:"required,oneof=clear reverse refund"`
	Amount    Number     `json:"amount"`
	RequestID string     `json:"request_id" binding:"required"`
}

func toAuthorizationData(item *model.Authorization) AuthorizationData {
	return AuthorizationData{
		ID:                 idconv.ToString(item.ID),
		AccountID:          idconv.ToString(item.AccountID),
		AccountName:        item.Account.GetName(),
		CardID:             idconv.ToString(item.CardID),
		Amount:             Number{item.Amount},
		Remaining:          Number{biz.CalculateRemainingAuthorization(item)},
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
		CardID:       cardID,
		Amount:       req.Amount.Decimal,
		Currency:     req.Currency,
		RequestID:    req.RequestID,
		MerchantName: &req.MerchantName,
	})
	if err != nil {
		return nil, err
	}
	result := toAuthorizationData(item)
	return &result, nil
}

func (s *PingPongUIService) ListAuthorizations(ctx context.Context, req *UIListAuthorizationsRequest) (*UIPage[AuthorizationData], error) {
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
		AccountID: accountID,
		CardID:    cardID,
		Offset:    (page - 1) * limit,
		Limit:     limit,
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
	kind := common.CardTransactionType_CLEAR
	if req.Stage == ping.Reverse {
		kind = common.CardTransactionType_VOID
	}
	if req.Stage == ping.Refund {
		kind = common.CardTransactionType_REFUND
	}
	if err := s.uc.Stage(ctx, &biz.AuthorizationStageRequest{
		ID:        id,
		Kind:      kind,
		Amount:    req.Amount.Decimal,
		RequestID: req.RequestID,
	}); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}
