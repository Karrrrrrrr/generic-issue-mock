package biz

import (
	"context"
	"time"

	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type CreateCardHolderRequest struct {
	AccountID              model.ID
	FirstName              string
	LastName               string
	Email                  string
	Mobile                 string
	MobilePrefix           string
	DateOfBirth            *time.Time
	NationalityCountryCode string
	ResidentialAddress     string
	ResidentialCity        string
	ResidentialCountryCode string
	ResidentialPostalCode  string
	ResidentialState       string
	CertType               string
	CertCountryCode        string
	CertID                 string
	Portrait               string
	ReverseSide            string
}

type UpdateCardHolderRequest struct {
	AccountID    model.ID
	CardholderID model.ID
	Email        *string
	Mobile       *string
	MobilePrefix *string
}

func (u *PhotonPayOpenAPIUsecase) CreateCardHolder(ctx context.Context, req *CreateCardHolderRequest) (*model.CardHolder, error) {
	holder := &model.CardHolder{
		AccountID:              req.AccountID,
		Channel:                common.Channel_PhotonPay,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		Mobile:                 req.Mobile,
		MobilePrefix:           req.MobilePrefix,
		DateOfBirth:            req.DateOfBirth,
		NationalityCountryCode: req.NationalityCountryCode,
		ResidentialAddress:     req.ResidentialAddress,
		ResidentialCity:        req.ResidentialCity,
		ResidentialCountryCode: req.ResidentialCountryCode,
		ResidentialPostalCode:  req.ResidentialPostalCode,
		ResidentialState:       req.ResidentialState,
		CertType:               req.CertType,
		CertCountryCode:        req.CertCountryCode,
		CertID:                 req.CertID,
		Portrait:               req.Portrait,
		ReverseSide:            req.ReverseSide,
		Status:                 common.CardHolderStatus_Normal,
		ReviewStatus:           common.CardHolderReviewStatus_Approved,
	}
	if err := u.cardHolderRepo.Create(ctx, holder); err != nil {
		zap.S().Errorw("create photonpay card holder", "error", err)

		return nil, ErrDatabaseOperation
	}
	return holder, nil
}

func (u *PhotonPayOpenAPIUsecase) UpdateCardHolder(ctx context.Context, req *UpdateCardHolderRequest) (*model.CardHolder, error) {
	var holder *model.CardHolder
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		resource := &ResourceRequest{
			AccountID: &req.AccountID,
			ID:        req.CardholderID,
		}
		if err := u.requireCardHolder(txCtx, resource); err != nil {
			return err
		}

		var err error
		holder, err = u.cardHolderRepo.FindCardHolderByAccountID(txCtx, (*CardHolderFindCardHolderByAccountIDRequest)(resource))
		if err != nil {
			zap.S().Errorw("find photonpay card holder", "error", err)

			return ErrDatabaseOperation
		}
		if req.Email != nil {
			holder.Email = *req.Email
		}
		if req.Mobile != nil {
			holder.Mobile = *req.Mobile
		}
		if req.MobilePrefix != nil {
			holder.MobilePrefix = *req.MobilePrefix
		}

		if err := u.cardHolderRepo.Save(txCtx, holder); err != nil {
			zap.S().Errorw("update photonpay card holder", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return holder, nil
}

func (u *PhotonPayOpenAPIUsecase) ListCardHolders(ctx context.Context, req *ListRequest) ([]*model.CardHolder, error) {
	holders, err := u.cardHolderRepo.List(ctx, &CardHolderListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay card holders", "error", err)

		return nil, ErrDatabaseOperation
	}

	return holders, nil
}

func (u *PhotonPayOpenAPIUsecase) requireCardHolder(
	ctx context.Context,
	req *ResourceRequest,
) error {
	exists, err := u.cardHolderRepo.ExistCardHolderByAccountID(ctx, (*CardHolderExistCardHolderByAccountIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check photonpay card holder", "error", err)

		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}
