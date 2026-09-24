package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"
)

func (repo *cardRepository) ExistsRequest(ctx context.Context, req *biz.CardRequestExistsRequest) (bool, error) {
	table := repo.DB(ctx).Card
	count, err := table.WithContext(ctx).Where(
		table.AccountID.Eq(req.AccountID),
		table.Channel.Eq(string(common.Channel_PingPong)),
		table.RequestID.Eq(req.RequestID),
	).Count()
	return count > 0, err
}

func (repo *cardRepository) FindRequest(ctx context.Context, req *biz.CardRequestFindRequest) (*model.Card, error) {
	table := repo.DB(ctx).Card
	return table.WithContext(ctx).Where(
		table.AccountID.Eq(req.AccountID),
		table.Channel.Eq(string(common.Channel_PingPong)),
		table.RequestID.Eq(req.RequestID),
	).Order(table.ID.Desc()).First()
}
