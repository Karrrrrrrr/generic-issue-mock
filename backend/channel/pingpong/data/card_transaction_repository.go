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

func (repo *cardTransactionRepository) List(ctx context.Context, req *biz.CardTransactionListRequest) ([]*model.CardTransaction, error) {
	table := repo.DB(ctx).CardTransaction
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if len(req.AccountIDs) != 0 {
		statement = statement.Where(table.AccountID.In(req.AccountIDs...))
	}
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	if len(req.CardIDs) != 0 {
		statement = statement.Where(table.CardID.In(req.CardIDs...))
	}
	if len(req.AuthorizationIDs) != 0 {
		statement = statement.Where(table.AuthorizationID.In(req.AuthorizationIDs...))
	}
	if len(req.Types) != 0 {
		values := make([]string, 0, len(req.Types))
		for _, value := range req.Types {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Type.In(values...))
	}
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Status.In(values...))
	}
	if req.CreatedFrom != nil {
		statement = statement.Where(table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		statement = statement.Where(table.CreatedAt.Lte(*req.CreatedTo))
	}
	if req.Limit != nil {
		statement = statement.Limit(*req.Limit)
	}
	return statement.Preload(table.Account).Order(table.ID.Desc()).Offset(req.Offset).Find()
}

func (repo *cardTransactionRepository) Count(ctx context.Context, req *biz.CardTransactionCountRequest) (int64, error) {
	table := repo.DB(ctx).CardTransaction
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if len(req.AccountIDs) != 0 {
		statement = statement.Where(table.AccountID.In(req.AccountIDs...))
	}
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	if len(req.CardIDs) != 0 {
		statement = statement.Where(table.CardID.In(req.CardIDs...))
	}
	if len(req.AuthorizationIDs) != 0 {
		statement = statement.Where(table.AuthorizationID.In(req.AuthorizationIDs...))
	}
	if len(req.Types) != 0 {
		values := make([]string, 0, len(req.Types))
		for _, value := range req.Types {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Type.In(values...))
	}
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Status.In(values...))
	}
	if req.CreatedFrom != nil {
		statement = statement.Where(table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		statement = statement.Where(table.CreatedAt.Lte(*req.CreatedTo))
	}
	return statement.Count()
}

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
