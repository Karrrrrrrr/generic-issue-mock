package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
)

type virtualAccountRepository struct {
	*Repository
}

var _ biz.VirtualAccountRepo = (*virtualAccountRepository)(nil)

func NewVirtualAccountRepository(injector do.Injector) (biz.VirtualAccountRepo, error) {
	return &virtualAccountRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *virtualAccountRepository) Exist(ctx context.Context, req *biz.VirtualAccountExistRequest) (bool, error) {
	table := repo.DB(ctx).VirtualAccount
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).Count()
	return count > 0, err
}

func (repo *virtualAccountRepository) Find(ctx context.Context, req *biz.VirtualAccountFindRequest) (*model.VirtualAccount, error) {
	table := repo.DB(ctx).VirtualAccount
	return table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.Wallet).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).First()
}

func (repo *virtualAccountRepository) List(ctx context.Context, req *biz.VirtualAccountListRequest) ([]*model.VirtualAccount, error) {
	db := repo.DB(ctx)
	table := db.VirtualAccount
	return table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.Wallet).
		Where(repo.buildPredicates(ctx, &req.VirtualAccountFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (repo *virtualAccountRepository) Count(ctx context.Context, req *biz.VirtualAccountCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.VirtualAccount.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.VirtualAccountFilters)...).
		Count()
}

func (repo *virtualAccountRepository) buildPredicates(ctx context.Context, req *biz.VirtualAccountFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.VirtualAccount
	predicates := []gen.Condition{table.Channel.Eq(string(req.Channel))}
	if len(req.IDs) != 0 {
		predicates = append(predicates, table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, table.AccountID.In(req.AccountIDs...))
	}
	return predicates
}

func (repo *virtualAccountRepository) Create(ctx context.Context, req *biz.VirtualAccountCreateRequest) error {
	table := repo.DB(ctx).VirtualAccount
	return table.WithContext(ctx).Create(req.VirtualAccount)
}
