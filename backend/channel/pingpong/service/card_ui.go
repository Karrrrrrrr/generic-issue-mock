package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	ping "generic-mock/channel/pingpong/enums"
	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
)

type UICardData struct {
	WalletID         model.ID          `json:"wallet_id"`
	CVV              string            `json:"cvv"`
	ExpiresAt        time.Time         `json:"expires_at"`
	CardType         common.CardType   `json:"card_type"`
	ID               model.ID          `json:"id"`
	AccountID        model.ID          `json:"account_id"`
	AccountName      string            `json:"account_name"`
	VirtualAccountID model.ID          `json:"virtual_account_id"`
	CardNumber       string            `json:"card_number"`
	CardBin          string            `json:"card_bin"`
	Status           common.CardStatus `json:"status"`
	Balance          Number            `json:"balance"`
	Reserved         Number            `json:"reserved"`
	Currency         common.Currency   `json:"currency"`
	CreatedAt        time.Time         `json:"created_at"`
}

type UIListCardsRequest struct {
	UIListRequest
	UIListTimeRange
	CardNumber *string            `form:"card_number" binding:"omitempty,min=1"`
	ID         *model.ID          `form:"card_id" binding:"omitempty,gt=0"`
	Status     *common.CardStatus `form:"status" binding:"omitempty,oneof=inactive active freezing frozen deleting deleted"`
}

type UIChangeCardRequest struct {
	UIResourceRequest
	Status common.CardStatus `json:"status" binding:"required,oneof=active frozen deleted"`
}

type UIFundCardRequest struct {
	UIResourceRequest
	Action    ping.FundingAction `json:"action" binding:"required,oneof=top_up withdraw"`
	Amount    Number             `json:"amount"`
	RequestID string             `json:"request_id" binding:"required"`
}

func (s *PingPongUIService) ListCards(ctx context.Context, req *UIListCardsRequest) (*UIPage[UICardData], error) {
	if err := req.UIListTimeRange.Validate(); err != nil {
		return nil, err
	}
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	cardID := req.ID
	if cardID != nil && *cardID <= 0 {
		return nil, pingerrors.ErrInvalid
	}

	page, limit, err := req.PageRequest.resolvePagination()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListCards(ctx, &biz.UIListCardsRequest{
		AccountID:   accountID,
		ID:          cardID,
		Status:      req.Status,
		CardNumber:  req.CardNumber,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		Offset:      (page - 1) * limit,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	result := &UIPage[UICardData]{
		Items: make([]UICardData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		if item.Wallet == nil || item.VirtualAccountID == nil {
			return nil, pingerrors.ErrInvalid
		}
		result.Items = append(result.Items, UICardData{
			ID:               item.ID,
			WalletID:         item.WalletID,
			CVV:              item.Cvv,
			ExpiresAt:        item.ExpireAt,
			CardType:         item.CardType,
			AccountID:        item.AccountID,
			AccountName:      item.Account.GetName(),
			VirtualAccountID: *item.VirtualAccountID,
			CardNumber:       item.CardNumber,
			CardBin:          item.CardBin,
			Status:           item.Status,
			Balance:          Number{item.Wallet.Available},
			Reserved:         Number{item.Wallet.PendingOut},
			Currency:         item.CardCurrency,
			CreatedAt:        item.CreatedAt,
		})
	}
	return result, nil
}

func (s *PingPongUIService) ChangeCardStatus(ctx context.Context, req *UIChangeCardRequest) (*Empty, error) {
	accountID, id := req.AccountID, req.ID
	if err := req.UIResourceRequest.Validate(); err != nil {
		return nil, err
	}
	if err := s.uc.ChangeCard(ctx, &biz.UIChangeCardRequest{
		AccountID: accountID,
		ID:        id,
		Status:    req.Status,
	}); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

func (s *PingPongUIService) FundCard(ctx context.Context, req *UIFundCardRequest) (*UIRecordData, error) {
	accountID, id := req.AccountID, req.ID
	if err := req.UIResourceRequest.Validate(); err != nil {
		return nil, err
	}
	item, err := s.uc.FundCard(ctx, &biz.UICardFundingRequest{
		AccountID: accountID,
		CardID:    id,
		Withdraw:  req.Action == ping.Withdraw,
		Amount:    req.Amount.Decimal,
		RequestID: req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	return &UIRecordData{RecordID: item.ID}, nil
}
