package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type productRepository struct{ *PingPongRepository }

func NewProductRepository(injector *do.Injector) (biz.PingPongProductRepository, error) {
	return &productRepository{PingPongRepository: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (repo *productRepository) Exists(ctx context.Context, req *biz.ProductExistsRequest) (bool, error) {
	table := repo.DB(ctx).CardProduct
	count, err := repo.DB(ctx).CardProduct.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
		).Count()
	return count > 0, err
}

func (repo *productRepository) Lock(ctx context.Context, req *biz.ProductLockRequest) (*model.CardProduct, error) {
	table := repo.DB(ctx).CardProduct
	return repo.DB(ctx).CardProduct.WithContext(ctx).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
		).First()
}

func (repo *productRepository) Save(ctx context.Context, item *model.CardProduct) error {
	table := repo.DB(ctx).CardProduct
	_, err := table.WithContext(ctx).Where(
		table.ID.Eq(item.ID),
		table.Channel.Eq(string(common.Channel_PingPong)),
	).UpdateSimple(
		table.NextCardNumber.Value(item.NextCardNumber),
	)
	return err
}

func (repo *productRepository) List(ctx context.Context, req *biz.ProductListRequest) ([]*model.CardProduct, error) {
	table := repo.DB(ctx).CardProduct
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	if req.Limit != nil {
		statement = statement.Limit(*req.Limit)
	}
	return statement.Order(table.ID.Desc()).Offset(req.Offset).Find()
}
