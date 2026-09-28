package biz

import (
	"context"

	payndaerrors "generic-mock/channel/paynda/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type PayndaCreateCardHolderRequest struct {
	AccountID              model.ID
	FirstName              string
	LastName               string
	MobilePrefix           string
	Mobile                 string
	Email                  string
	ResidentialAddress     string
	ResidentialCity        string
	ResidentialCountryCode string
	ResidentialPostalCode  string
	ResidentialState       string
}

type PayndaUpdateCardHolderRequest struct {
	AccountID              model.ID
	ID                     model.ID
	FirstName              string
	LastName               string
	MobilePrefix           string
	Mobile                 string
	Email                  string
	ResidentialAddress     string
	ResidentialCity        string
	ResidentialCountryCode string
	ResidentialPostalCode  string
	ResidentialState       string
}

func (u *PayndaOpenAPIUsecase) CreateCardHolder(
	ctx context.Context,
	req *PayndaCreateCardHolderRequest,
) (*model.CardHolder, error) {
	holder := &model.CardHolder{
		AccountID:              req.AccountID,
		Channel:                enums.Channel_Paynda,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		MobilePrefix:           req.MobilePrefix,
		Mobile:                 req.Mobile,
		Email:                  req.Email,
		ResidentialAddress:     req.ResidentialAddress,
		ResidentialCity:        req.ResidentialCity,
		ResidentialCountryCode: req.ResidentialCountryCode,
		ResidentialPostalCode:  req.ResidentialPostalCode,
		ResidentialState:       req.ResidentialState,
		Status:                 enums.CardHolderStatus_Normal,
		ReviewStatus:           enums.CardHolderReviewStatus_Approved,
		Shared:                 true,
	}
	if err := u.cardHolderRepository.Create(ctx, holder); err != nil {
		zap.S().Errorw("create paynda card holder", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return holder, nil
}

func (u *PayndaOpenAPIUsecase) GetCardHolder(
	ctx context.Context,
	req *PayndaResourceRequest,
) (*model.CardHolder, error) {
	if err := u.requireCardHolder(ctx, req); err != nil {
		return nil, err
	}

	holder, err := u.cardHolderRepository.FindByAccountID(ctx, (*CardHolderFindByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("find paynda card holder", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return holder, nil
}

func (u *PayndaOpenAPIUsecase) UpdateCardHolder(
	ctx context.Context,
	req *PayndaUpdateCardHolderRequest,
) (*model.CardHolder, error) {
	var holder *model.CardHolder
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		resource := &PayndaResourceRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		}
		if err := u.requireCardHolder(txCtx, resource); err != nil {
			return err
		}

		var err error
		holder, err = u.cardHolderRepository.FindByAccountID(txCtx, (*CardHolderFindByAccountIDRequest)(resource))
		if err != nil {
			zap.S().Errorw("find paynda card holder for update", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		holder.FirstName = req.FirstName
		holder.LastName = req.LastName
		holder.MobilePrefix = req.MobilePrefix
		holder.Mobile = req.Mobile
		holder.Email = req.Email
		holder.ResidentialAddress = req.ResidentialAddress
		holder.ResidentialCity = req.ResidentialCity
		holder.ResidentialCountryCode = req.ResidentialCountryCode
		holder.ResidentialPostalCode = req.ResidentialPostalCode
		holder.ResidentialState = req.ResidentialState
		if err := u.cardHolderRepository.Save(txCtx, holder); err != nil {
			zap.S().Errorw("update paynda card holder", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return holder, nil
}

func (u *PayndaOpenAPIUsecase) ListCardHolders(
	ctx context.Context,
	req *PayndaListRequest,
) ([]*model.CardHolder, error) {
	items, err := u.cardHolderRepository.List(ctx, &CardHolderListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list paynda card holders", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaOpenAPIUsecase) requireCardHolder(
	ctx context.Context,
	req *PayndaResourceRequest,
) error {
	exists, err := u.cardHolderRepository.ExistByAccountID(ctx, (*CardHolderExistByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check paynda card holder", "error", err)
		return payndaerrors.ErrDatabaseOperation
	}
	if !exists {
		return payndaerrors.ErrResourceNotFound
	}

	return nil
}
