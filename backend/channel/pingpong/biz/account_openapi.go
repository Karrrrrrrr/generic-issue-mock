package biz

import (
	"context"

	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

func (uc *PingPongOpenAPIUsecase) Account(ctx context.Context, accountID model.ID) (*model.Account, error) {
	if accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.accountRepo.Exists(ctx, &AccountExistsRequest{ID: accountID})
	if err != nil {
		zap.S().Errorw("check pingpong token account", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	account, err := uc.accountRepo.Find(ctx, &AccountFindRequest{ID: accountID})
	if err != nil {
		zap.S().Errorw("find pingpong token account", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return account, nil
}

func (uc *PingPongOpenAPIUsecase) lockAccount(ctx context.Context, accountID model.ID) (*model.Account, error) {
	if accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.accountRepo.Exists(ctx, &AccountExistsRequest{ID: accountID})
	if err != nil {
		zap.S().Errorw("check pingpong account", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	account, err := uc.accountRepo.Lock(ctx, &AccountLockRequest{ID: accountID})
	if err != nil {
		zap.S().Errorw("lock pingpong account", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return account, nil
}
