package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type authorizationConfigRepository struct {
	repository *SlashRepository
}

func NewAuthorizationConfigRepository(
	injector *do.Injector,
) (biz.SlashAuthorizationConfigRepository, error) {
	return &authorizationConfigRepository{
		repository: do.MustInvoke[*SlashRepository](injector),
	}, nil
}

func (r *authorizationConfigRepository) Create(
	ctx context.Context,
	item *model.AuthorizationConfig,
) error {
	return r.repository.DB(ctx).AuthorizationConfig.WithContext(ctx).Create(item)
}

func (r *authorizationConfigRepository) ExistByAccountID(
	ctx context.Context,
	accountID model.ID,
) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.AuthorizationConfig.WithContext(ctx).Where(
		db.AuthorizationConfig.AccountID.Eq(accountID),
		db.AuthorizationConfig.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
	return count > 0, err
}

func (r *authorizationConfigRepository) FindByAccountID(
	ctx context.Context,
	accountID model.ID,
) (*model.AuthorizationConfig, error) {
	db := r.repository.DB(ctx)
	return db.AuthorizationConfig.WithContext(ctx).
		Preload(db.AuthorizationConfig.Account).
		Where(
			db.AuthorizationConfig.AccountID.Eq(accountID),
			db.AuthorizationConfig.Channel.Eq(string(enums.Channel_Slash)),
		).First()
}

func (r *authorizationConfigRepository) Save(
	ctx context.Context,
	item *model.AuthorizationConfig,
) error {
	return r.repository.DB(ctx).AuthorizationConfig.WithContext(ctx).Save(item)
}

var _ biz.SlashAuthorizationConfigRepository = (*authorizationConfigRepository)(nil)
