package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
)

type cardHolderRepository struct {
	*Repository
}

var _ biz.CardHolderRepo = (*cardHolderRepository)(nil)

func NewCardHolderRepository(injector do.Injector) (biz.CardHolderRepo, error) {
	return &cardHolderRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *cardHolderRepository) Create(ctx context.Context, req *biz.CardHolderCreateRequest) error {
	return repo.DB(ctx).CardHolder.WithContext(ctx).Create(req.CardHolder)
}

func (repo *cardHolderRepository) Exist(ctx context.Context, req *biz.CardHolderExistRequest) (bool, error) {
	table := repo.DB(ctx).CardHolder
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
		).Count()
	return count > 0, err
}

func (repo *cardHolderRepository) List(ctx context.Context, req *biz.CardHolderListRequest) ([]*model.CardHolder, error) {
	db := repo.DB(ctx)
	table := db.CardHolder
	query := table.WithContext(ctx).
		Preload(table.Account).
		Where(repo.buildPredicates(ctx, &req.CardHolderFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset)
	if req.Limit != nil {
		query = query.Limit(*req.Limit)
	}
	return query.Find()
}

func (repo *cardHolderRepository) Count(ctx context.Context, req *biz.CardHolderCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.CardHolder.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.CardHolderFilters)...).
		Count()
}

func (repo *cardHolderRepository) buildPredicates(ctx context.Context, req *biz.CardHolderFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.CardHolder
	predicates := []gen.Condition{table.Channel.Eq(string(req.Channel))}
	if len(req.IDs) != 0 {
		predicates = append(predicates, table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, table.AccountID.In(req.AccountIDs...))
	}
	return predicates
}
