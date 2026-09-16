package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
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

func (r *authorizationRepository) List(ctx context.Context, req *biz.ListRequest) ([]*model.Authorization, error) {
	db := r.repository.DB(ctx)
	return db.Authorization.WithContext(ctx).Where(db.Authorization.Channel.Eq(string(enums.Channel_PhotonPay))).Order(db.Authorization.ID.Desc()).Offset(req.Offset).Limit(req.Limit).Find()
}

var _ biz.AuthorizationRepository = (*authorizationRepository)(nil)
