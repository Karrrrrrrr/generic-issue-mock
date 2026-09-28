package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do"
)

type virtualAccountRepository struct {
	*Repository
}

func NewVirtualAccountRepository(injector *do.Injector) (biz.VirtualAccountRepo, error) {
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
		Preload(table.Wallet).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).First()
}
