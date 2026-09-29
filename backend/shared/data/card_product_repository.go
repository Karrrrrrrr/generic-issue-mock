package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

type cardProductRepository struct {
	*Repository
}

var _ biz.CardProductRepo = (*cardProductRepository)(nil)

func NewCardProductRepository(injector do.Injector) (biz.CardProductRepo, error) {
	return &cardProductRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *cardProductRepository) Exist(ctx context.Context, req *biz.CardProductExistRequest) (bool, error) {
	table := repo.DB(ctx).CardProduct
	count, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(req.Channel)),
		).Count()
	return count > 0, err
}

func (repo *cardProductRepository) FindByIDWithLock(ctx context.Context, req *biz.CardProductFindByIDWithLockRequest) (*model.CardProduct, error) {
	table := repo.DB(ctx).CardProduct
	return table.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(req.Channel)),
		).First()
}

func (repo *cardProductRepository) UpdateSeq(ctx context.Context, req *biz.CardProductUpdateSeqRequest) error {
	table := repo.DB(ctx).CardProduct
	_, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(req.Channel)),
		).
		UpdateSimple(table.NextCardNumber.Value(req.NextSequence))
	return err
}

func (repo *cardProductRepository) List(ctx context.Context, req *biz.CardProductListRequest) ([]*model.CardProduct, error) {
	db := repo.DB(ctx)
	table := db.CardProduct
	query := table.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.CardProductFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset)
	if req.Limit != nil {
		query = query.Limit(*req.Limit)
	}
	return query.Find()
}

func (repo *cardProductRepository) Count(ctx context.Context, req *biz.CardProductCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.CardProduct.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.CardProductFilters)...).
		Count()
}

func (repo *cardProductRepository) buildPredicates(ctx context.Context, req *biz.CardProductFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.CardProduct
	predicates := []gen.Condition{table.Channel.Eq(string(req.Channel))}
	if len(req.IDs) != 0 {
		predicates = append(predicates, table.ID.In(req.IDs...))
	}
	return predicates
}
