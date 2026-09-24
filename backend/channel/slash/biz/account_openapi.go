package biz

import (
	"context"

	"generic-mock/model"

	"go.uber.org/zap"
)

func (usecase *SlashOpenAPIUsecase) ListAccounts(ctx context.Context, accountID model.ID) ([]*model.Account, error) {
	if accountID <= 0 {
		return nil, ErrInvalidOperation
	}
	items, err := usecase.accountRepository.List(ctx, &AccountListRequest{
		IDs:   []model.ID{accountID},
		Limit: -1,
	})
	if err != nil {
		zap.S().Errorw("list slash OpenAPI accounts", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

func (usecase *SlashOpenAPIUsecase) GetAccount(ctx context.Context, id model.ID) (*model.Account, error) {
	exists, err := usecase.accountRepository.Exist(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash OpenAPI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := usecase.accountRepository.Find(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash OpenAPI account", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}
