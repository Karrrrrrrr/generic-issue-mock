package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type cardRepository struct {
	repository *PayndaRepository
}

var _ biz.PayndaCardRepository = (*cardRepository)(nil)

func NewCardRepository(injector *do.Injector) (biz.PayndaCardRepository, error) {
	return &cardRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *cardRepository) Create(ctx context.Context, item *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Create(item)
}

func (r *cardRepository) ExistByID(ctx context.Context, req *biz.CardExistByIDRequest) (bool, error) {
	db := r.repository.DB(ctx)
	query := db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(req.ID),
		db.Card.Channel.Eq(string(enums.Channel_Paynda)),
	)
	if req.AccountID != nil {
		query = query.Where(db.Card.AccountID.Eq(*req.AccountID))
	}
	count, err := query.Count()

	return count > 0, err
}

func (r *cardRepository) ExistByRequestID(ctx context.Context, req *biz.CardExistByRequestIDRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).
		Where(
			db.Card.RequestID.Eq(req.RequestID),
			db.Card.AccountID.Eq(req.AccountID), db.Card.Channel.Eq(string(enums.Channel_Paynda)),
		).
		Count()

	return count > 0, err
}

func (r *cardRepository) ExistByLastOperationRequestID(
	ctx context.Context,
	req *biz.CardExistByLastOperationRequestIDRequest,
) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).
		Where(
			db.Card.LastOperationRequestID.Eq(req.RequestID),
			db.Card.AccountID.Eq(req.AccountID), db.Card.Channel.Eq(string(enums.Channel_Paynda)),
		).
		Count()

	return count > 0, err
}

func (r *cardRepository) FindByID(ctx context.Context, req *biz.CardFindByIDRequest) (*model.Card, error) {
	db := r.repository.DB(ctx)
	query := db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.ID.Eq(req.ID),
			db.Card.Channel.Eq(string(enums.Channel_Paynda)),
		)
	if req.AccountID != nil {
		query = query.Where(db.Card.AccountID.Eq(*req.AccountID))
	}
	return query.
		Preload(db.Card.Wallet).
		First()
}

func (r *cardRepository) FindByRequestID(ctx context.Context, req *biz.CardFindByRequestIDRequest) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.RequestID.Eq(req.RequestID),
			db.Card.AccountID.Eq(req.AccountID),
			db.Card.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *cardRepository) FindByLastOperationRequestID(
	ctx context.Context,
	req *biz.CardFindByLastOperationRequestIDRequest,
) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.LastOperationRequestID.Eq(req.RequestID),
			db.Card.AccountID.Eq(req.AccountID),
			db.Card.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *cardRepository) List(ctx context.Context, req *biz.CardListRequest) ([]*model.Card, error) {
	db := r.repository.DB(ctx)

	query := db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(db.Card.Channel.Eq(string(enums.Channel_Paynda)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.Card.AccountID.In(req.AccountIDs...))
	}
	return query.
		Preload(db.Card.Wallet).
		Order(db.Card.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *cardRepository) Save(ctx context.Context, item *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Save(item)
}

func (r *cardRepository) FindCard(ctx context.Context, req *biz.FindCardRequest) (*model.Card, error) {
	db := r.repository.DB(ctx)
	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.ID.Eq(req.ID),
			db.Card.AccountID.Eq(req.AccountID),
			db.Card.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}
