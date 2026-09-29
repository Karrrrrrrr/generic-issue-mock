package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

type cardRepository struct {
	*Repository
}

var _ biz.CardRepo = (*cardRepository)(nil)

func NewCardRepository(injector do.Injector) (biz.CardRepo, error) {
	return &cardRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *cardRepository) Create(ctx context.Context, req *biz.CardCreateRequest) error {
	return repo.DB(ctx).Card.WithContext(ctx).Create(req.Card)
}

func (repo *cardRepository) Exist(ctx context.Context, req *biz.CardExistRequest) (bool, error) {
	table := repo.DB(ctx).Card
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).Count()
	return count > 0, err
}

func (repo *cardRepository) ExistByRequestID(ctx context.Context, req *biz.CardExistByRequestIDRequest) (bool, error) {
	table := repo.DB(ctx).Card
	count, err := table.WithContext(ctx).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.RequestID.Eq(req.RequestID),
		).Count()
	return count > 0, err
}

func (repo *cardRepository) FindByRequestID(ctx context.Context, req *biz.CardFindByRequestIDRequest) (*model.Card, error) {
	table := repo.DB(ctx).Card
	return table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.Wallet).
		Preload(table.VirtualAccount.Wallet).
		Preload(table.CardHolder).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.RequestID.Eq(req.RequestID),
		).
		Order(table.ID.Desc()).
		First()
}

func (repo *cardRepository) FindByIDWithLock(ctx context.Context, req *biz.CardFindByIDWithLockRequest) (*model.Card, error) {
	db := repo.DB(ctx)
	table := db.Card
	return table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.VirtualAccount).
		Preload(table.Wallet.Where(
			db.Wallet.AccountID.Eq(req.AccountID),
			db.Wallet.Channel.Eq(string(req.Channel)),
		)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).First()
}

func (repo *cardRepository) List(ctx context.Context, req *biz.CardListRequest) ([]*model.Card, error) {
	db := repo.DB(ctx)
	table := db.Card
	query := table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.Wallet).
		Preload(table.VirtualAccount.Wallet).
		Preload(table.CardHolder).
		Where(repo.buildPredicates(ctx, &req.CardFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset)
	if req.Limit != nil {
		query = query.Limit(*req.Limit)
	}
	return query.Find()
}

func (repo *cardRepository) Count(ctx context.Context, req *biz.CardCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.Card.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.CardFilters)...).
		Count()
}

func (repo *cardRepository) buildPredicates(ctx context.Context, req *biz.CardFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.Card
	predicates := []gen.Condition{table.Channel.Eq(string(req.Channel))}
	if len(req.IDs) != 0 {
		predicates = append(predicates, table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, table.AccountID.In(req.AccountIDs...))
	}
	if len(req.VirtualAccountIDs) != 0 {
		predicates = append(predicates, table.VirtualAccountID.In(req.VirtualAccountIDs...))
	}
	if len(req.RequestIDs) != 0 {
		predicates = append(predicates, table.RequestID.In(req.RequestIDs...))
	}
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		predicates = append(predicates, table.Status.In(values...))
	}
	if req.CardNumber != nil {
		predicates = append(predicates, table.CardNumber.Like("%"+*req.CardNumber+"%"))
	}
	if req.CreatedFrom != nil {
		predicates = append(predicates, table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		predicates = append(predicates, table.CreatedAt.Lte(*req.CreatedTo))
	}
	return predicates
}

func (repo *cardRepository) Find(ctx context.Context, req *biz.CardFindRequest) (*model.Card, error) {
	table := repo.DB(ctx).Card
	return table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.Wallet).
		Preload(table.VirtualAccount).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.AccountID.Eq(req.AccountID),
			table.ID.Eq(req.ID),
		).
		First()
}

func (repo *cardRepository) UpdateStatus(ctx context.Context, req *biz.CardUpdateStatusRequest) error {
	table := repo.DB(ctx).Card
	_, err := table.WithContext(ctx).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.AccountID.Eq(req.AccountID),
			table.ID.Eq(req.ID),
		).
		UpdateSimple(table.Status.Value(string(req.Status)))
	return err
}

func (repo *cardRepository) ExistForSimulation(ctx context.Context, req *biz.CardSimulationExistRequest) (bool, error) {
	table := repo.DB(ctx).Card
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(req.Channel)),
		).Count()
	return count > 0, err
}

func (repo *cardRepository) FindForSimulation(ctx context.Context, req *biz.CardSimulationFindRequest) (*model.Card, error) {
	table := repo.DB(ctx).Card
	return table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(req.Channel)),
		).First()
}
