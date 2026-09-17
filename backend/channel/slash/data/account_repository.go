package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type accountRepository struct{ repository *SlashRepository }

func NewAccountRepository(injector *do.Injector) (biz.SlashAccountRepository, error) {
	return &accountRepository{repository: do.MustInvoke[*SlashRepository](injector)}, nil
}

func (r *accountRepository) Create(ctx context.Context, item *model.Account) error {
	return r.repository.DB(ctx).Account.WithContext(ctx).Create(item)
}

func (r *accountRepository) Save(ctx context.Context, item *model.Account) error {
	return r.repository.DB(ctx).Account.WithContext(ctx).Save(item)
}

func (r *accountRepository) Exist(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Account.WithContext(ctx).Where(
		db.Account.ID.Eq(id),
		db.Account.Channel.Eq(string(enums.Channel_Slash)),
	).Count()

	return count > 0, err
}

func (r *accountRepository) Find(ctx context.Context, id model.ID) (*model.Account, error) {
	db := r.repository.DB(ctx)

	return db.Account.WithContext(ctx).Where(
		db.Account.ID.Eq(id),
		db.Account.Channel.Eq(string(enums.Channel_Slash)),
	).First()
}

func (r *accountRepository) Count(ctx context.Context) (int64, error) {
	db := r.repository.DB(ctx)

	return db.Account.WithContext(ctx).Where(
		db.Account.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
}

func (r *accountRepository) List(
	ctx context.Context,
	req *biz.ListAccountsRequest,
) ([]*model.Account, error) {
	db := r.repository.DB(ctx)

	return db.Account.WithContext(ctx).
		Where(db.Account.Channel.Eq(string(enums.Channel_Slash))).
		Preload(db.Account.Wallet).
		Order(db.Account.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

var _ biz.SlashAccountRepository = (*accountRepository)(nil)
