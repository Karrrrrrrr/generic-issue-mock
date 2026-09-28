package biz

import (
	"context"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type UICreateCardHolderRequest struct {
	AccountID model.ID
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (u *PhotonPayUIUsecase) CreateCardHolder(ctx context.Context, req *UICreateCardHolderRequest) (*model.CardHolder, error) {
	dateOfBirth := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	holder := &model.CardHolder{
		AccountID:              req.AccountID,
		Channel:                enums.Channel_PhotonPay,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		Mobile:                 req.Mobile,
		MobilePrefix:           photon.DefaultMobilePrefix,
		DateOfBirth:            &dateOfBirth,
		NationalityCountryCode: photon.DefaultNationalityCountryCode,
		Status:                 enums.CardHolderStatus_Normal,
		ReviewStatus:           enums.CardHolderReviewStatus_Approved,
	}
	if err := u.cardHolderRepo.Create(ctx, holder); err != nil {
		zap.S().Errorw("create photonpay UI card holder", "error", err)

		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return holder, nil
}

func (u *PhotonPayUIUsecase) ListCardHolders(ctx context.Context, req *ListRequest) ([]*model.CardHolder, int64, error) {
	holders, err := u.cardHolderRepo.List(ctx, &CardHolderListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay UI card holders", "error", err)

		return nil, 0, photonpayerrors.ErrDatabaseOperation
	}

	total, err := u.cardHolderRepo.Count(ctx, &CardHolderCountRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
	})
	if err != nil {
		zap.S().Errorw("count photonpay UI card holders", "error", err)
		return nil, 0, photonpayerrors.ErrDatabaseOperation
	}

	return holders, total, nil
}
