package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type accountRepository struct {
	*Repository
}

func NewAccountRepository(injector *do.Injector) (biz.AccountRepo, error) {
	return &accountRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *accountRepository) Exist(ctx context.Context, req *biz.AccountExistRequest) (bool, error) {
	table := repo.DB(ctx).Account
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(req.Channel)),
		).Count()
	return count > 0, err
}

func (repo *accountRepository) FindByIDWithLock(ctx context.Context, req *biz.AccountFindByIDWithLockRequest) (*model.Account, error) {
	table := repo.DB(ctx).Account
	return table.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(req.Channel)),
		).First()
}
