package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"
)

func (repo *transferRepository) ExistsRequest(ctx context.Context, req *biz.TransferRequestExistsRequest) (bool, error) {
	table := repo.DB(ctx).WalletTransfer
	count, err := table.WithContext(ctx).Where(
		table.AccountID.Eq(req.AccountID),
		table.Channel.Eq(string(common.Channel_PingPong)),
		table.RequestID.Eq(req.RequestID),
	).Count()
	return count > 0, err
}

func (repo *transferRepository) FindRequest(ctx context.Context, req *biz.TransferRequestFindRequest) (*model.WalletTransfer, error) {
	table := repo.DB(ctx).WalletTransfer
	return table.WithContext(ctx).Where(
		table.AccountID.Eq(req.AccountID),
		table.Channel.Eq(string(common.Channel_PingPong)),
		table.RequestID.Eq(req.RequestID),
	).Order(table.ID.Desc()).First()
}
