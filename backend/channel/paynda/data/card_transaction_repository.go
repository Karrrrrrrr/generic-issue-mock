package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	"generic-mock/internal/query"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"gorm.io/gen"
)

type cardTransactionRepository struct {
	repository *PayndaRepository
}

var _ biz.PayndaCardTransactionRepository = (*cardTransactionRepository)(nil)

func NewCardTransactionRepository(injector *do.Injector) (biz.PayndaCardTransactionRepository, error) {
	return &cardTransactionRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *cardTransactionRepository) Create(ctx context.Context, item *model.CardTransaction) error {
	return r.repository.DB(ctx).CardTransaction.WithContext(ctx).Create(item)
}

func (r *cardTransactionRepository) Save(ctx context.Context, item *model.CardTransaction) error {
	return r.repository.DB(ctx).CardTransaction.WithContext(ctx).Save(item)
}

func (r *cardTransactionRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).
		Where(
			db.CardTransaction.ID.Eq(id),
			db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
		).
		Count()

	return count > 0, err
}

func (r *cardTransactionRepository) ExistByRequestID(ctx context.Context, req *biz.CardTransactionExistByRequestIDRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).
		Where(
			db.CardTransaction.RequestID.Eq(req.RequestID),
			db.CardTransaction.AccountID.Eq(req.AccountID), db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
		).
		Count()

	return count > 0, err
}

func (r *cardTransactionRepository) FindByID(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	db := r.repository.DB(ctx)

	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Where(
			db.CardTransaction.ID.Eq(id),
			db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *cardTransactionRepository) ExistByAccountID(
	ctx context.Context,
	req *biz.CardTransactionExistByAccountIDRequest,
) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(req.ID),
		db.CardTransaction.AccountID.Eq(req.AccountID),
		db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *cardTransactionRepository) FindByAccountID(
	ctx context.Context,
	req *biz.CardTransactionFindByAccountIDRequest,
) (*model.CardTransaction, error) {
	db := r.repository.DB(ctx)

	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Where(
			db.CardTransaction.ID.Eq(req.ID),
			db.CardTransaction.AccountID.Eq(req.AccountID),
			db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *cardTransactionRepository) FindByRequestID(
	ctx context.Context,
	req *biz.CardTransactionFindByRequestIDRequest,
) (*model.CardTransaction, error) {
	db := r.repository.DB(ctx)

	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Where(
			db.CardTransaction.RequestID.Eq(req.RequestID),
			db.CardTransaction.AccountID.Eq(req.AccountID),
			db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *cardTransactionRepository) List(
	ctx context.Context,
	req *biz.CardTransactionListRequest,
) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	predicates := payndaTransactionPredicates(ctx, db, req)

	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Where(predicates...).
		Order(db.CardTransaction.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func payndaTransactionPredicates(
	ctx context.Context,
	db *query.Query,
	req *biz.CardTransactionListRequest,
) []gen.Condition {
	predicates := []gen.Condition{db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda))}
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, db.CardTransaction.AccountID.In(req.AccountIDs...))
	}
	if len(req.CardIDs) != 0 {
		predicates = append(predicates, db.CardTransaction.CardID.In(req.CardIDs...))
	}
	if req.StartCreatedAt != nil {
		predicates = append(predicates, db.CardTransaction.CreatedAt.Gte(*req.StartCreatedAt))
	}
	if req.EndCreatedAt != nil {
		predicates = append(predicates, db.CardTransaction.CreatedAt.Lte(*req.EndCreatedAt))
	}
	if len(req.Types) != 0 {
		predicates = append(predicates, db.CardTransaction.Type.In(
			types.BulkConvertSlice(req.Types, func(value enums.CardTransactionType) string {
				return string(value)
			})...,
		))
	}

	return predicates
}

func (r *cardTransactionRepository) ListStages(ctx context.Context, req *biz.ListAuthorizationStagesRequest) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Where(
			db.CardTransaction.AuthorizationID.Eq(req.ID),
			db.CardTransaction.AccountID.Eq(req.AccountID),
			db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
		).Order(db.CardTransaction.ID.Desc()).Find()
}
