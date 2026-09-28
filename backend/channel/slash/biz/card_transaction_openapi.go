package biz

import (
	"context"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type OpenAPIListTransactionsRequest struct {
	AccountID       model.ID
	Offset          int
	Limit           int
	CardID          *model.ID
	AuthorizationID *model.ID
}

func (u *SlashOpenAPIUsecase) ListTransactions(ctx context.Context, req *OpenAPIListTransactionsRequest) ([]*model.CardTransaction, error) {
	items, err := u.cardTransactionRepository.List(ctx, &CardTransactionListRequest{
		AccountIDs:       []model.ID{req.AccountID},
		Offset:           req.Offset,
		Limit:            req.Limit,
		CardIDs:          types.PointerSlice(req.CardID),
		AuthorizationIDs: types.PointerSlice(req.AuthorizationID),
	})
	if err != nil {
		zap.S().Errorw("list slash openapi transactions", "error", err)

		return nil, slasherrors.ErrDatabaseOperation
	}

	return items, nil
}

func (u *SlashOpenAPIUsecase) GetTransaction(ctx context.Context, req *ResourceRequest) (*model.CardTransaction, error) {
	exists, err := u.cardTransactionRepository.ExistByAccountID(ctx, (*CardTransactionExistByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check slash openapi transaction", "error", err)

		return nil, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, slasherrors.ErrResourceNotFound
	}

	item, err := u.cardTransactionRepository.FindByAccountID(ctx, (*CardTransactionFindByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("find slash openapi transaction", "error", err)

		return nil, slasherrors.ErrDatabaseOperation
	}

	return item, nil
}
