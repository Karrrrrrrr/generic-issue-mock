package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
	"gorm.io/gorm/clause"
)

type cardRepository struct {
	repository *Repository
}

func NewCardRepository(injector do.Injector) (biz.CardRepository, error) {
	return &cardRepository{
		repository: do.MustInvoke[*Repository](injector),
	}, nil
}

func (r *cardRepository) CreateCard(ctx context.Context, card *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Create(card)
}

func (r *cardRepository) ExistCardByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).
		Where(
			db.Card.ID.Eq(id),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		Count()

	return count > 0, err
}

func (r *cardRepository) FindCardByID(ctx context.Context, id model.ID) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Preload(db.Card.Wallet).
		Preload(db.Card.VirtualAccount.Wallet).
		Where(
			db.Card.ID.Eq(id),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		First()
}

func (r *cardRepository) ExistCardByRequestID(ctx context.Context, requestID string) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).
		Where(
			db.Card.RequestID.Eq(requestID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		Count()

	return count > 0, err
}

func (r *cardRepository) FindByRequestID(ctx context.Context, requestID string) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.RequestID.Eq(requestID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		First()
}

func (r *cardRepository) ExistCardByLastOperationRequestID(ctx context.Context, requestID string) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).
		Where(
			db.Card.LastOperationRequestID.Eq(requestID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		Count()

	return count > 0, err
}

func (r *cardRepository) FindByLastOperationRequestID(ctx context.Context, requestID string) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.LastOperationRequestID.Eq(requestID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		First()
}

func (r *cardRepository) ExistCardByRequestIDForAccount(
	ctx context.Context,
	req *biz.CardExistCardByRequestIDForAccountRequest,
) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).Where(
		db.Card.AccountID.Eq(req.AccountID),
		db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		db.Card.RequestID.Eq(req.RequestID),
	).Count()

	return count > 0, err
}

func (r *cardRepository) FindCardByRequestIDForAccount(
	ctx context.Context,
	req *biz.CardFindCardByRequestIDForAccountRequest,
) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.AccountID.Eq(req.AccountID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
			db.Card.RequestID.Eq(req.RequestID),
		).First()
}

func (r *cardRepository) ExistCardByLastOperationRequestIDForAccount(
	ctx context.Context,
	req *biz.CardExistCardByLastOperationRequestIDForAccountRequest,
) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).Where(
		db.Card.AccountID.Eq(req.AccountID),
		db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		db.Card.LastOperationRequestID.Eq(req.RequestID),
	).Count()

	return count > 0, err
}

func (r *cardRepository) FindCardByLastOperationRequestIDForAccount(
	ctx context.Context,
	req *biz.CardFindCardByLastOperationRequestIDForAccountRequest,
) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.AccountID.Eq(req.AccountID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
			db.Card.LastOperationRequestID.Eq(req.RequestID),
		).First()
}

func (r *cardRepository) ListCards(ctx context.Context, req *biz.CardListCardsRequest) ([]*model.Card, error) {
	db := r.repository.DB(ctx)
	query := db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(db.Card.Channel.Eq(string(enums.Channel_PhotonPay)))
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
		Preload(db.Card.VirtualAccount.Wallet).
		Order(db.Card.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *cardRepository) SaveCard(ctx context.Context, card *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Save(card)
}

func (r *cardRepository) ExistCardByAccountID(ctx context.Context, req *biz.CardExistCardByAccountIDRequest) (bool, error) {
	db := r.repository.DB(ctx)
	query := db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(req.ID),
		db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
	)
	if req.AccountID != nil {
		query = query.Where(db.Card.AccountID.Eq(*req.AccountID))
	}
	count, err := query.Count()
	return count > 0, err
}

func (r *cardRepository) FindCardByAccountID(ctx context.Context, req *biz.CardFindCardByAccountIDRequest) (*model.Card, error) {
	db := r.repository.DB(ctx)
	query := db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.ID.Eq(req.ID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		)
	if req.AccountID != nil {
		query = query.Where(db.Card.AccountID.Eq(*req.AccountID))
	}
	return query.First()
}

var _ biz.CardRepository = (*cardRepository)(nil)

func (r *cardRepository) FindCard(ctx context.Context, req *biz.FindCardRequest) (*model.Card, error) {
	db := r.repository.DB(ctx)
	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Where(
			db.Card.ID.Eq(req.ID),
			db.Card.AccountID.Eq(req.AccountID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).First()
}

func (r *cardRepository) Count(
	ctx context.Context,
	req *biz.CardCountRequest,
) (int64, error) {
	db := r.repository.DB(ctx)
	query := db.Card.WithContext(ctx).
		Where(db.Card.Channel.Eq(string(enums.Channel_PhotonPay)))
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
		db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
	).Count()
	return count > 0, err
}

func (r *cardRepository) LockForStatusChange(ctx context.Context, req *biz.CardStatusLockRequest) (*model.Card, error) {
	db := r.repository.DB(ctx)
	return db.Card.WithContext(ctx).
		Preload(db.Card.Account).
		Preload(db.Card.Wallet).
		Preload(db.Card.VirtualAccount.Wallet).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).
		Where(
			db.Card.ID.Eq(req.ID),
			db.Card.AccountID.Eq(req.AccountID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).First()
}

func (r *cardRepository) SaveStatus(ctx context.Context, req *biz.CardStatusSaveRequest) error {
	db := r.repository.DB(ctx)
	_, err := db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(req.ID),
		db.Card.AccountID.Eq(req.AccountID),
		db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
	).UpdateSimple(
		db.Card.Status.Value(string(req.Status)),
		db.Card.LastOperationRequestID.Value(req.LastOperationRequestID),
		db.Card.LastOperationType.Value(string(req.LastOperationType)),
		db.Card.LastOperationStatus.Value(string(req.LastOperationStatus)),
	)
	return err
}

func (r *cardRepository) ListForFunding(ctx context.Context, req *biz.CardListForFundingRequest) ([]*model.Card, error) {
	table := r.repository.DB(ctx).Card
	query := table.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(enums.Channel_PhotonPay)),
		)
	if len(req.WalletIDs) != 0 {
		query = query.Where(table.WalletID.In(req.WalletIDs...))
	}
	return query.Order(table.ID.Desc()).Find()
}

func (r *cardRepository) UpdateOperation(ctx context.Context, req *biz.CardOperationUpdateRequest) error {
	db := r.repository.DB(ctx)
	_, err := db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(req.ID),
		db.Card.AccountID.Eq(req.AccountID),
		db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
	).UpdateSimple(
		db.Card.LastOperationRequestID.Value(req.RequestID),
		db.Card.LastOperationType.Value(string(req.Type)),
		db.Card.LastOperationStatus.Value(string(req.Status)),
	)
	return err
}
