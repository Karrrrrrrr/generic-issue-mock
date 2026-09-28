package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"

	"gorm.io/gorm/clause"
)

type walletRepository struct {
	repository *SlashRepository
}

var _ biz.SlashWalletRepository = (*walletRepository)(nil)

func NewWalletRepository(injector *do.Injector) (biz.SlashWalletRepository, error) {
	return &walletRepository{
		repository: do.MustInvoke[*SlashRepository](injector),
	}, nil
}

func (r *walletRepository) Create(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Create(item)
}

func (r *walletRepository) FindByIDForUpdate(ctx context.Context, id model.ID) (*model.Wallet, error) {
	db := r.repository.DB(ctx)
	return db.Wallet.WithContext(ctx).
		Preload(db.Wallet.Account).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).Where(
		db.Wallet.ID.Eq(id),
		db.Wallet.Channel.Eq(string(enums.Channel_Slash)),
	).First()
}

func (r *walletRepository) Save(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Save(item)
}

func (r *walletRepository) FindByAccountIDForUpdate(ctx context.Context, req *biz.WalletFindByAccountIDForUpdateRequest) (*model.Wallet, error) {
	db := r.repository.DB(ctx)
	return db.Wallet.WithContext(ctx).
		Preload(db.Wallet.Account).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).Where(
		db.Wallet.ID.Eq(req.ID),
		db.Wallet.AccountID.Eq(*req.AccountID),
		db.Wallet.Channel.Eq(string(enums.Channel_Slash)),
	).First()
}

func (r *walletRepository) WalletExists(ctx context.Context, req *biz.ExistWalletRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Wallet.WithContext(ctx).Where(
		db.Wallet.ID.Eq(req.ID),
		db.Wallet.AccountID.Eq(req.AccountID),
		db.Wallet.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
	return count > 0, err
}

func (r *walletRepository) LockWallet(ctx context.Context, req *biz.LockWalletRequest) (*model.Wallet, error) {
	db := r.repository.DB(ctx)
	return db.Wallet.WithContext(ctx).
		Preload(db.Wallet.Account).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).Where(
		db.Wallet.ID.Eq(req.ID),
		db.Wallet.AccountID.Eq(req.AccountID),
		db.Wallet.Channel.Eq(string(enums.Channel_Slash)),
	).First()
}

func (r *walletRepository) ListWallets(ctx context.Context, req *biz.WalletListRequest) ([]*model.Wallet, error) {
	db := r.repository.DB(ctx)
	query := db.Wallet.WithContext(ctx).
		Preload(db.Wallet.Account).
		Where(
			db.Wallet.Channel.Eq(string(enums.Channel_Slash)),
		)
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.Wallet.AccountID.In(req.AccountIDs...))
	}
	return query.Order(db.Wallet.ID.Desc()).Find()
}

func (r *walletRepository) SaveWallet(ctx context.Context, item *model.Wallet) error {
	db := r.repository.DB(ctx)
	_, err := db.Wallet.WithContext(ctx).Where(
		db.Wallet.ID.Eq(item.ID),
		db.Wallet.AccountID.Eq(item.AccountID),
		db.Wallet.Channel.Eq(string(enums.Channel_Slash)),
	).Select(db.Wallet.Available, db.Wallet.In, db.Wallet.Out).Updates(item)
	return err
}
