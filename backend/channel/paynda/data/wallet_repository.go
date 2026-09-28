package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
	"gorm.io/gorm/clause"
)

type walletRepository struct {
	repository *PayndaRepository
}

var _ biz.PayndaWalletRepository = (*walletRepository)(nil)

func NewWalletRepository(injector do.Injector) (biz.PayndaWalletRepository, error) {
	return &walletRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *walletRepository) Create(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Create(item)
}

func (r *walletRepository) ExistByID(ctx context.Context, req *biz.WalletExistByIDRequest) (bool, error) {
	db := r.repository.DB(ctx)
	query := db.Wallet.WithContext(ctx).Where(
		db.Wallet.ID.Eq(req.ID),
		db.Wallet.Channel.Eq(string(enums.Channel_Paynda)),
	)
	if req.AccountID != nil {
		query = query.Where(db.Wallet.AccountID.Eq(*req.AccountID))
	}
	count, err := query.Count()

	return count > 0, err
}

func (r *walletRepository) FindByID(ctx context.Context, req *biz.WalletFindByIDRequest) (*model.Wallet, error) {
	db := r.repository.DB(ctx)
	query := db.Wallet.WithContext(ctx).
		Preload(db.Wallet.Account).
		Where(
			db.Wallet.ID.Eq(req.ID),
			db.Wallet.Channel.Eq(string(enums.Channel_Paynda)),
		)
	if req.AccountID != nil {
		query = query.Where(db.Wallet.AccountID.Eq(*req.AccountID))
	}
	return query.First()
}

func (r *walletRepository) FindByIDForUpdate(ctx context.Context, req *biz.WalletFindByIDForUpdateRequest) (*model.Wallet, error) {
	db := r.repository.DB(ctx)
	query := db.Wallet.WithContext(ctx).
		Preload(db.Wallet.Account).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).
		Where(
			db.Wallet.ID.Eq(req.ID),
			db.Wallet.Channel.Eq(string(enums.Channel_Paynda)),
		)
	if req.AccountID != nil {
		query = query.Where(db.Wallet.AccountID.Eq(*req.AccountID))
	}
	return query.First()
}

func (r *walletRepository) Save(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Save(item)
}

func (r *walletRepository) WalletExists(ctx context.Context, req *biz.ExistWalletRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Wallet.WithContext(ctx).Where(
		db.Wallet.ID.Eq(req.ID),
		db.Wallet.AccountID.Eq(req.AccountID),
		db.Wallet.Channel.Eq(string(enums.Channel_Paynda)),
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
		db.Wallet.Channel.Eq(string(enums.Channel_Paynda)),
	).First()
}

func (r *walletRepository) ListWallets(ctx context.Context, req *biz.WalletListRequest) ([]*model.Wallet, error) {
	db := r.repository.DB(ctx)
	query := db.Wallet.WithContext(ctx).
		Preload(db.Wallet.Account).
		Where(
			db.Wallet.Channel.Eq(string(enums.Channel_Paynda)),
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
		db.Wallet.Channel.Eq(string(enums.Channel_Paynda)),
	).Select(db.Wallet.Available, db.Wallet.In, db.Wallet.Out).Updates(item)
	return err
}
