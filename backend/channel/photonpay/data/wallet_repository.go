package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type walletRepository struct{ repository *Repository }

func NewWalletRepository(injector *do.Injector) (biz.WalletRepository, error) {
	return &walletRepository{repository: do.MustInvoke[*Repository](injector)}, nil
}

func (r *walletRepository) Create(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Create(item)
}

func (r *walletRepository) FindByIDForUpdate(
	ctx context.Context,
	req *biz.ResourceRequest,
) (*model.Wallet, error) {
	db := r.repository.DB(ctx)

	return db.Wallet.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where(
		db.Wallet.ID.Eq(req.ID),
		db.Wallet.AccountID.Eq(*req.AccountID),
		db.Wallet.Channel.Eq(string(enums.Channel_PhotonPay)),
	).First()
}

func (r *walletRepository) Save(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Save(item)
}

var _ biz.WalletRepository = (*walletRepository)(nil)
