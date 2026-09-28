package service

import (
	"context"
	"encoding/json"
	"time"

	"generic-mock/channel/pingpong/biz"
	ping "generic-mock/channel/pingpong/enums"
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
)

type Amount struct {
	Amount   Number          `json:"amount"`
	Currency common.Currency `json:"currency"`
}

type CreateCardRequest struct {
	OpenAPIRequest
	RequestID           string          `json:"request_id" binding:"required"`
	CardProductCode     string          `json:"card_product_code" binding:"required"`
	CardCurrency        common.Currency `json:"card_currency" binding:"required,oneof=USD"`
	BudgetID            string          `json:"budget_id" binding:"required"`
	CardholderID        *string         `json:"cardholder_id"`         // Invalid: 未实现持卡人和 3DS 资料管理。
	PerTransactionLimit *Amount         `json:"per_transaction_limit"` // Invalid: 不执行渠道限额。
	DailyLimit          *Amount         `json:"daily_limit"`           // Invalid: 不执行渠道限额。
	WeeklyLimit         *Amount         `json:"weekly_limit"`          // Invalid: 不执行渠道限额。
	MonthlyLimit        *Amount         `json:"monthly_limit"`         // Invalid: 不执行渠道限额。
	LifetimeLimit       *Amount         `json:"lifetime_limit"`        // Invalid: 不执行渠道限额。
	CouponApplied       *bool           `json:"coupon_applied"`        // Invalid: mock 不模拟优惠券计费。
	Remark              *string         `json:"remark"`                // Invalid: 不持久化卡片备注。
}

type CardIDData struct {
	CardID string `json:"card_id"`
}

type CardRequest struct {
	OpenAPIRequest
	CardID string `form:"card_id" binding:"required"`
}

type CardDetails struct {
	CardID                 string          `json:"card_id"`
	BudgetID               string          `json:"budget_id"`
	CardStatus             ping.CardStatus `json:"card_status"`
	CardType               ping.CardType   `json:"card_type"` // Invalid: SDK 未提供该字段的取值契约。
	CardNumber             string          `json:"card_number"`
	CVC                    string          `json:"cvc"`
	CardExpiryDate         string          `json:"card_expiry_date"`
	WithdrawalAllowed      bool            `json:"withdrawal_allowed"`
	CancellationInProgress bool            `json:"cancellation_in_progress"`
	Cancelled              bool            `json:"cancelled"`
	BillingCurrency        common.Currency `json:"billing_currency"`
	MaxTransactions        int             `json:"max_transactions"`      // Invalid: 未实现交易次数限制。
	PerTransactionLimit    *Amount         `json:"per_transaction_limit"` // Invalid: 未实现渠道限额。
	DailyLimit             *Amount         `json:"daily_limit"`           // Invalid: 未实现渠道限额。
	WeeklyLimit            *Amount         `json:"weekly_limit"`          // Invalid: 未实现渠道限额。
	MonthlyLimit           *Amount         `json:"monthly_limit"`         // Invalid: 未实现渠道限额。
	LifetimeLimit          *Amount         `json:"lifetime_limit"`        // Invalid: 未实现渠道限额。
	Remark                 string          `json:"remark"`                // Invalid: 未持久化卡片备注。
	CreatedAt              time.Time       `json:"created_at"`
}

type CardBalance struct {
	CardNumber       string          `json:"card_number"`
	AvailableBalance Number          `json:"available_balance"`
	Currency         common.Currency `json:"currency"`
}

type CardActionRequest struct {
	OpenAPIRequest
	CardID string          `json:"card_id" binding:"required"`
	Action ping.CardAction `json:"action" binding:"required,oneof=freeze unfreeze close update_remark"`
	Remark *string         `json:"remark"`
}

func (s *PingPongOpenAPIService) CreateCard(ctx context.Context, req *CreateCardRequest) (*CardIDData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	productID, err := idconv.FromString(req.CardProductCode)
	if err != nil {
		return nil, err
	}
	virtualAccountID, err := idconv.FromString(req.BudgetID)
	if err != nil {
		return nil, err
	}
	if req.CardholderID != nil {
		if _, err := idconv.FromString(*req.CardholderID); err != nil {
			return nil, err
		}
	}
	if req.Remark != nil && *req.Remark == "" {
		return nil, pingerrors.ErrInvalid
	}
	limits := []*Amount{
		req.PerTransactionLimit,
		req.DailyLimit,
		req.WeeklyLimit,
		req.MonthlyLimit,
		req.LifetimeLimit,
	}
	for _, limit := range limits {
		if limit != nil && (!limit.Amount.IsPositive() || limit.Currency != req.CardCurrency) {
			return nil, pingerrors.ErrInvalid
		}
	}
	raw, err := json.Marshal(req)
	if err != nil {
		return nil, pingerrors.ErrInvalid
	}
	card, err := s.uc.CreateCard(ctx, &biz.CreateCardRequest{
		Notificator:      s.webhook,
		AccountID:        accountID,
		ProductID:        productID,
		VirtualAccountID: virtualAccountID,
		Currency:         req.CardCurrency,
		RequestID:        req.RequestID,
		RawRequest:       raw,
	})
	if err != nil {
		return nil, err
	}
	return &CardIDData{CardID: idconv.ToString(card.ID)}, nil
}

func (s *PingPongOpenAPIService) GetCardDetails(ctx context.Context, req *CardRequest) (*CardDetails, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.uc.GetCard(ctx, &biz.GetCardRequest{
		AccountID: accountID,
		ID:        cardID,
	})
	if err != nil {
		return nil, err
	}
	if card.VirtualAccountID == nil {
		return nil, pingerrors.ErrInvalid
	}
	return &CardDetails{
		CardID:            req.CardID,
		BudgetID:          idconv.ToString(*card.VirtualAccountID),
		CardStatus:        ping.FromGenericCardStatus(card.Status),
		CardNumber:        card.CardNumber,
		CVC:               card.Cvv,
		CardExpiryDate:    card.ExpireAt.Format("01/06"),
		WithdrawalAllowed: true,
		Cancelled:         card.Status == common.CardStatus_Deleted,
		BillingCurrency:   card.CardCurrency,
		CreatedAt:         card.CreatedAt,
	}, nil
}

func (s *PingPongOpenAPIService) GetCardBalance(ctx context.Context, req *CardRequest) (*CardBalance, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.uc.GetCard(ctx, &biz.GetCardRequest{
		AccountID: accountID,
		ID:        cardID,
	})
	if err != nil {
		return nil, err
	}
	if card.Wallet == nil {
		return nil, pingerrors.ErrInvalid
	}
	return &CardBalance{
		CardNumber:       card.CardNumber,
		AvailableBalance: Number{card.Wallet.Available},
		Currency:         card.CardCurrency,
	}, nil
}

func (s *PingPongOpenAPIService) ApplyCardAction(ctx context.Context, req *CardActionRequest) (*Empty, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	if req.Action == ping.UpdateRemark {
		return nil, pingerrors.ErrUnsupported
	}
	if req.Action == ping.Close && (req.Remark == nil || *req.Remark == "") {
		return nil, pingerrors.ErrInvalid
	}
	status := common.CardStatus_Frozen
	if req.Action == ping.Unfreeze {
		status = common.CardStatus_Active
	}
	if req.Action == ping.Close {
		status = common.CardStatus_Deleted
	}
	if err := s.uc.ChangeCard(ctx, &biz.ChangeCardRequest{
		AccountID: accountID,
		ID:        cardID,
		Status:    status,
	}); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}
