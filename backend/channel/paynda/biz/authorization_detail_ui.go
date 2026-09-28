package biz

import (
	"context"

	payndaerrors "generic-mock/channel/paynda/errors"
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

func (u *PayndaUIUsecase) GetAuthorizationDetail(ctx context.Context, req *GetAuthorizationDetailRequest) (*AuthorizationDetail, error) {
	if req.AccountID <= 0 || req.ID <= 0 {
		return nil, payndaerrors.ErrInvalidOperation
	}
	exists, err := u.authorizationRepository.AuthorizationExists(ctx, &ExistAuthorizationRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check paynda authorization detail", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, payndaerrors.ErrResourceNotFound
	}
	auth, err := u.authorizationRepository.FindAuthorizationDetail(ctx, &FindAuthorizationDetailRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find paynda authorization detail", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	stages, err := u.cardTransactionRepository.ListStages(ctx, &ListAuthorizationStagesRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("list paynda authorization detail stages", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	card, err := u.cardRepository.FindCard(ctx, &FindCardRequest{
		AccountID: req.AccountID,
		ID:        auth.CardID,
	})
	if err != nil {
		zap.S().Errorw("find paynda authorization detail card", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
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
