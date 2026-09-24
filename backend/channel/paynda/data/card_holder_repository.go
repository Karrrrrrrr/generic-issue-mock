package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type cardHolderRepository struct {
	repository *PayndaRepository
}

var _ biz.PayndaCardHolderRepository = (*cardHolderRepository)(nil)

func NewCardHolderRepository(injector *do.Injector) (biz.PayndaCardHolderRepository, error) {
	return &cardHolderRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *cardHolderRepository) Create(ctx context.Context, item *model.CardHolder) error {
	return r.repository.DB(ctx).CardHolder.WithContext(ctx).Create(item)
}

func (r *cardHolderRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardHolder.WithContext(ctx).
		Where(
			db.CardHolder.ID.Eq(id),
			db.CardHolder.Channel.Eq(string(enums.Channel_Paynda)),
		).
		Count()

	return count > 0, err
}

func (r *cardHolderRepository) FindByID(ctx context.Context, id model.ID) (*model.CardHolder, error) {
	db := r.repository.DB(ctx)

	return db.CardHolder.WithContext(ctx).
		Preload(db.CardHolder.Account).
		Where(
			db.CardHolder.ID.Eq(id),
			db.CardHolder.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *cardHolderRepository) ExistByAccountID(
	ctx context.Context,
	req *biz.CardHolderExistByAccountIDRequest,
) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardHolder.WithContext(ctx).Where(
		db.CardHolder.ID.Eq(req.ID),
		db.CardHolder.AccountID.Eq(req.AccountID),
		db.CardHolder.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *cardHolderRepository) FindByAccountID(
	ctx context.Context,
	req *biz.CardHolderFindByAccountIDRequest,
) (*model.CardHolder, error) {
	db := r.repository.DB(ctx)

	return db.CardHolder.WithContext(ctx).
		Preload(db.CardHolder.Account).
		Where(
			db.CardHolder.ID.Eq(req.ID),
			db.CardHolder.AccountID.Eq(req.AccountID),
			db.CardHolder.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *cardHolderRepository) List(
	ctx context.Context,
	req *biz.CardHolderListRequest,
) ([]*model.CardHolder, error) {
	db := r.repository.DB(ctx)

	query := db.CardHolder.WithContext(ctx).
		Preload(db.CardHolder.Account).
		Where(db.CardHolder.Channel.Eq(string(enums.Channel_Paynda)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.CardHolder.AccountID.In(req.AccountIDs...))
	}
	return query.
		Order(db.CardHolder.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *cardHolderRepository) Save(ctx context.Context, item *model.CardHolder) error {
	return r.repository.DB(ctx).CardHolder.WithContext(ctx).Save(item)
}

func (r *cardHolderRepository) Count(
	ctx context.Context,
	req *biz.CardHolderCountRequest,
) (int64, error) {
	db := r.repository.DB(ctx)
	query := db.CardHolder.WithContext(ctx).
		Where(db.CardHolder.Channel.Eq(string(enums.Channel_Paynda)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.CardHolder.AccountID.In(req.AccountIDs...))
	}

	return query.Count()
}
