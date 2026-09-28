package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type cardRepository struct {
	*Repository
}

func NewCardRepository(injector *do.Injector) (biz.CardRepo, error) {
	return &cardRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *cardRepository) Create(ctx context.Context, card *model.Card) error {
	return repo.DB(ctx).Card.WithContext(ctx).Create(card)
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

func (repo *cardRepository) FindByIDWithLock(ctx context.Context, req *biz.CardFindByIDWithLockRequest) (*model.Card, error) {
	db := repo.DB(ctx)
	table := db.Card
	return table.WithContext(ctx).
		Preload(table.Account).
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
