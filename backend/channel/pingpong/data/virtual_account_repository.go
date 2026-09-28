package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

type virtualAccountRepository struct{ *PingPongRepository }

func NewVirtualAccountRepository(injector do.Injector) (biz.PingPongVirtualAccountRepository, error) {
	return &virtualAccountRepository{PingPongRepository: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (repo *virtualAccountRepository) Exists(ctx context.Context, req *biz.VirtualAccountExistsRequest) (bool, error) {
	table := repo.DB(ctx).VirtualAccount
	count, err := repo.DB(ctx).VirtualAccount.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
			table.AccountID.Eq(req.AccountID),
		).Count()
	return count > 0, err
}

func (repo *virtualAccountRepository) Find(ctx context.Context, req *biz.VirtualAccountFindRequest) (*model.VirtualAccount, error) {
	table := repo.DB(ctx).VirtualAccount
	return repo.DB(ctx).VirtualAccount.WithContext(ctx).
		Preload(table.Wallet).
		Preload(table.Account).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
			table.AccountID.Eq(req.AccountID),
		).First()
}

func (repo *virtualAccountRepository) Create(ctx context.Context, item *model.VirtualAccount) error {
	return repo.DB(ctx).VirtualAccount.WithContext(ctx).Create(item)
}

func (repo *virtualAccountRepository) List(ctx context.Context, req *biz.VirtualAccountListRequest) ([]*model.VirtualAccount, error) {
	table := repo.DB(ctx).VirtualAccount
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		statement = statement.Where(table.AccountID.In(req.AccountIDs...))
	}
	statement = statement.Preload(table.Wallet)
	statement = statement.Preload(table.Account)
	if req.Limit != nil {
		statement = statement.Limit(*req.Limit)
	}
	return statement.Order(table.ID.Desc()).Offset(req.Offset).Find()
}

func (repo *virtualAccountRepository) Count(ctx context.Context, req *biz.VirtualAccountCountRequest) (int64, error) {
	table := repo.DB(ctx).VirtualAccount
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		statement = statement.Where(table.AccountID.In(req.AccountIDs...))
	}
	return statement.Count()
}
