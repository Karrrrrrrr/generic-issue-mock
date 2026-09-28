package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
)

type UITransferData struct {
	ID             model.ID                  `json:"id"`
	AccountID      model.ID                  `json:"account_id"`
	AccountName    string                    `json:"account_name"`
	RequestID      string                    `json:"request_id"`
	Kind           common.WalletTransferKind `json:"kind"`
	Amount         Number                    `json:"amount"`
	Currency       common.Currency           `json:"currency"`
	SourceWalletID model.ID                  `json:"source_wallet_id"`
	TargetWalletID model.ID                  `json:"target_wallet_id"`
	Status         common.OperationStatus    `json:"status"`
	CreatedAt      time.Time                 `json:"created_at"`
}

func (s *PingPongUIService) ListTransfers(ctx context.Context, req *UIListRequest) (*UIPage[UITransferData], error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, pingerrors.ErrInvalid
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
			ID:             item.ID,
			AccountID:      item.AccountID,
			AccountName:    item.Account.GetName(),
			RequestID:      item.RequestID,
			Kind:           item.Kind,
			Amount:         Number{item.Amount},
			Currency:       item.Currency,
			SourceWalletID: item.SourceWalletID,
			TargetWalletID: item.TargetWalletID,
			Status:         common.OperationStatus_Succeed,
			CreatedAt:      item.CreatedAt,
		})
	}
	return result, nil
}

type UIRecordData struct {
	RecordID model.ID `json:"record_id"`
}
