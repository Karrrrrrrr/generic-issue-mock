package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"generic-mock/channel/pingpong/biz"
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/pkg/idconv"
	"generic-mock/model"

	"github.com/samber/do/v2"
	"github.com/shopspring/decimal"
)

type PingPongOpenAPIService struct {
	uc      *biz.PingPongOpenAPIUsecase
	webhook *biz.PingPongWebhookUsecase
	apps    map[string]string
}

func NewOpenAPIService(injector do.Injector) (*PingPongOpenAPIService, error) {
	apps := make(map[string]string)
	if value, exists := os.LookupEnv("PINGPONG_APP_ACCOUNTS"); exists {
		if err := json.Unmarshal([]byte(value), &apps); err != nil {
			return nil, fmt.Errorf("parse PINGPONG_APP_ACCOUNTS: %w", err)
		}
		for appID, accountID := range apps {
			if strings.TrimSpace(appID) == "" {
				return nil, fmt.Errorf("PINGPONG_APP_ACCOUNTS has empty application ID")
			}
			if _, err := idconv.FromString(accountID); err != nil {
				return nil, fmt.Errorf("PINGPONG_APP_ACCOUNTS: invalid account ID for application %q", appID)
			}
		}
	}
	return &PingPongOpenAPIService{
		uc:      do.MustInvoke[*biz.PingPongOpenAPIUsecase](injector),
		apps:    apps,
		webhook: do.MustInvoke[*biz.PingPongWebhookUsecase](injector),
	}, nil
}

type OpenAPIRequest struct {
	Token string `header:"Authorization" json:"-" binding:"required"`
}

func (s *PingPongOpenAPIService) resolveAccountID(ctx context.Context, req *OpenAPIRequest) (int64, error) {
	if !strings.HasPrefix(req.Token, "Bearer ") {
		return 0, pingerrors.ErrInvalid
	}
	id, err := idconv.FromString(strings.TrimPrefix(req.Token, "Bearer "))
	if err != nil {
		return 0, err
	}
	if _, err := s.uc.Account(ctx, id); err != nil {
		return 0, err
	}
	return id, nil
}

type Number struct{ decimal.Decimal }

func (value Number) MarshalJSON() ([]byte, error) { return []byte(value.String()), nil }

type PageRequest struct {
	PageNo   *int `form:"page_no" binding:"omitempty,min=1"`
	PageSize *int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

func (req PageRequest) resolvePagination() (int, int, error) {
	page := 1
	limit := 20
	if req.PageNo != nil {
		page = *req.PageNo
	}
	if req.PageSize != nil {
		limit = *req.PageSize
	}
	if page < 1 || limit < 1 || limit > 100 || page > 1000000 {
		return 0, 0, pingerrors.ErrInvalid
	}
	return page, limit, nil
}

type UIListRequest struct {
	PageRequest
	AccountID *model.ID `form:"account_id" binding:"omitempty,gt=0"`
}

type Empty struct{}

type UIPage[Item any] struct {
	Items []Item `json:"items"`
	Total int64  `json:"total"`
}

func (s *PingPongOpenAPIService) RejectUnsupportedOperation(ctx context.Context, req *OpenAPIRequest) (*Empty, error) {
	if _, err := s.resolveAccountID(ctx, req); err != nil {
		return nil, err
	}
	return nil, pingerrors.ErrUnsupported
}
