package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

type authorizationRepository struct {
	*Repository
}

var _ biz.AuthorizationRepo = (*authorizationRepository)(nil)

func NewAuthorizationRepository(injector do.Injector) (biz.AuthorizationRepo, error) {
	return &authorizationRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *authorizationRepository) Create(ctx context.Context, req *biz.AuthorizationCreateRequest) error {
	return repo.DB(ctx).Authorization.WithContext(ctx).Create(req.Authorization)
}

func (repo *authorizationRepository) Save(ctx context.Context, req *biz.AuthorizationSaveRequest) error {
	return repo.DB(ctx).Authorization.WithContext(ctx).Save(req.Authorization)
}

func (repo *authorizationRepository) ExistForCard(ctx context.Context, req *biz.AuthorizationExistForCardRequest) (bool, error) {
	table := repo.DB(ctx).Authorization
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.CardID.Eq(req.CardID),
		).Count()
	return count > 0, err
}

func (repo *authorizationRepository) FindByIDWithLock(ctx context.Context, req *biz.AuthorizationFindByIDWithLockRequest) (*model.Authorization, error) {
	table := repo.DB(ctx).Authorization
	return table.WithContext(ctx).
		Preload(table.Account).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.CardID.Eq(req.CardID),
		).First()
}

func (repo *authorizationRepository) ExistForSimulation(ctx context.Context, req *biz.AuthorizationSimulationExistRequest) (bool, error) {
	table := repo.DB(ctx).Authorization
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(req.Channel)),
		).Count()
	return count > 0, err
}

func (repo *authorizationRepository) FindForSimulation(ctx context.Context, req *biz.AuthorizationSimulationFindRequest) (*model.Authorization, error) {
	table := repo.DB(ctx).Authorization
	return table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(req.Channel)),
		).First()
}

func (repo *authorizationRepository) List(ctx context.Context, req *biz.AuthorizationListRequest) ([]*model.Authorization, error) {
	db := repo.DB(ctx)
	table := db.Authorization
	query := table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.CardTransactions.
			Where(db.CardTransaction.Channel.Eq(string(req.Channel))).
			Order(db.CardTransaction.ID.Desc())).
		Where(repo.buildPredicates(ctx, &req.AuthorizationFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset)
	if req.Limit != nil {
		query = query.Limit(*req.Limit)
	}
	return query.Find()
}

func (repo *authorizationRepository) Count(ctx context.Context, req *biz.AuthorizationCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.Authorization.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.AuthorizationFilters)...).
		Count()
}

func (repo *authorizationRepository) buildPredicates(ctx context.Context, req *biz.AuthorizationFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.Authorization
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
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		predicates = append(predicates, table.Status.In(values...))
	}
	if req.MerchantName != nil {
		predicates = append(predicates, table.MerchantName.Like("%"+*req.MerchantName+"%"))
	}
	if req.CreatedFrom != nil {
		predicates = append(predicates, table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		predicates = append(predicates, table.CreatedAt.Lte(*req.CreatedTo))
	}
	return predicates
}

func (repo *authorizationRepository) Exist(ctx context.Context, req *biz.AuthorizationExistRequest) (bool, error) {
	table := repo.DB(ctx).Authorization
	count, err := table.WithContext(ctx).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.AccountID.Eq(req.AccountID),
			table.ID.Eq(req.ID),
		).
		Count()
	return count > 0, err
}

func (repo *authorizationRepository) Find(ctx context.Context, req *biz.AuthorizationFindRequest) (*model.Authorization, error) {
	db := repo.DB(ctx)
	table := db.Authorization
	return table.WithContext(ctx).
		Preload(table.Account).
		Preload(table.CardTransactions.
			Where(
				db.CardTransaction.AccountID.Eq(req.AccountID),
				db.CardTransaction.Channel.Eq(string(req.Channel)),
			).
			Order(db.CardTransaction.ID.Desc())).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.AccountID.Eq(req.AccountID),
			table.ID.Eq(req.ID),
		).
		First()
}
