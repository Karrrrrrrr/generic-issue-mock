package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

type walletRepository struct {
	*Repository
}

var _ biz.WalletRepo = (*walletRepository)(nil)

func NewWalletRepository(injector do.Injector) (biz.WalletRepo, error) {
	return &walletRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *walletRepository) Create(ctx context.Context, req *biz.WalletCreateRequest) error {
	return repo.DB(ctx).Wallet.WithContext(ctx).Create(req.Wallet)
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

func (repo *walletRepository) ListByIDsWithLock(ctx context.Context, req *biz.WalletListByIDsWithLockRequest) ([]*model.Wallet, error) {
	table := repo.DB(ctx).Wallet
	return table.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			table.ID.In(req.IDs...),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).
		Order(table.ID.Asc()).
		Find()
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

func (repo *walletRepository) List(ctx context.Context, req *biz.WalletListRequest) ([]*model.Wallet, error) {
	db := repo.DB(ctx)
	table := db.Wallet
	return table.WithContext(ctx).
		Preload(table.Account).
		Where(repo.buildPredicates(ctx, &req.WalletFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (repo *walletRepository) Count(ctx context.Context, req *biz.WalletCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.Wallet.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.WalletFilters)...).
		Count()
}

func (repo *walletRepository) buildPredicates(ctx context.Context, req *biz.WalletFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.Wallet
	predicates := []gen.Condition{table.Channel.Eq(string(req.Channel))}
	if len(req.IDs) != 0 {
		predicates = append(predicates, table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, table.AccountID.In(req.AccountIDs...))
	}
	if len(req.Types) != 0 {
		values := make([]string, 0, len(req.Types))
		for _, value := range req.Types {
			values = append(values, string(value))
		}
		predicates = append(predicates, table.Type.In(values...))
	}
	return predicates
}
