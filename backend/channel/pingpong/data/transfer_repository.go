package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

type transferRepository struct{ *PingPongRepository }

func NewTransferRepository(injector do.Injector) (biz.PingPongTransferRepository, error) {
	return &transferRepository{PingPongRepository: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (repo *transferRepository) Exists(ctx context.Context, req *biz.TransferExistsRequest) (bool, error) {
	table := repo.DB(ctx).WalletTransfer
	count, err := repo.DB(ctx).WalletTransfer.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
			table.AccountID.Eq(req.AccountID),
		).Count()
	return count > 0, err
}

func (repo *transferRepository) Find(ctx context.Context, req *biz.TransferFindRequest) (*model.WalletTransfer, error) {
	table := repo.DB(ctx).WalletTransfer
	return repo.DB(ctx).WalletTransfer.WithContext(ctx).
		Preload(table.Account).
		Preload(table.Card).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
			table.AccountID.Eq(req.AccountID),
		).First()
}

func (repo *transferRepository) Create(ctx context.Context, item *model.WalletTransfer) error {
	return repo.DB(ctx).WalletTransfer.WithContext(ctx).Create(item)
}

func (repo *transferRepository) List(ctx context.Context, req *biz.TransferListRequest) ([]*model.WalletTransfer, error) {
	table := repo.DB(ctx).WalletTransfer
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		statement = statement.Where(table.AccountID.In(req.AccountIDs...))
	}
	if len(req.RequestIDs) != 0 {
		statement = statement.Where(table.RequestID.In(req.RequestIDs...))
	}
	if len(req.CardIDs) != 0 {
		statement = statement.Where(table.CardID.In(req.CardIDs...))
	}
	if len(req.Kinds) != 0 {
		values := make([]string, 0, len(req.Kinds))
		for _, value := range req.Kinds {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Kind.In(values...))
	}
	if req.CreatedFrom != nil {
		statement = statement.Where(table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		statement = statement.Where(table.CreatedAt.Lte(*req.CreatedTo))
	}
	statement = statement.Preload(table.Account)
	statement = statement.Preload(table.Card)
	if req.Limit != nil {
		statement = statement.Limit(*req.Limit)
	}
	return statement.Order(table.ID.Desc()).Offset(req.Offset).Find()
}

func (repo *transferRepository) Count(ctx context.Context, req *biz.TransferCountRequest) (int64, error) {
	table := repo.DB(ctx).WalletTransfer
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		statement = statement.Where(table.AccountID.In(req.AccountIDs...))
	}
	if len(req.RequestIDs) != 0 {
		statement = statement.Where(table.RequestID.In(req.RequestIDs...))
	}
	if len(req.CardIDs) != 0 {
		statement = statement.Where(table.CardID.In(req.CardIDs...))
	}
	if len(req.Kinds) != 0 {
		values := make([]string, 0, len(req.Kinds))
		for _, value := range req.Kinds {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Kind.In(values...))
	}
	if req.CreatedFrom != nil {
		statement = statement.Where(table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		statement = statement.Where(table.CreatedAt.Lte(*req.CreatedTo))
	}
	return statement.Count()
}
