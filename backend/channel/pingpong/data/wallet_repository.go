package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type walletRepository struct{ *PingPongRepository }

func NewWalletRepository(injector *do.Injector) (biz.PingPongWalletRepository, error) {
	return &walletRepository{PingPongRepository: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (repo *walletRepository) Lock(ctx context.Context, req *biz.WalletLockRequest) (*model.Wallet, error) {
	table := repo.DB(ctx).Wallet
	return repo.DB(ctx).Wallet.WithContext(ctx).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
			table.AccountID.Eq(req.AccountID),
		).First()
}

func (repo *walletRepository) Create(ctx context.Context, item *model.Wallet) error {
	return repo.DB(ctx).Wallet.WithContext(ctx).Create(item)
}

func (repo *walletRepository) Save(ctx context.Context, item *model.Wallet) error {
	table := repo.DB(ctx).Wallet
	_, err := table.WithContext(ctx).Where(
		table.ID.Eq(item.ID),
		table.Channel.Eq(string(common.Channel_PingPong)),
		table.AccountID.Eq(item.AccountID),
	).UpdateSimple(
		table.Available.Value(item.Available),
		table.PendingOut.Value(item.PendingOut),
		table.In.Value(item.In),
		table.Out.Value(item.Out),
	)
	return err
}
