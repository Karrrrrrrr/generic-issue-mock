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

type UICardData struct {
	ID               string          `json:"id"`
	AccountID        string          `json:"account_id"`
	AccountName      string          `json:"account_name"`
	VirtualAccountID string          `json:"virtual_account_id"`
	CardNumber       string          `json:"card_number"`
	CardBin          string          `json:"card_bin"`
	Status           ping.CardStatus `json:"status"`
	Balance          Number          `json:"balance"`
	Reserved         Number          `json:"reserved"`
	Currency         common.Currency `json:"currency"`
	CreatedAt        time.Time       `json:"created_at"`
}

type UIListCardsRequest struct {
	UIListRequest
	ID     *string          `form:"card_id"`
	Status *ping.CardStatus `form:"status" binding:"omitempty,oneof=ACTIVE REVOKED CANCELLED"`
}

type UIChangeCardRequest struct {
	UIResourceRequest
	Status ping.CardStatus `json:"status" binding:"required,oneof=ACTIVE REVOKED CANCELLED"`
}

type UIFundCardRequest struct {
	UIResourceRequest
	Action    ping.FundingAction `json:"action" binding:"required,oneof=top_up withdraw"`
	Amount    Number             `json:"amount"`
	RequestID string             `json:"request_id" binding:"required"`
}

func (s *PingPongUIService) ListCards(ctx context.Context, req *UIListCardsRequest) (*UIPage[UICardData], error) {
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromOptionalString(req.ID)
	if err != nil {
		return nil, err
	}
	var status *common.CardStatus
	if req.Status != nil {
		value := ping.ToGenericCardStatus(*req.Status)
		status = &value
	}
	page, limit, err := req.PageRequest.resolvePagination()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListCards(ctx, &biz.UIListCardsRequest{
		AccountID: accountID,
		ID:        cardID,
		Status:    status,
		Offset:    (page - 1) * limit,
		Limit:     limit,
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
			ID:               idconv.ToString(item.ID),
			AccountID:        idconv.ToString(item.AccountID),
			AccountName:      item.Account.GetName(),
			VirtualAccountID: idconv.ToString(*item.VirtualAccountID),
			CardNumber:       item.CardNumber,
			CardBin:          item.CardBin,
			Status:           ping.FromGenericCardStatus(item.Status),
			Balance:          Number{item.Wallet.Available},
			Reserved:         Number{item.Wallet.PendingOut},
			Currency:         item.CardCurrency,
			CreatedAt:        item.CreatedAt,
		})
	}
	return result, nil
}

func (s *PingPongUIService) ChangeCardStatus(ctx context.Context, req *UIChangeCardRequest) (*Empty, error) {
	accountID, id, err := req.UIResourceRequest.ParseAccountAndResourceIDs()
	if err != nil {
		return nil, err
	}
	if err := s.uc.ChangeCard(ctx, &biz.UIChangeCardRequest{
		AccountID: accountID,
		ID:        id,
		Status:    ping.ToGenericCardStatus(req.Status),
	}); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

func (s *PingPongUIService) FundCard(ctx context.Context, req *UIFundCardRequest) (*RecordData, error) {
	accountID, id, err := req.UIResourceRequest.ParseAccountAndResourceIDs()
	if err != nil {
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
	return &RecordData{RecordID: idconv.ToString(item.ID)}, nil
}
