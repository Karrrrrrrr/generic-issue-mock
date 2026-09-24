package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"
)

func (repo *cardTransactionRepository) ExistsRequest(ctx context.Context, req *biz.CardTransactionRequestExistsRequest) (bool, error) {
	table := repo.DB(ctx).CardTransaction
	count, err := table.WithContext(ctx).Where(
		table.AccountID.Eq(req.AccountID),
		table.Channel.Eq(string(common.Channel_PingPong)),
		table.RequestID.Eq(req.RequestID),
	).Count()
	return count > 0, err
}

func (repo *cardTransactionRepository) FindRequest(ctx context.Context, req *biz.CardTransactionRequestFindRequest) (*model.CardTransaction, error) {
	table := repo.DB(ctx).CardTransaction
	return table.WithContext(ctx).Where(
		table.AccountID.Eq(req.AccountID),
		table.Channel.Eq(string(common.Channel_PingPong)),
		table.RequestID.Eq(req.RequestID),
	).Order(table.ID.Desc()).First()
}
