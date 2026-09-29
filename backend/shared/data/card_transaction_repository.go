package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
)

type cardTransactionRepository struct {
	*Repository
}

var _ biz.CardTransactionRepo = (*cardTransactionRepository)(nil)

func NewCardTransactionRepository(injector do.Injector) (biz.CardTransactionRepo, error) {
	return &cardTransactionRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *cardTransactionRepository) Create(ctx context.Context, req *biz.CardTransactionCreateRequest) error {
	return repo.DB(ctx).CardTransaction.WithContext(ctx).Create(req.CardTransaction)
}

func (repo *cardTransactionRepository) ExistByRequestID(ctx context.Context, req *biz.CardTransactionExistByRequestIDRequest) (bool, error) {
	table := repo.DB(ctx).CardTransaction
	count, err := table.WithContext(ctx).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.RequestID.Eq(req.RequestID),
		).Count()
	return count > 0, err
}

func (repo *cardTransactionRepository) FindByRequestID(ctx context.Context, req *biz.CardTransactionFindByRequestIDRequest) (*model.CardTransaction, error) {
	table := repo.DB(ctx).CardTransaction
	return table.WithContext(ctx).
		Preload(table.Account).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.RequestID.Eq(req.RequestID),
		).
		Order(table.ID.Desc()).
		First()
}

func (repo *cardTransactionRepository) ListStages(ctx context.Context, req *biz.CardTransactionListStagesRequest) ([]*model.CardTransaction, error) {
	table := repo.DB(ctx).CardTransaction
	return table.WithContext(ctx).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.CardID.Eq(req.CardID),
			table.AuthorizationID.Eq(req.AuthorizationID),
		).
		Order(table.ID.Desc()).
		Find()
}

func (repo *cardTransactionRepository) List(ctx context.Context, req *biz.CardTransactionListRequest) ([]*model.CardTransaction, error) {
	db := repo.DB(ctx)
	table := db.CardTransaction
	query := table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.Authorization)
	query = query.
		Where(repo.buildPredicates(ctx, &req.CardTransactionFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset)
	if req.Limit != nil {
		query = query.Limit(*req.Limit)
	}
	return query.Find()
}

func (repo *cardTransactionRepository) Count(ctx context.Context, req *biz.CardTransactionCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.CardTransaction.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.CardTransactionFilters)...).
		Count()
}

func (repo *cardTransactionRepository) buildPredicates(ctx context.Context, req *biz.CardTransactionFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.CardTransaction
	predicates := []gen.Condition{table.Channel.Eq(string(req.Channel))}
	if len(req.IDs) != 0 {
		predicates = append(predicates, table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, table.AccountID.In(req.AccountIDs...))
	}
	if len(req.CardIDs) != 0 {
		predicates = append(predicates, table.CardID.In(req.CardIDs...))
	}
	if len(req.VirtualAccountIDs) != 0 {
		cardTable := db.Card
		subQuery := cardTable.WithContext(ctx).
			Select(cardTable.ID).
			Where(
				cardTable.Channel.Eq(string(req.Channel)),
				cardTable.VirtualAccountID.In(req.VirtualAccountIDs...),
			)
		if len(req.AccountIDs) != 0 {
			subQuery = subQuery.Where(cardTable.AccountID.In(req.AccountIDs...))
		}
		predicates = append(predicates, table.Columns(table.CardID).In(subQuery))
	}
	if len(req.AuthorizationIDs) != 0 {
		predicates = append(predicates, table.AuthorizationID.In(req.AuthorizationIDs...))
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
	if len(req.Types) != 0 {
		values := make([]string, 0, len(req.Types))
		for _, value := range req.Types {
			values = append(values, string(value))
		}
		predicates = append(predicates, table.Type.In(values...))
	}
	if req.CreatedFrom != nil {
		predicates = append(predicates, table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		predicates = append(predicates, table.CreatedAt.Lte(*req.CreatedTo))
	}
	return predicates
}

func (repo *cardTransactionRepository) Exist(ctx context.Context, req *biz.CardTransactionExistRequest) (bool, error) {
	table := repo.DB(ctx).CardTransaction
	count, err := table.WithContext(ctx).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.AccountID.Eq(req.AccountID),
			table.ID.Eq(req.ID),
		).
		Count()
	return count > 0, err
}

func (repo *cardTransactionRepository) Find(ctx context.Context, req *biz.CardTransactionFindRequest) (*model.CardTransaction, error) {
	table := repo.DB(ctx).CardTransaction
	return table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.Authorization).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.AccountID.Eq(req.AccountID),
			table.ID.Eq(req.ID),
		).
		First()
}

func (repo *cardTransactionRepository) ExistForSimulation(ctx context.Context, req *biz.CardTransactionExistForSimulationRequest) (bool, error) {
	table := repo.DB(ctx).CardTransaction
	count, err := table.WithContext(ctx).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		Count()
	return count > 0, err
}

func (repo *cardTransactionRepository) FindForSimulation(ctx context.Context, req *biz.CardTransactionFindForSimulationRequest) (*model.CardTransaction, error) {
	table := repo.DB(ctx).CardTransaction
	return table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.Authorization).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		First()
}
