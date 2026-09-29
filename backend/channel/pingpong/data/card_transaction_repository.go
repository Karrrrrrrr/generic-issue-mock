package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

type cardTransactionRepository struct{ *PingPongRepository }

func NewCardTransactionRepository(injector do.Injector) (biz.PingPongCardTransactionRepository, error) {
	return &cardTransactionRepository{PingPongRepository: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (repo *cardTransactionRepository) Create(ctx context.Context, item *model.CardTransaction) error {
	return repo.DB(ctx).CardTransaction.WithContext(ctx).Create(item)
}

var _ biz.PingPongCardTransactionRepository = (*cardTransactionRepository)(nil)

func (repo *cardTransactionRepository) ExistsForSimulation(ctx context.Context, req *biz.CardTransactionExistsForSimulationRequest) (bool, error) {
	table := repo.DB(ctx).CardTransaction
	count, err := table.WithContext(ctx).Where(
		table.ID.Eq(req.ID),
		table.Channel.Eq(string(common.Channel_PingPong)),
	).Count()
	return count > 0, err
}

func (repo *cardTransactionRepository) FindForSimulation(ctx context.Context, req *biz.CardTransactionFindForSimulationRequest) (*model.CardTransaction, error) {
	table := repo.DB(ctx).CardTransaction
	return table.WithContext(ctx).Where(
		table.ID.Eq(req.ID),
		table.Channel.Eq(string(common.Channel_PingPong)),
	).First()
}
