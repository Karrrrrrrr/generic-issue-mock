package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type PayndaUICreateCardHolderRequest struct {
	AccountID model.ID
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (u *PayndaUIUsecase) CreateCardHolder(ctx context.Context, req *PayndaUICreateCardHolderRequest) (*model.CardHolder, error) {
	exists, err := u.accountRepository.ExistByID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check paynda UI card holder account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	holder := &model.CardHolder{
		AccountID:    req.AccountID,
		Channel:      enums.Channel_Paynda,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Mobile:       req.Mobile,
		Status:       enums.CardHolderStatus_Normal,
		ReviewStatus: enums.CardHolderReviewStatus_Approved,
		Shared:       true,
	}
	if err := u.cardHolderRepository.Create(ctx, holder); err != nil {
		zap.S().Errorw("create paynda UI card holder", "error", err)
		return nil, ErrDatabaseOperation
	}
	return holder, nil
}

func (u *PayndaUIUsecase) ListCardHolders(ctx context.Context, req *PayndaListRequest) ([]*model.CardHolder, int64, error) {
	items, err := u.cardHolderRepository.List(ctx, &CardHolderListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list paynda UI card holders", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	total, err := u.cardHolderRepository.Count(ctx, &CardHolderCountRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
	})
	if err != nil {
		zap.S().Errorw("count paynda UI card holders", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}
