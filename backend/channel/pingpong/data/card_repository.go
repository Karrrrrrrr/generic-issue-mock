package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type cardRepository struct{ *PingPongRepository }

func NewCardRepository(injector *do.Injector) (biz.PingPongCardRepository, error) {
	return &cardRepository{PingPongRepository: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (repo *cardRepository) Exists(ctx context.Context, req *biz.CardExistsRequest) (bool, error) {
	table := repo.DB(ctx).Card
	count, err := repo.DB(ctx).Card.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
			table.AccountID.Eq(req.AccountID),
		).Count()
	return count > 0, err
}

func (repo *cardRepository) Find(ctx context.Context, req *biz.CardFindRequest) (*model.Card, error) {
	table := repo.DB(ctx).Card
	return repo.DB(ctx).Card.WithContext(ctx).
		Preload(table.Wallet).
		Preload(table.Account).
		Preload(table.VirtualAccount.Wallet).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
			table.AccountID.Eq(req.AccountID),
		).First()
}

func (repo *cardRepository) Create(ctx context.Context, item *model.Card) error {
	return repo.DB(ctx).Card.WithContext(ctx).Create(item)
}

func (repo *cardRepository) Save(ctx context.Context, item *model.Card) error {
	table := repo.DB(ctx).Card
	_, err := table.WithContext(ctx).Where(
		table.ID.Eq(item.ID),
		table.Channel.Eq(string(common.Channel_PingPong)),
		table.AccountID.Eq(item.AccountID),
	).UpdateSimple(
		table.Status.Value(string(item.Status)),
	)
	return err
}

func (repo *cardRepository) List(ctx context.Context, req *biz.CardListRequest) ([]*model.Card, error) {
	table := repo.DB(ctx).Card
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
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Status.In(values...))
	}
	statement = statement.Preload(table.Wallet)
	statement = statement.Preload(table.Account)
	statement = statement.Preload(table.VirtualAccount.Wallet)
	if req.Limit != nil {
		statement = statement.Limit(*req.Limit)
	}
	return statement.Order(table.ID.Desc()).Offset(req.Offset).Find()
}

func (repo *cardRepository) Count(ctx context.Context, req *biz.CardCountRequest) (int64, error) {
	table := repo.DB(ctx).Card
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
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Status.In(values...))
	}
	return statement.Count()
}
