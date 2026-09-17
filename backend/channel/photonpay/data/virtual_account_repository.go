package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type virtualAccountRepository struct {
	repository *Repository
}

func NewVirtualAccountRepository(injector *do.Injector) (biz.VirtualAccountRepository, error) {
	return &virtualAccountRepository{
		repository: do.MustInvoke[*Repository](injector),
	}, nil
}

func (r *virtualAccountRepository) Create(ctx context.Context, item *model.VirtualAccount) error {
	return r.repository.DB(ctx).VirtualAccount.WithContext(ctx).Create(item)
}

func (r *virtualAccountRepository) List(ctx context.Context) ([]*model.VirtualAccount, error) {
	db := r.repository.DB(ctx)
	return db.VirtualAccount.WithContext(ctx).
		Preload(db.VirtualAccount.Wallet).
		Where(db.VirtualAccount.Channel.Eq(string(enums.Channel_PhotonPay))).
		Order(db.VirtualAccount.ID.Desc()).
		Find()
}

var _ biz.VirtualAccountRepository = (*virtualAccountRepository)(nil)
