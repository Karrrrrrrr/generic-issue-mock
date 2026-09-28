package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

type accountRepository struct {
	*Repository
}

var _ biz.AccountRepo = (*accountRepository)(nil)

func NewAccountRepository(injector do.Injector) (biz.AccountRepo, error) {
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

func (repo *accountRepository) List(ctx context.Context, req *biz.AccountListRequest) ([]*model.Account, error) {
	db := repo.DB(ctx)
	table := db.Account
	return table.WithContext(ctx).
		Preload(table.Wallet).
		Where(repo.buildPredicates(ctx, &req.AccountFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (repo *accountRepository) Count(ctx context.Context, req *biz.AccountCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.Account.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.AccountFilters)...).
		Count()
}

func (repo *accountRepository) buildPredicates(ctx context.Context, req *biz.AccountFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.Account
	predicates := []gen.Condition{table.Channel.Eq(string(req.Channel))}
	if len(req.IDs) != 0 {
		predicates = append(predicates, table.ID.In(req.IDs...))
	}
	return predicates
}

func (repo *accountRepository) Find(ctx context.Context, req *biz.AccountFindRequest) (*model.Account, error) {
	table := repo.DB(ctx).Account
	return table.WithContext(ctx).
		Preload(table.Wallet).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		First()
}

func (repo *accountRepository) Create(ctx context.Context, req *biz.AccountCreateRequest) error {
	table := repo.DB(ctx).Account
	return table.WithContext(ctx).Create(req.Account)
}

func (repo *accountRepository) Rename(ctx context.Context, req *biz.AccountRenameRequest) error {
	table := repo.DB(ctx).Account
	_, err := table.WithContext(ctx).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		UpdateSimple(table.Name.Value(req.Name))
	return err
}

func (repo *accountRepository) SetWallet(ctx context.Context, req *biz.AccountSetWalletRequest) error {
	table := repo.DB(ctx).Account
	_, err := table.WithContext(ctx).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		UpdateSimple(table.WalletID.Value(req.WalletID))
	return err
}
