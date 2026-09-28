package service

import (
	"context"

	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/pkg/idconv"
)

type TokenRequest struct {
	AppID     string  `form:"app_id" binding:"required"`
	AppSecret *string `form:"app_secret"` // Invalid: mock 不校验应用密钥。
}

type TokenData struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (s *PingPongOpenAPIService) GetAccessToken(ctx context.Context, req *TokenRequest) (*TokenData, error) {
	value, exists := s.apps[req.AppID]
	if !exists {
		return nil, pingerrors.ErrAppMapping
	}
	accountID, err := idconv.FromString(value)
	if err != nil {
		return nil, err
	}
	if _, err := s.uc.Account(ctx, accountID); err != nil {
		return nil, err
	}
	return &TokenData{
		AccessToken: idconv.ToString(accountID),
		ExpiresIn:   7200,
	}, nil
}
