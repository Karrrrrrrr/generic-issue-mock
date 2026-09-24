package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	ping "generic-mock/channel/pingpong/enums"
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
)

type CardFundingRequest struct {
	OpenAPIRequest
	CardID        string             `json:"card_id" binding:"required"`
	Action        ping.FundingAction `json:"action" binding:"required,oneof=top_up withdraw"`
	Amount        Number             `json:"amount"`
	UniqueOrderID string             `json:"unique_order_id" binding:"required"`
}

type BudgetFundingRequest struct {
	OpenAPIRequest
	BudgetID       string             `json:"budget_id" binding:"required"`
	Action         ping.FundingAction `json:"action" binding:"required,oneof=top_up transfer"`
	Amount         Number             `json:"amount"`
	Currency       common.Currency    `json:"currency" binding:"required,oneof=USD"`
	UniqueOrderID  *string            `json:"unique_order_id" binding:"omitempty,min=1,max=36"`
	TargetBudgetID *string            `json:"target_budget_id"`
	TargetCurrency *common.Currency   `json:"target_currency" binding:"omitempty,oneof=USD"`
}

type RecordData struct {
	RecordID string `json:"record_id"`
}

type CardFundingData struct {
	RecordID string             `json:"record_id"`
	Action   ping.FundingAction `json:"action"`
	Amount   Number             `json:"amount"`
	Currency common.Currency    `json:"currency"`
}

type BudgetOrderRequest struct {
	OpenAPIRequest
	OrderID string             `form:"order_id" binding:"required"`
	Action  ping.FundingAction `form:"action" binding:"required,oneof=top_up transfer"`
}

type BudgetOrderData struct {
	OrderID string             `json:"order_id"`
	Action  ping.FundingAction `json:"action"`
	Status  ping.FundingStatus `json:"status"`
}

type CardOrdersRequest struct {
	OpenAPIRequest
	PageRequest
	CardID    *string             `form:"card_id"`
	RequestID *string             `form:"unique_order_id" binding:"omitempty,min=1"`
	Status    *ping.FundingStatus `form:"status" binding:"omitempty,oneof=SUCCESS FAIL PROCESSING"`
	StartDate *time.Time          `form:"start_date" time_format:"2006-01-02" time_utc:"true"`
	EndDate   *time.Time          `form:"end_date" time_format:"2006-01-02" time_utc:"true"`
}

type CardOrderData struct {
	RecordID      string             `json:"record_id"`
	UniqueOrderID string             `json:"unique_order_id"`
	Created       time.Time          `json:"created"`
	Amount        Number             `json:"amount"`
	Currency      common.Currency    `json:"currency"`
	CardNumber    string             `json:"card_number"`
	CardID        string             `json:"card_id"`
	Remark        string             `json:"remark"` // Invalid: 未实现订单备注。
	Status        ping.FundingStatus `json:"status"`
	Action        ping.FundingAction `json:"action"`
}

type CardOrdersData struct {
	TotalNum int64           `json:"total_num"`
	PageNo   int             `json:"page_no"`
	PageSize int             `json:"page_size"`
	List     []CardOrderData `json:"list"`
}

func (s *PingPongOpenAPIService) CardFunding(ctx context.Context, req *CardFundingRequest) (*CardFundingData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.uc.FundCard(ctx, &biz.CardFundingRequest{
		AccountID: accountID,
		CardID:    cardID,
		Withdraw:  req.Action == ping.Withdraw,
		Amount:    req.Amount.Decimal,
		RequestID: req.UniqueOrderID,
	})
	if err != nil {
		return nil, err
	}
	return &CardFundingData{
		RecordID: idconv.ToString(item.ID),
		Action:   req.Action,
		Amount:   Number{item.Amount},
		Currency: item.Currency,
	}, nil
}

func (s *PingPongOpenAPIService) BudgetFunding(ctx context.Context, req *BudgetFundingRequest) (*RecordData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	virtualAccountID, err := idconv.FromString(req.BudgetID)
	if err != nil {
		return nil, err
	}
	targetVirtualAccountID, err := idconv.FromOptionalString(req.TargetBudgetID)
	if err != nil {
		return nil, err
	}
	if (req.Action == ping.Transfer) != (targetVirtualAccountID != nil) || (req.TargetCurrency != nil && *req.TargetCurrency != req.Currency) {
		return nil, pingerrors.ErrInvalid
	}
	item, err := s.uc.FundVirtualAccount(ctx, &biz.VirtualAccountFundingRequest{
		AccountID:              accountID,
		VirtualAccountID:       virtualAccountID,
		TargetVirtualAccountID: targetVirtualAccountID,
		Currency:               req.Currency,
		Amount:                 req.Amount.Decimal,
		RequestID:              req.UniqueOrderID,
	})
	if err != nil {
		return nil, err
	}
	return &RecordData{RecordID: idconv.ToString(item.ID)}, nil
}

func (s *PingPongOpenAPIService) BudgetOrder(ctx context.Context, req *BudgetOrderRequest) (*BudgetOrderData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	orderID, err := idconv.FromString(req.OrderID)
	if err != nil {
		return nil, err
	}
	kind := common.WalletTransfer_VirtualAccountTopUp
	if req.Action == ping.Transfer {
		kind = common.WalletTransfer_VirtualAccountTransfer
	}
	item, err := s.uc.VirtualAccountOrder(ctx, &biz.VirtualAccountOrderRequest{
		AccountID: accountID,
		ID:        orderID,
		Kind:      kind,
	})
	if err != nil {
		return nil, err
	}
	return &BudgetOrderData{
		OrderID: idconv.ToString(item.ID),
		Action:  req.Action,
		Status:  ping.FundingSuccess,
	}, nil
}

func (s *PingPongOpenAPIService) CardOrders(ctx context.Context, req *CardOrdersRequest) (*CardOrdersData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
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
	if (req.StartDate != nil && req.StartDate.IsZero()) || (req.EndDate != nil && req.EndDate.IsZero()) || (req.StartDate != nil && req.EndDate != nil && req.StartDate.After(*req.EndDate)) {
		return nil, pingerrors.ErrInvalid
	}
	var end *time.Time
	if req.EndDate != nil {
		inclusive := req.EndDate.AddDate(0, 0, 1).Add(-time.Nanosecond)
		end = &inclusive
	}
	items, total, err := s.uc.CardOrders(ctx, &biz.CardOrdersRequest{
		AccountID:      accountID,
		CardID:         cardID,
		RequestID:      req.RequestID,
		From:           req.StartDate,
		To:             end,
		Offset:         (page - 1) * limit,
		Limit:          limit,
		SuccessfulOnly: req.Status == nil || *req.Status == ping.FundingSuccess,
	})
	if err != nil {
		return nil, err
	}
	result := &CardOrdersData{
		TotalNum: total,
		PageNo:   page,
		PageSize: limit,
		List:     make([]CardOrderData, 0, len(items)),
	}
	for _, item := range items {
		if item.CardID == nil || item.Card == nil {
			return nil, pingerrors.ErrInvalid
		}
		action := ping.TopUp
		if item.Kind == common.WalletTransfer_CardWithdraw {
			action = ping.Withdraw
		}
		result.List = append(result.List, CardOrderData{
			RecordID:      idconv.ToString(item.ID),
			UniqueOrderID: item.RequestID,
			Created:       item.CreatedAt,
			Amount:        Number{item.Amount},
			Currency:      item.Currency,
			CardNumber:    item.Card.CardNumber,
			CardID:        idconv.ToString(*item.CardID),
			Status:        ping.FundingSuccess,
			Action:        action,
		})
	}
	return result, nil
}
