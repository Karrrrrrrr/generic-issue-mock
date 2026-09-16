package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type cardHolderRepository struct {
	repository *SlashRepository
}

func NewCardHolderRepository(injector *do.Injector) (biz.CardHolderRepository, error) {
	return &cardHolderRepository{
		repository: do.MustInvoke[*SlashRepository](injector),
	}, nil
}

func (r *cardHolderRepository) Create(ctx context.Context, item *model.CardHolder) error {
	return r.repository.DB(ctx).CardHolder.WithContext(ctx).Create(item)
}

func (r *cardHolderRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardHolder.WithContext(ctx).Where(
		db.CardHolder.ID.Eq(id),
		db.CardHolder.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
	return count > 0, err
}

func (r *cardHolderRepository) FindByID(ctx context.Context, id model.ID) (*model.CardHolder, error) {
	db := r.repository.DB(ctx)
	return db.CardHolder.WithContext(ctx).Where(
		db.CardHolder.ID.Eq(id),
		db.CardHolder.Channel.Eq(string(enums.Channel_Slash)),
	).First()
}

func (r *cardHolderRepository) Count(ctx context.Context, _ *biz.ListCardHoldersRequest) (int64, error) {
	db := r.repository.DB(ctx)
	return db.CardHolder.WithContext(ctx).Where(db.CardHolder.Channel.Eq(string(enums.Channel_Slash))).Count()
}

func (r *cardHolderRepository) List(ctx context.Context, req *biz.ListCardHoldersRequest) ([]*model.CardHolder, error) {
	db := r.repository.DB(ctx)
	return db.CardHolder.WithContext(ctx).Where(
		db.CardHolder.Channel.Eq(string(enums.Channel_Slash)),
	).Order(db.CardHolder.ID.Desc()).Offset(req.Offset).Limit(req.Limit).Find()
}

func (r *cardHolderRepository) Save(ctx context.Context, item *model.CardHolder) error {
	return r.repository.DB(ctx).CardHolder.WithContext(ctx).Save(item)
}

var _ biz.CardHolderRepository = (*cardHolderRepository)(nil)
