package biz

import (
	"context"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

func (usecase *SlashOpenAPIUsecase) ListAccounts(ctx context.Context, accountID model.ID) ([]*model.Account, error) {
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	items, err := usecase.accountRepository.List(ctx, &AccountListRequest{
		IDs:   []model.ID{accountID},
		Limit: -1,
	})
	if err != nil {
		zap.S().Errorw("list slash OpenAPI accounts", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	return items, nil
}

func (usecase *SlashOpenAPIUsecase) GetAccount(ctx context.Context, id model.ID) (*model.Account, error) {
	exists, err := usecase.accountRepository.Exist(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash OpenAPI account", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, slasherrors.ErrResourceNotFound
	}
	item, err := usecase.accountRepository.Find(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash OpenAPI account", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	return item, nil
}
