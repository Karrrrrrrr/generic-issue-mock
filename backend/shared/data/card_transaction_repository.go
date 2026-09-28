package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
)

type cardTransactionRepository struct {
	*Repository
}

func NewCardTransactionRepository(injector do.Injector) (biz.CardTransactionRepo, error) {
	return &cardTransactionRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *cardTransactionRepository) Create(ctx context.Context, transaction *model.CardTransaction) error {
	return repo.DB(ctx).CardTransaction.WithContext(ctx).Create(transaction)
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
