package biz

import (
	"context"

	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type PayndaCardTransactionDetail struct {
	Transaction   *model.CardTransaction
	Authorization *model.Authorization
}

func (u *PayndaOpenAPIUsecase) GetCardTransaction(
	ctx context.Context,
	req *PayndaResourceRequest,
) (*PayndaCardTransactionDetail, error) {
	exists, err := u.cardTransactionRepository.ExistByAccountID(ctx, (*CardTransactionExistByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check paynda card transaction", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	transaction, err := u.cardTransactionRepository.FindByAccountID(ctx, (*CardTransactionFindByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("find paynda card transaction", "error", err)
		return nil, ErrDatabaseOperation
	}

	details, err := u.cardTransactionDetails(ctx, []*model.CardTransaction{transaction})
	if err != nil {
		return nil, err
	}

	return details[0], nil
}

func (u *PayndaOpenAPIUsecase) ListCardTransactions(
	ctx context.Context,
	req *PayndaListTransactionsRequest,
) ([]*PayndaCardTransactionDetail, error) {
	items, err := u.cardTransactionRepository.List(ctx, &CardTransactionListRequest{
		AccountIDs:     types.PointerSlice(req.AccountID),
		Offset:         req.Offset,
		Limit:          req.Limit,
		CardIDs:        types.PointerSlice(req.CardID),
		StartCreatedAt: req.StartCreatedAt,
		EndCreatedAt:   req.EndCreatedAt,
		Types:          req.Types,
	})
	if err != nil {
		zap.S().Errorw("list paynda card transactions", "error", err)
		return nil, ErrDatabaseOperation
	}

	return u.cardTransactionDetails(ctx, items)
}

func (u *PayndaOpenAPIUsecase) cardTransactionDetails(
	ctx context.Context,
	transactions []*model.CardTransaction,
) ([]*PayndaCardTransactionDetail, error) {
	authorizationIDs := make([]model.ID, 0, len(transactions))
	seenAuthorizationIDs := make(map[model.ID]struct{}, len(transactions))
	for _, transaction := range transactions {
		if transaction.AuthorizationID == 0 {
			continue
		}
		if _, exists := seenAuthorizationIDs[transaction.AuthorizationID]; exists {
			continue
		}
		seenAuthorizationIDs[transaction.AuthorizationID] = struct{}{}
		authorizationIDs = append(authorizationIDs, transaction.AuthorizationID)
	}

	authorizationsByID := make(map[model.ID]*model.Authorization, len(authorizationIDs))
	if len(authorizationIDs) > 0 {
		authorizations, err := u.authorizationRepository.ListByIDs(ctx, &PayndaListAuthorizationsByIDsRequest{
			IDs: authorizationIDs,
		})
		if err != nil {
			zap.S().Errorw("list paynda transaction authorizations", "error", err)
			return nil, ErrDatabaseOperation
		}
		for _, authorization := range authorizations {
			authorizationsByID[authorization.ID] = authorization
		}
	}

	details := make([]*PayndaCardTransactionDetail, 0, len(transactions))
	for _, transaction := range transactions {
		details = append(details, &PayndaCardTransactionDetail{
			Transaction:   transaction,
			Authorization: authorizationsByID[transaction.AuthorizationID],
		})
	}

	return details, nil
}
