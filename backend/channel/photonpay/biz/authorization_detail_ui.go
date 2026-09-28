package biz

import (
	"context"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

type GetAuthorizationDetailRequest struct {
	AccountID model.ID
	ID        model.ID
}

type AuthorizationDetail struct {
	Balance      *AuthorizationBalance
	Card         *model.Card
	Transactions []*model.CardTransaction
}

func (u *PhotonPayUIUsecase) GetAuthorizationDetail(ctx context.Context, req *GetAuthorizationDetailRequest) (*AuthorizationDetail, error) {
	if req.AccountID <= 0 || req.ID <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	exists, err := u.authorizationRepo.AuthorizationExists(ctx, &ExistAuthorizationRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check photonpay authorization detail", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}
	auth, err := u.authorizationRepo.FindAuthorizationDetail(ctx, &FindAuthorizationDetailRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find photonpay authorization detail", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	stages, err := u.cardTransactionRepo.ListStages(ctx, &ListAuthorizationStagesRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("list photonpay authorization detail stages", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	card, err := u.cardRepo.FindCard(ctx, &FindCardRequest{
		AccountID: req.AccountID,
		ID:        auth.CardID,
	})
	if err != nil {
		zap.S().Errorw("find photonpay authorization detail card", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	return &AuthorizationDetail{
		Balance: authorizationBalance(&authorizationBalanceRequest{
			Authorization: auth,
			Stages:        stages,
		}),
		Card:         card,
		Transactions: stages,
	}, nil
}
