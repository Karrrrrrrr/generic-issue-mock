package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gorm/clause"
)

type cardProductRepository struct {
	*Repository
}

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
