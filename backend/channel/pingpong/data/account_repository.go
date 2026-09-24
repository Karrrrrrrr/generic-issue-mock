package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type accountRepository struct{ *PingPongRepository }

func NewAccountRepository(injector *do.Injector) (biz.PingPongAccountRepository, error) {
	return &accountRepository{PingPongRepository: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (repo *accountRepository) Exists(ctx context.Context, req *biz.AccountExistsRequest) (bool, error) {
	table := repo.DB(ctx).Account
	count, err := repo.DB(ctx).Account.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
		).Count()
	return count > 0, err
}

func (repo *accountRepository) Find(ctx context.Context, req *biz.AccountFindRequest) (*model.Account, error) {
	table := repo.DB(ctx).Account
	return repo.DB(ctx).Account.WithContext(ctx).
		Preload(table.Wallet).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
		).First()
}

func (repo *accountRepository) Lock(ctx context.Context, req *biz.AccountLockRequest) (*model.Account, error) {
	table := repo.DB(ctx).Account
	return repo.DB(ctx).Account.WithContext(ctx).
		Preload(table.Wallet).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
		).First()
}

func (repo *accountRepository) Create(ctx context.Context, item *model.Account) error {
	return repo.DB(ctx).Account.WithContext(ctx).Create(item)
}

func (repo *accountRepository) Save(ctx context.Context, item *model.Account) error {
	table := repo.DB(ctx).Account
	_, err := table.WithContext(ctx).Where(
		table.ID.Eq(item.ID),
		table.Channel.Eq(string(common.Channel_PingPong)),
	).UpdateSimple(
		table.Name.Value(item.Name),
		table.WalletID.Value(item.WalletID),
	)
	return err
}

func (repo *accountRepository) List(ctx context.Context, req *biz.AccountListRequest) ([]*model.Account, error) {
	table := repo.DB(ctx).Account
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	statement = statement.Preload(table.Wallet)
	if req.Limit != nil {
		statement = statement.Limit(*req.Limit)
	}
	return statement.Order(table.ID.Desc()).Offset(req.Offset).Find()
}

func (repo *accountRepository) Count(ctx context.Context, req *biz.AccountCountRequest) (int64, error) {
	table := repo.DB(ctx).Account
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	return statement.Count()
}
