package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/internal/query"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do/v2"
	"gorm.io/gen"
)

type cardTransactionRepository struct {
	repository *SlashRepository
}

func NewCardTransactionRepository(injector do.Injector) (biz.SlashCardTransactionRepository, error) {
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
	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Preload(db.CardTransaction.Authorization).
		Where(
			db.CardTransaction.ID.Eq(id),
			db.CardTransaction.Channel.Eq(string(enums.Channel_Slash)),
		).First()
}

func (r *cardTransactionRepository) Count(ctx context.Context, req *biz.CardTransactionCountRequest) (int64, error) {
	db := r.repository.DB(ctx)
	return db.CardTransaction.WithContext(ctx).
		Where(cardTransactionPredicates(db, (*biz.CardTransactionListRequest)(req))...).
		Count()
}

func (r *cardTransactionRepository) List(ctx context.Context, req *biz.CardTransactionListRequest) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Preload(db.CardTransaction.Authorization).
		Where(cardTransactionPredicates(db, req)...).
		Order(db.CardTransaction.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *cardTransactionRepository) ExistByAccountID(ctx context.Context, req *biz.CardTransactionExistByAccountIDRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(req.ID),
		db.CardTransaction.AccountID.Eq(*req.AccountID),
		db.CardTransaction.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
	return count > 0, err
}

func (r *cardTransactionRepository) FindByAccountID(ctx context.Context, req *biz.CardTransactionFindByAccountIDRequest) (*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Preload(db.CardTransaction.Authorization).
		Where(
			db.CardTransaction.ID.Eq(req.ID),
			db.CardTransaction.AccountID.Eq(*req.AccountID),
			db.CardTransaction.Channel.Eq(string(enums.Channel_Slash)),
		).First()
}

func cardTransactionPredicates(db *query.Query, req *biz.CardTransactionListRequest) []gen.Condition {
	predicates := make([]gen.Condition, 0, 6)
	predicates = append(predicates, db.CardTransaction.Channel.Eq(string(enums.Channel_Slash)))
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, db.CardTransaction.AccountID.In(req.AccountIDs...))
	}
	if len(req.IDs) != 0 {
		predicates = append(predicates, db.CardTransaction.ID.In(req.IDs...))
	}
	if len(req.CardIDs) != 0 {
		predicates = append(predicates, db.CardTransaction.CardID.In(req.CardIDs...))
	}
	if len(req.AuthorizationIDs) != 0 {
		predicates = append(predicates, db.CardTransaction.AuthorizationID.In(req.AuthorizationIDs...))
	}
	if len(req.Types) != 0 {
		predicates = append(predicates, db.CardTransaction.Type.In(types.BulkConvertSlice(req.Types, func(value enums.CardTransactionType) string {
			return string(value)
		})...))
	}
	if len(req.Statuses) != 0 {
		predicates = append(predicates, db.CardTransaction.Status.In(types.BulkConvertSlice(req.Statuses, func(value enums.CardTransactionStatus) string {
			return string(value)
		})...))
	}
	if req.CreatedFrom != nil {
		predicates = append(predicates, db.CardTransaction.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		predicates = append(predicates, db.CardTransaction.CreatedAt.Lte(*req.CreatedTo))
	}
	return predicates
}

var _ biz.SlashCardTransactionRepository = (*cardTransactionRepository)(nil)

func (r *cardTransactionRepository) ListStages(ctx context.Context, req *biz.ListAuthorizationStagesRequest) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Where(
			db.CardTransaction.AuthorizationID.Eq(req.ID),
			db.CardTransaction.AccountID.Eq(req.AccountID),
			db.CardTransaction.Channel.Eq(string(enums.Channel_Slash)),
		).Order(db.CardTransaction.ID.Desc()).Find()
}
