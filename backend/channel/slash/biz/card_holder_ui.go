package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type ListCardHoldersRequest struct {
	AccountID *model.ID
	Offset    int
	Limit     int
}

type CreateCardHolderRequest struct {
	AccountID model.ID
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (u *SlashUIUsecase) CreateCardHolder(ctx context.Context, req *CreateCardHolderRequest) (*model.CardHolder, error) {
	exists, err := u.accountRepository.Exist(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check slash holder account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	holder := &model.CardHolder{
		AccountID:    req.AccountID,
		Channel:      enums.Channel_Slash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Mobile:       req.Mobile,
		Status:       enums.CardHolderStatus_Normal,
		ReviewStatus: enums.CardHolderReviewStatus_Approved,
		Shared:       true,
	}
	if err := u.cardHolderRepository.Create(ctx, holder); err != nil {
		zap.S().Errorw("create slash card holder", "error", err)
		return nil, ErrDatabaseOperation
	}

	return holder, nil
}

func (u *SlashUIUsecase) ListCardHolders(ctx context.Context, req *ListCardHoldersRequest) ([]*model.CardHolder, int64, error) {
	items, err := u.cardHolderRepository.List(ctx, &CardHolderListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list slash card holders", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.cardHolderRepository.Count(ctx, &CardHolderCountRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("count slash card holders", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *SlashUIUsecase) requireCardHolder(ctx context.Context, id model.ID) error {
	exists, err := u.cardHolderRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card holder", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	return nil
}
