package biz

import (
	"context"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

func (u *PhotonPayOpenAPIUsecase) ListTransactions(ctx context.Context, req *ListRequest) ([]*model.CardTransaction, error) {
	transactions, err := u.cardTransactionRepo.ListTransactions(ctx, &CardTransactionListTransactionsRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay card transactions", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return transactions, nil
}
