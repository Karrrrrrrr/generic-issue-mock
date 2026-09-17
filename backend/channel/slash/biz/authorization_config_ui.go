package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"

	"go.uber.org/zap"
)

type UpdateAuthorizationConfigRequest struct {
	AccountID     model.ID
	TargetURL     string
	Enabled       bool
	TimeoutMillis int
}

func (u *SlashUIUsecase) GetAuthorizationConfig(
	ctx context.Context,
	accountID model.ID,
) (*model.AuthorizationConfig, error) {
	accountExists, err := u.accountRepository.Exist(ctx, accountID)
	if err != nil {
		zap.S().Errorw("check slash authorization config account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !accountExists {
		return nil, ErrResourceNotFound
	}
	exists, err := u.authorizationConfigRepo.ExistByAccountID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("check slash authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := u.authorizationConfigRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("find slash authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *SlashUIUsecase) UpdateAuthorizationConfig(
	ctx context.Context,
	req *UpdateAuthorizationConfigRequest,
) (*model.AuthorizationConfig, error) {
	accountExists, err := u.accountRepository.Exist(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check slash authorization config account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !accountExists {
		return nil, ErrResourceNotFound
	}
	exists, err := u.authorizationConfigRepo.ExistByAccountID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check slash authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		account, err := u.accountRepository.Find(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find slash authorization config owner", "error", err)
			return nil, ErrDatabaseOperation
		}
		item := &model.AuthorizationConfig{
			Account:       account,
			AccountID:     req.AccountID,
			Channel:       enums.Channel_Slash,
			TargetURL:     req.TargetURL,
			Enabled:       req.Enabled,
			TimeoutMillis: req.TimeoutMillis,
		}
		if err := u.authorizationConfigRepo.Create(ctx, item); err != nil {
			zap.S().Errorw("create slash authorization config", "error", err)
			return nil, ErrDatabaseOperation
		}
		return item, nil
	}
	item, err := u.authorizationConfigRepo.FindByAccountID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("find slash authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.TargetURL = req.TargetURL
	item.Enabled = req.Enabled
	item.TimeoutMillis = req.TimeoutMillis
	if err := u.authorizationConfigRepo.Save(ctx, item); err != nil {
		zap.S().Errorw("save slash authorization config", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}
