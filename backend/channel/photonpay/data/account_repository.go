package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

type accountRepository struct{ repository *Repository }

func NewAccountRepository(injector do.Injector) (biz.AccountRepository, error) {
	return &accountRepository{repository: do.MustInvoke[*Repository](injector)}, nil
}

func (r *accountRepository) Create(ctx context.Context, item *model.Account) error {
	return r.repository.DB(ctx).Account.WithContext(ctx).Create(item)
}

func (r *accountRepository) Exist(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Account.WithContext(ctx).Where(
		db.Account.ID.Eq(id),
		db.Account.Channel.Eq(string(enums.Channel_PhotonPay)),
	).Count()
	return count > 0, err
}

func (r *accountRepository) Find(ctx context.Context, id model.ID) (*model.Account, error) {
	db := r.repository.DB(ctx)
	return db.Account.WithContext(ctx).Where(db.Account.ID.Eq(id), db.Account.Channel.Eq(string(enums.Channel_PhotonPay))).First()
}

func (r *accountRepository) Count(ctx context.Context, req *biz.AccountCountRequest) (int64, error) {
	db := r.repository.DB(ctx)
	query := db.Account.WithContext(ctx).Where(db.Account.Channel.Eq(string(enums.Channel_PhotonPay)))
	if len(req.IDs) != 0 {
		query = query.Where(db.Account.ID.In(req.IDs...))
	}
	return query.Count()
}

func (r *accountRepository) List(ctx context.Context, req *biz.AccountListRequest) ([]*model.Account, error) {
	db := r.repository.DB(ctx)
	query := db.Account.WithContext(ctx).Where(db.Account.Channel.Eq(string(enums.Channel_PhotonPay)))
	if len(req.IDs) != 0 {
		query = query.Where(db.Account.ID.In(req.IDs...))
	}
	return query.
		Preload(db.Account.Wallet).
		Order(db.Account.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *accountRepository) Save(ctx context.Context, item *model.Account) error {
	return r.repository.DB(ctx).Account.WithContext(ctx).Save(item)
}

var _ biz.AccountRepository = (*accountRepository)(nil)
