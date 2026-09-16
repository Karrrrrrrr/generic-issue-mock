package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/internal/query"
	"generic-mock/model"

	"github.com/samber/do"
	"gorm.io/gen"
)

type cardTransactionRepository struct {
	repository *SlashRepository
}

func NewCardTransactionRepository(injector *do.Injector) (biz.CardTransactionRepository, error) {
	return &cardTransactionRepository{
		repository: do.MustInvoke[*SlashRepository](injector),
	}, nil
}

func (r *cardTransactionRepository) Create(ctx context.Context, item *model.CardTransaction) error {
	return r.repository.DB(ctx).CardTransaction.WithContext(ctx).Create(item)
}

func (r *cardTransactionRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(id),
		db.CardTransaction.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
	return count > 0, err
}

func (r *cardTransactionRepository) FindByID(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	return db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(id),
		db.CardTransaction.Channel.Eq(string(enums.Channel_Slash)),
	).First()
}

func (r *cardTransactionRepository) Count(ctx context.Context, req *biz.ListCardTransactionsRequest) (int64, error) {
	db := r.repository.DB(ctx)
	return db.CardTransaction.WithContext(ctx).Where(cardTransactionPredicates(db, req)...).Count()
}

func (r *cardTransactionRepository) List(ctx context.Context, req *biz.ListCardTransactionsRequest) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	return db.CardTransaction.WithContext(ctx).Where(cardTransactionPredicates(db, req)...).Order(db.CardTransaction.ID.Desc()).Offset(req.Offset).Limit(req.Limit).Find()
}

func cardTransactionPredicates(db *query.Query, req *biz.ListCardTransactionsRequest) []gen.Condition {
	predicates := make([]gen.Condition, 0, 6)
	predicates = append(predicates, db.CardTransaction.Channel.Eq(string(enums.Channel_Slash)))
	if req.ID != "" {
		predicates = append(predicates, db.CardTransaction.ID.Eq(req.ID))
	}
	if req.CardID != "" {
		predicates = append(predicates, db.CardTransaction.CardID.Eq(req.CardID))
	}
	if req.AuthorizationID != "" {
		predicates = append(predicates, db.CardTransaction.AuthorizationID.Eq(req.AuthorizationID))
	}
	if req.Type != "" {
		predicates = append(predicates, db.CardTransaction.Type.Eq(string(req.Type)))
	}
	if req.Status != "" {
		predicates = append(predicates, db.CardTransaction.Status.Eq(string(req.Status)))
	}
	return predicates
}

var _ biz.CardTransactionRepository = (*cardTransactionRepository)(nil)
