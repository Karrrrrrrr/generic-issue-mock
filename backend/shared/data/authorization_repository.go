package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type authorizationRepository struct {
	*Repository
}

func NewAuthorizationRepository(injector *do.Injector) (biz.AuthorizationRepo, error) {
	return &authorizationRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *authorizationRepository) Create(ctx context.Context, authorization *model.Authorization) error {
	return repo.DB(ctx).Authorization.WithContext(ctx).Create(authorization)
}

func (repo *authorizationRepository) Exist(ctx context.Context, req *biz.AuthorizationExistRequest) (bool, error) {
	table := repo.DB(ctx).Authorization
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.CardID.Eq(req.CardID),
		).Count()
	return count > 0, err
}

func (repo *authorizationRepository) FindByIDWithLock(ctx context.Context, req *biz.AuthorizationFindByIDWithLockRequest) (*model.Authorization, error) {
	table := repo.DB(ctx).Authorization
	return table.WithContext(ctx).
		Preload(table.Account).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.CardID.Eq(req.CardID),
		).First()
}
