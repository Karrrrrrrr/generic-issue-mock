package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do"
)

type cardHolderRepository struct {
	*Repository
}

func NewCardHolderRepository(injector *do.Injector) (biz.CardHolderRepo, error) {
	return &cardHolderRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *cardHolderRepository) Create(ctx context.Context, holder *model.CardHolder) error {
	return repo.DB(ctx).CardHolder.WithContext(ctx).Create(holder)
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
