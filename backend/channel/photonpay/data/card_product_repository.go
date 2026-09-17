package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type cardProductRepository struct {
	repository *Repository
}

func NewCardProductRepository(injector *do.Injector) (biz.CardProductRepository, error) {
	return &cardProductRepository{
		repository: do.MustInvoke[*Repository](injector),
	}, nil
}

func (r *cardProductRepository) ExistByPrefix(ctx context.Context, prefix string) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardProduct.WithContext(ctx).Where(
		db.CardProduct.Channel.Eq(string(enums.Channel_PhotonPay)),
		db.CardProduct.Prefix.Eq(prefix),
	).Count()

	return count > 0, err
}

func (r *cardProductRepository) FindByPrefixForUpdate(ctx context.Context, prefix string) (*model.CardProduct, error) {
	db := r.repository.DB(ctx)

	return db.CardProduct.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where(
		db.CardProduct.Channel.Eq(string(enums.Channel_PhotonPay)),
		db.CardProduct.Prefix.Eq(prefix),
	).Order(db.CardProduct.ID.Desc()).First()
}

func (r *cardProductRepository) List(ctx context.Context) ([]*model.CardProduct, error) {
	db := r.repository.DB(ctx)

	return db.CardProduct.WithContext(ctx).Where(
		db.CardProduct.Channel.Eq(string(enums.Channel_PhotonPay)),
	).Order(db.CardProduct.ID.Desc()).Find()
}

func (r *cardProductRepository) ListByAccountID(ctx context.Context, accountID model.ID) ([]*model.CardProduct, error) {
	db := r.repository.DB(ctx)
	return db.CardProduct.WithContext(ctx).Where(
		db.CardProduct.AccountID.Eq(accountID),
		db.CardProduct.Channel.Eq(string(enums.Channel_PhotonPay)),
	).Order(db.CardProduct.ID.Desc()).Find()
}

func (r *cardProductRepository) Save(ctx context.Context, item *model.CardProduct) error {
	return r.repository.DB(ctx).CardProduct.WithContext(ctx).Save(item)
}

var _ biz.CardProductRepository = (*cardProductRepository)(nil)
