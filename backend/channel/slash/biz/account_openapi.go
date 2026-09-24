package biz

import (
	"context"

	"generic-mock/model"

	"go.uber.org/zap"
)

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
