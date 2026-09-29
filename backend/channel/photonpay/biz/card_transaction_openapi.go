package biz

import (
	"context"
	"time"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharedbiz "generic-mock/shared/biz"

	"go.uber.org/zap"
)

type ListTransactionsRequest struct {
	AccountID   *model.ID
	CardID      *model.ID
	ID          *model.ID
	RequestID   *string
	Types       []enums.CardTransactionType
	Statuses    []enums.CardTransactionStatus
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Offset      int
	Limit       int
}

func (u *PhotonPayOpenAPIUsecase) ListTransactions(ctx context.Context, req *ListTransactionsRequest) ([]*model.CardTransaction, error) {
	transactions, err := u.cardTransactionRepo.List(ctx, &sharedbiz.CardTransactionListRequest{
		CardTransactionFilters: sharedbiz.CardTransactionFilters{
			Channel:     enums.Channel_PhotonPay,
			AccountIDs:  types.PointerSlice(req.AccountID),
			IDs:         types.PointerSlice(req.ID),
			RequestIDs:  types.PointerSlice(req.RequestID),
			CardIDs:     types.PointerSlice(req.CardID),
			Types:       req.Types,
			Statuses:    req.Statuses,
			CreatedFrom: req.CreatedFrom,
			CreatedTo:   req.CreatedTo,
		},
		Limit:  &req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay card transactions", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return transactions, nil
}
