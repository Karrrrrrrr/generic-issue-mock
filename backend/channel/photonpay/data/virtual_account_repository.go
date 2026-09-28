package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

type virtualAccountRepository struct {
	repository *Repository
}

func NewVirtualAccountRepository(injector do.Injector) (biz.VirtualAccountRepository, error) {
	return &virtualAccountRepository{
		repository: do.MustInvoke[*Repository](injector),
	}, nil
}

func (r *virtualAccountRepository) Create(ctx context.Context, item *model.VirtualAccount) error {
	return r.repository.DB(ctx).VirtualAccount.WithContext(ctx).Create(item)
}

func (r *virtualAccountRepository) FindByAccountID(
	ctx context.Context,
	accountID model.ID,
) (*model.VirtualAccount, error) {
	db := r.repository.DB(ctx)
	return db.VirtualAccount.WithContext(ctx).
		Preload(db.VirtualAccount.Account).
		Preload(db.VirtualAccount.Wallet).
		Where(
			db.VirtualAccount.AccountID.Eq(accountID),
			db.VirtualAccount.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		Order(db.VirtualAccount.ID.Asc()).
		First()
}

var _ biz.VirtualAccountRepository = (*virtualAccountRepository)(nil)

func (r *virtualAccountRepository) ListVirtualAccounts(ctx context.Context, req *biz.VirtualAccountListRequest) ([]*model.VirtualAccount, error) {
	db := r.repository.DB(ctx)
	query := db.VirtualAccount.WithContext(ctx).
		Preload(db.VirtualAccount.Account).
		Preload(db.VirtualAccount.Wallet).
		Where(
			db.VirtualAccount.Channel.Eq(string(enums.Channel_PhotonPay)),
		)
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.VirtualAccount.AccountID.In(req.AccountIDs...))
	}
	return query.Order(db.VirtualAccount.ID.Desc()).Find()
}
