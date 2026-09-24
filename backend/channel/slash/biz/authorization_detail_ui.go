package biz

import (
	"context"

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

func (u *SlashUIUsecase) GetAuthorizationDetail(ctx context.Context, req *GetAuthorizationDetailRequest) (*AuthorizationDetail, error) {
	if req.AccountID <= 0 || req.ID <= 0 {
		return nil, ErrInvalidOperation
	}
	exists, err := u.authorizationRepository.AuthorizationExists(ctx, &ExistAuthorizationRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check slash authorization detail", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	auth, err := u.authorizationRepository.FindAuthorizationDetail(ctx, &FindAuthorizationDetailRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find slash authorization detail", "error", err)
		return nil, ErrDatabaseOperation
	}
	stages, err := u.cardTransactionRepository.ListStages(ctx, &ListAuthorizationStagesRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("list slash authorization detail stages", "error", err)
		return nil, ErrDatabaseOperation
	}
	card, err := u.cardRepository.FindCard(ctx, &FindCardRequest{
		AccountID: req.AccountID,
		ID:        auth.CardID,
	})
	if err != nil {
		zap.S().Errorw("find slash authorization detail card", "error", err)
		return nil, ErrDatabaseOperation
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
