package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/model"

	"github.com/samber/do"
)

type authorizationRepository struct {
	repository *Repository
}

func NewAuthorizationRepository(injector *do.Injector) (biz.AuthorizationRepository, error) {
	return &authorizationRepository{
		repository: do.MustInvoke[*Repository](injector),
	}, nil
}

func (r *authorizationRepository) Create(ctx context.Context, item *model.Authorization) error {
	return r.repository.DB(ctx).Authorization.WithContext(ctx).Create(item)
}

var _ biz.AuthorizationRepository = (*authorizationRepository)(nil)
