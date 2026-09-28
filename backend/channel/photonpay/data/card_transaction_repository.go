package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

type cardTransactionRepository struct {
	repository *Repository
}

func NewCardTransactionRepository(injector do.Injector) (biz.CardTransactionRepository, error) {
	return &cardTransactionRepository{
		repository: do.MustInvoke[*Repository](injector),
	}, nil
}

func (r *cardTransactionRepository) Create(ctx context.Context, item *model.CardTransaction) error {
	return r.repository.DB(ctx).CardTransaction.WithContext(ctx).Create(item)
}

func (r *cardTransactionRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(id),
		db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)),
	).Count()

	return count > 0, err
}

func (r *cardTransactionRepository) FindByID(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	db := r.repository.DB(ctx)

	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Where(
			db.CardTransaction.ID.Eq(id),
			db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)),
		).First()
}

func (r *cardTransactionRepository) ExistByAccountID(
	ctx context.Context,
	req *biz.CardTransactionExistByAccountIDRequest,
) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(req.ID),
		db.CardTransaction.AccountID.Eq(*req.AccountID),
		db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)),
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
			db.CardTransaction.AccountID.Eq(*req.AccountID),
			db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)),
		).First()
}

func (r *cardTransactionRepository) ListTransactions(ctx context.Context, req *biz.CardTransactionListTransactionsRequest) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	query := db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Where(db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.CardTransaction.AccountID.In(req.AccountIDs...))
	}
	if len(req.IDs) != 0 {
		query = query.Where(db.CardTransaction.ID.In(req.IDs...))
	}
	if len(req.CardIDs) != 0 {
		query = query.Where(db.CardTransaction.CardID.In(req.CardIDs...))
	}
	if len(req.AuthorizationIDs) != 0 {
		query = query.Where(db.CardTransaction.AuthorizationID.In(req.AuthorizationIDs...))
	}
	if len(req.Types) != 0 {
		values := make([]string, 0, len(req.Types))
		for _, value := range req.Types {
			values = append(values, string(value))
		}
		query = query.Where(db.CardTransaction.Type.In(values...))
	}
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		query = query.Where(db.CardTransaction.Status.In(values...))
	}
	if req.CreatedFrom != nil {
		query = query.Where(db.CardTransaction.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		query = query.Where(db.CardTransaction.CreatedAt.Lte(*req.CreatedTo))
	}
	return query.
		Order(db.CardTransaction.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

var _ biz.CardTransactionRepository = (*cardTransactionRepository)(nil)

func (r *cardTransactionRepository) ListStages(ctx context.Context, req *biz.ListAuthorizationStagesRequest) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	return db.CardTransaction.WithContext(ctx).
		Preload(db.CardTransaction.Account).
		Where(
			db.CardTransaction.AuthorizationID.Eq(req.ID),
			db.CardTransaction.AccountID.Eq(req.AccountID),
			db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)),
		).Order(db.CardTransaction.ID.Desc()).Find()
}

func (r *cardTransactionRepository) Count(
	ctx context.Context,
	req *biz.CardTransactionCountRequest,
) (int64, error) {
	db := r.repository.DB(ctx)
	query := db.CardTransaction.WithContext(ctx).
		Where(db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.CardTransaction.AccountID.In(req.AccountIDs...))
	}

	if len(req.IDs) != 0 {
		query = query.Where(db.CardTransaction.ID.In(req.IDs...))
	}
	if len(req.CardIDs) != 0 {
		query = query.Where(db.CardTransaction.CardID.In(req.CardIDs...))
	}
	if len(req.AuthorizationIDs) != 0 {
		query = query.Where(db.CardTransaction.AuthorizationID.In(req.AuthorizationIDs...))
	}
	if len(req.Types) != 0 {
		values := make([]string, 0, len(req.Types))
		for _, value := range req.Types {
			values = append(values, string(value))
		}
		query = query.Where(db.CardTransaction.Type.In(values...))
	}
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		query = query.Where(db.CardTransaction.Status.In(values...))
	}
	if req.CreatedFrom != nil {
		query = query.Where(db.CardTransaction.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		query = query.Where(db.CardTransaction.CreatedAt.Lte(*req.CreatedTo))
	}
	return query.Count()
}
