package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type cardProductRepository struct {
	repository *SlashRepository
}

func NewCardProductRepository(injector *do.Injector) (biz.CardProductRepository, error) {
	return &cardProductRepository{
		repository: do.MustInvoke[*SlashRepository](injector),
	}, nil
}

func (r *cardProductRepository) Create(ctx context.Context, item *model.CardProduct) error {
	return r.repository.DB(ctx).CardProduct.WithContext(ctx).Create(item)
}

func (r *cardProductRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardProduct.WithContext(ctx).Where(
		db.CardProduct.ID.Eq(id),
		db.CardProduct.Channel.Eq(string(enums.Channel_Slash)),
	).Count()

	return count > 0, err
}

func (r *cardProductRepository) FindByID(ctx context.Context, id model.ID) (*model.CardProduct, error) {
	db := r.repository.DB(ctx)

	return db.CardProduct.WithContext(ctx).Where(
		db.CardProduct.ID.Eq(id),
		db.CardProduct.Channel.Eq(string(enums.Channel_Slash)),
	).First()
}

func (r *cardProductRepository) ExistDefault(ctx context.Context) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardProduct.WithContext(ctx).Where(
		db.CardProduct.Channel.Eq(string(enums.Channel_Slash)),
		db.CardProduct.IsDefault.Is(true),
	).Count()

	return count > 0, err
}

func (r *cardProductRepository) FindDefault(ctx context.Context) (*model.CardProduct, error) {
	db := r.repository.DB(ctx)

	return db.CardProduct.WithContext(ctx).Where(
		db.CardProduct.Channel.Eq(string(enums.Channel_Slash)),
		db.CardProduct.IsDefault.Is(true),
	).Order(db.CardProduct.ID.Desc()).First()
}

func (r *cardProductRepository) List(ctx context.Context) ([]*model.CardProduct, error) {
	db := r.repository.DB(ctx)

	return db.CardProduct.WithContext(ctx).Where(
		db.CardProduct.Channel.Eq(string(enums.Channel_Slash)),
	).Order(db.CardProduct.ID.Desc()).Find()
}

var _ biz.CardProductRepository = (*cardProductRepository)(nil)
