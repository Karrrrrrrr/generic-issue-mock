package service

import (
	"context"
	"time"

	"generic-mock/channel/slash/biz"
	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/model"
)

type AuthorizationConfigData struct {
	AccountName   string    `json:"account_name"`
	AccountID     model.ID  `json:"account_id"`
	TargetURL     string    `json:"target_url"`
	Enabled       bool      `json:"enabled"`
	TimeoutMillis int       `json:"timeout_millis"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type GetAuthorizationConfigRequest struct {
	AccountID model.ID `form:"account_id" binding:"required,gt=0"`
}

func (s *SlashUIService) GetAuthorizationConfig(
	ctx context.Context,
	req *GetAuthorizationConfigRequest,
) (*AuthorizationConfigData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.GetAuthorizationConfig(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return toAuthorizationConfigData(item), nil
}

type UpdateAuthorizationConfigRequest struct {
	AccountID     model.ID `json:"account_id" binding:"required,gt=0"`
	TargetURL     string   `json:"target_url" binding:"required,url"`
	Enabled       bool     `json:"enabled"`
	TimeoutMillis int      `json:"timeout_millis" binding:"required,min=1"`
}

func (s *SlashUIService) UpdateAuthorizationConfig(
	ctx context.Context,
	req *UpdateAuthorizationConfigRequest,
) (*AuthorizationConfigData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.usecase.UpdateAuthorizationConfig(ctx, &biz.UpdateAuthorizationConfigRequest{
		AccountID:     accountID,
		TargetURL:     req.TargetURL,
		Enabled:       req.Enabled,
		TimeoutMillis: req.TimeoutMillis,
	})
	if err != nil {
		return nil, err
	}
	return toAuthorizationConfigData(item), nil
}

func toAuthorizationConfigData(item *model.AuthorizationConfig) *AuthorizationConfigData {
	return &AuthorizationConfigData{
		AccountID:     item.AccountID,
		AccountName:   item.Account.GetName(),
		TargetURL:     item.TargetURL,
		Enabled:       item.Enabled,
		TimeoutMillis: item.TimeoutMillis,
		UpdatedAt:     item.UpdatedAt,
	}
}
