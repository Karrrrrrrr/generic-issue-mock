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

func (r *accountRepository) List(ctx context.Context) ([]*model.Account, error) {
	db := r.repository.DB(ctx)
	return db.Account.WithContext(ctx).Where(
		db.Account.Channel.Eq(string(enums.Channel_Slash)),
	).Order(db.Account.ID.Desc()).Find()
}

var _ biz.SlashAccountRepository = (*accountRepository)(nil)
