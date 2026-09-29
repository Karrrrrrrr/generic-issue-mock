package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

var _ biz.PingPongCardRepository = (*cardRepository)(nil)

type cardRepository struct{ *PingPongRepository }

func NewCardRepository(injector do.Injector) (biz.PingPongCardRepository, error) {
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
		Preload(table.CardHolder).
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
