package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gorm/clause"
)

type walletRepository struct {
	*Repository
}

func NewWalletRepository(injector do.Injector) (biz.WalletRepo, error) {
	return &walletRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *walletRepository) Create(ctx context.Context, wallet *model.Wallet) error {
	return repo.DB(ctx).Wallet.WithContext(ctx).Create(wallet)
}

func (repo *walletRepository) Exist(ctx context.Context, req *biz.WalletExistRequest) (bool, error) {
	table := repo.DB(ctx).Wallet
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).Count()
	return count > 0, err
}

func (repo *walletRepository) FindByIDWithLock(ctx context.Context, req *biz.WalletFindByIDWithLockRequest) (*model.Wallet, error) {
	table := repo.DB(ctx).Wallet
	return table.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).First()
}

func (repo *walletRepository) UpdateBalance(ctx context.Context, req *biz.WalletUpdateBalanceRequest) error {
	table := repo.DB(ctx).Wallet
	_, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).
		UpdateSimple(
			table.Available.Value(req.Available),
			table.PendingOut.Value(req.PendingOut),
			table.In.Value(req.In),
			table.Out.Value(req.Out),
		)
	return err
}
