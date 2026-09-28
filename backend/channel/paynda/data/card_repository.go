package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
	"gorm.io/gorm/clause"
)

type cardRepository struct {
	repository *PayndaRepository
}

var _ biz.PayndaCardRepository = (*cardRepository)(nil)

func NewCardRepository(injector do.Injector) (biz.PayndaCardRepository, error) {
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
	if len(req.IDs) != 0 {
		query = query.Where(db.Card.ID.In(req.IDs...))
	}
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		query = query.Where(db.Card.Status.In(values...))
	}
	if req.CardNumber != nil {
		query = query.Where(db.Card.CardNumber.Like("%" + *req.CardNumber + "%"))
	}
	if req.CreatedFrom != nil {
		query = query.Where(db.Card.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		query = query.Where(db.Card.CreatedAt.Lte(*req.CreatedTo))
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

func (r *cardRepository) Count(
	ctx context.Context,
	req *biz.CardCountRequest,
) (int64, error) {
	db := r.repository.DB(ctx)
	query := db.Card.WithContext(ctx).
		Where(db.Card.Channel.Eq(string(enums.Channel_Paynda)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.Card.AccountID.In(req.AccountIDs...))
	}

	if len(req.IDs) != 0 {
		query = query.Where(db.Card.ID.In(req.IDs...))
	}
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		query = query.Where(db.Card.Status.In(values...))
	}
	if req.CardNumber != nil {
		query = query.Where(db.Card.CardNumber.Like("%" + *req.CardNumber + "%"))
	}
	if req.CreatedFrom != nil {
		query = query.Where(db.Card.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		query = query.Where(db.Card.CreatedAt.Lte(*req.CreatedTo))
	}
	return query.Count()
}

func (r *cardRepository) ExistForStatusChange(ctx context.Context, req *biz.CardStatusExistsRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(req.ID),
		db.Card.AccountID.Eq(req.AccountID),
		db.Card.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()
	return count > 0, err
}

func (r *cardRepository) LockForStatusChange(ctx context.Context, req *biz.CardStatusLockRequest) (*model.Card, error) {
	db := r.repository.DB(ctx)
	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Preload(db.Card.Wallet).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).
		Where(
			db.Card.ID.Eq(req.ID),
			db.Card.AccountID.Eq(req.AccountID),
			db.Card.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *cardRepository) SaveStatus(ctx context.Context, req *biz.CardStatusSaveRequest) error {
	db := r.repository.DB(ctx)
	_, err := db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(req.ID),
		db.Card.AccountID.Eq(req.AccountID),
		db.Card.Channel.Eq(string(enums.Channel_Paynda)),
	).UpdateSimple(
		db.Card.Status.Value(string(req.Status)),
		db.Card.LastOperationRequestID.Value(req.LastOperationRequestID),
		db.Card.LastOperationType.Value(string(req.LastOperationType)),
		db.Card.LastOperationStatus.Value(string(req.LastOperationStatus)),
	)
	return err
}
