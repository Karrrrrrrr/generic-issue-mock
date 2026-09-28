package biz

import (
	"context"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/enums"
	"generic-mock/model"

	"go.uber.org/zap"
)

type UIUpdateAuthorizationConfigRequest struct {
	AccountID     model.ID
	TargetURL     string
	Enabled       bool
	TimeoutMillis int
}

func (u *PhotonPayUIUsecase) GetAuthorizationConfig(ctx context.Context, accountID model.ID) (*model.AuthorizationConfig, error) {
	accountExists, err := u.accountRepo.Exist(ctx, accountID)
	if err != nil {
		zap.S().Errorw("check photonpay authorization config account", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !accountExists {
		return nil, photonpayerrors.ErrResourceNotFound
	}
	exists, err := u.authorizationConfigRepo.ExistByAccountID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("check photonpay authorization config", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}
	item, err := u.authorizationConfigRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("find photonpay authorization config", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	return item, nil
}

func (u *PhotonPayUIUsecase) UpdateAuthorizationConfig(ctx context.Context, req *UIUpdateAuthorizationConfigRequest) (*model.AuthorizationConfig, error) {
	accountExists, err := u.accountRepo.Exist(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check photonpay authorization config account", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !accountExists {
		return nil, photonpayerrors.ErrResourceNotFound
	}
	exists, err := u.authorizationConfigRepo.ExistByAccountID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check photonpay authorization config", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		account, err := u.accountRepo.Find(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay authorization config owner", "error", err)
			return nil, photonpayerrors.ErrDatabaseOperation
		}
		item := &model.AuthorizationConfig{
			Account:       account,
			AccountID:     req.AccountID,
			Channel:       enums.Channel_PhotonPay,
			TargetURL:     req.TargetURL,
			Enabled:       req.Enabled,
			TimeoutMillis: req.TimeoutMillis,
		}
		if err := u.authorizationConfigRepo.Create(ctx, item); err != nil {
			zap.S().Errorw("create photonpay authorization config", "error", err)
			return nil, photonpayerrors.ErrDatabaseOperation
		}
		return item, nil
	}
	item, err := u.authorizationConfigRepo.FindByAccountID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("find photonpay authorization config", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	item.TargetURL = req.TargetURL
	item.Enabled = req.Enabled
	item.TimeoutMillis = req.TimeoutMillis
	if err := u.authorizationConfigRepo.Save(ctx, item); err != nil {
		zap.S().Errorw("save photonpay authorization config", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	return item, nil
}
