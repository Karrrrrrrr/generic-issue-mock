package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/internal/query"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"gorm.io/gen"
)

type cardRepository struct {
	repository *SlashRepository
}

func NewCardRepository(injector *do.Injector) (biz.SlashCardRepository, error) {
	return &cardRepository{
		repository: do.MustInvoke[*SlashRepository](injector),
	}, nil
}

func (r *cardRepository) Create(ctx context.Context, item *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Create(item)
}

func (r *cardRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(id),
		db.Card.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
	return count > 0, err
}

func (r *cardRepository) FindByID(ctx context.Context, id model.ID) (*model.Card, error) {
	db := r.repository.DB(ctx)
	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Preload(db.Card.Wallet).
		Preload(db.Card.VirtualAccount.Wallet).
		Where(
			db.Card.ID.Eq(id),
			db.Card.Channel.Eq(string(enums.Channel_Slash)),
		).First()
}

func (r *cardRepository) Count(ctx context.Context, req *biz.CardCountRequest) (int64, error) {
	db := r.repository.DB(ctx)
	return db.Card.WithContext(ctx).
		Where(cardPredicates(db, (*biz.CardListRequest)(req))...).
		Count()
}

func (r *cardRepository) List(ctx context.Context, req *biz.CardListRequest) ([]*model.Card, error) {
	db := r.repository.DB(ctx)
	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Preload(db.Card.Wallet).
		Preload(db.Card.VirtualAccount.Wallet).
		Where(cardPredicates(db, req)...).
		Order(db.Card.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *cardRepository) Save(ctx context.Context, item *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Save(item)
}

func (r *cardRepository) ExistByAccountID(ctx context.Context, req *biz.CardExistByAccountIDRequest) (bool, error) {
	db := r.repository.DB(ctx)
	query := db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(req.ID),
		db.Card.Channel.Eq(string(enums.Channel_Slash)),
	)
	if req.AccountID != nil {
		query = query.Where(db.Card.AccountID.Eq(*req.AccountID))
	}
	count, err := query.Count()
	return count > 0, err
}

func (r *cardRepository) FindByAccountID(ctx context.Context, req *biz.CardFindByAccountIDRequest) (*model.Card, error) {
	db := r.repository.DB(ctx)
	query := db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.ID.Eq(req.ID),
			db.Card.Channel.Eq(string(enums.Channel_Slash)),
		)
	if req.AccountID != nil {
		query = query.Where(db.Card.AccountID.Eq(*req.AccountID))
	}
	return query.
		Preload(db.Card.Wallet).
		Preload(db.Card.VirtualAccount.Wallet).
		First()
}

func cardPredicates(db *query.Query, req *biz.CardListRequest) []gen.Condition {
	predicates := make([]gen.Condition, 0, 4)
	predicates = append(predicates, db.Card.Channel.Eq(string(enums.Channel_Slash)))
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, db.Card.AccountID.In(req.AccountIDs...))
	}
	if req.IDContains != nil {
		predicates = append(predicates, db.Card.ID.Like("%"+*req.IDContains+"%"))
	}
	if req.CardNumber != nil {
		predicates = append(predicates, db.Card.CardNumber.Like("%"+*req.CardNumber+"%"))
	}
	if len(req.Statuses) != 0 {
		predicates = append(predicates, db.Card.Status.In(types.BulkConvertSlice(req.Statuses, func(value enums.CardStatus) string {
			return string(value)
		})...))
	}
	return predicates
}

var _ biz.SlashCardRepository = (*cardRepository)(nil)

func (r *cardRepository) FindCard(ctx context.Context, req *biz.FindCardRequest) (*model.Card, error) {
	db := r.repository.DB(ctx)
	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.ID.Eq(req.ID),
			db.Card.AccountID.Eq(req.AccountID),
			db.Card.Channel.Eq(string(enums.Channel_Slash)),
		).First()
}
