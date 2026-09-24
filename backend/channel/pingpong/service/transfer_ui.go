package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	ping "generic-mock/channel/pingpong/enums"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
)

type UITransferData struct {
	ID             string             `json:"id"`
	AccountID      string             `json:"account_id"`
	AccountName    string             `json:"account_name"`
	RequestID      string             `json:"request_id"`
	Kind           ping.TransferKind  `json:"kind"`
	Amount         Number             `json:"amount"`
	Currency       common.Currency    `json:"currency"`
	SourceWalletID string             `json:"source_wallet_id"`
	TargetWalletID string             `json:"target_wallet_id"`
	Status         ping.FundingStatus `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
}

func (s *PingPongUIService) Transfers(ctx context.Context, req *UIListRequest) (*UIPage[UITransferData], error) {
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
	}
	page, limit, err := req.PageRequest.resolvePagination()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListTransfers(ctx, &biz.UIListTransfersRequest{
		AccountID: accountID,
		Offset:    (page - 1) * limit,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	result := &UIPage[UITransferData]{
		Items: make([]UITransferData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, UITransferData{
			ID:             idconv.ToString(item.ID),
			AccountID:      idconv.ToString(item.AccountID),
			AccountName:    item.Account.GetName(),
			RequestID:      item.RequestID,
			Kind:           ping.FromGenericTransferKind(item.Kind),
			Amount:         Number{item.Amount},
			Currency:       item.Currency,
			SourceWalletID: idconv.ToString(item.SourceWalletID),
			TargetWalletID: idconv.ToString(item.TargetWalletID),
			Status:         ping.FundingSuccess,
			CreatedAt:      item.CreatedAt,
		})
	}
	return result, nil
}
