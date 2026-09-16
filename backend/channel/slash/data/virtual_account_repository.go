package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/model"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type virtualAccountRepository struct{ repository *SlashRepository }

func NewVirtualAccountRepository(injector *do.Injector) (biz.SlashVirtualAccountRepository, error) {
	return &virtualAccountRepository{repository: do.MustInvoke[*SlashRepository](injector)}, nil
}

func (r *virtualAccountRepository) List(ctx context.Context) ([]*model.VirtualAccount, error) {
	db := r.repository.DB(ctx)
	return db.VirtualAccount.WithContext(ctx).Preload(db.VirtualAccount.Wallet).Order(db.VirtualAccount.ID.Desc()).Find()
}
func (r *virtualAccountRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.VirtualAccount.WithContext(ctx).Where(db.VirtualAccount.ID.Eq(id)).Count()
	return count > 0, err
}
func (r *virtualAccountRepository) FindByID(ctx context.Context, id model.ID) (*model.VirtualAccount, error) {
	db := r.repository.DB(ctx)
	return db.VirtualAccount.WithContext(ctx).Where(db.VirtualAccount.ID.Eq(id)).First()
}

type walletRepository struct{ repository *SlashRepository }

func NewWalletRepository(injector *do.Injector) (biz.SlashWalletRepository, error) {
	return &walletRepository{repository: do.MustInvoke[*SlashRepository](injector)}, nil
}
func (r *walletRepository) FindByIDForUpdate(ctx context.Context, id model.ID) (*model.Wallet, error) {
	db := r.repository.DB(ctx)
	return db.Wallet.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where(db.Wallet.ID.Eq(id)).First()
}
func (r *walletRepository) Save(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Save(item)
}

var _ biz.SlashVirtualAccountRepository = (*virtualAccountRepository)(nil)
var _ biz.SlashWalletRepository = (*walletRepository)(nil)
