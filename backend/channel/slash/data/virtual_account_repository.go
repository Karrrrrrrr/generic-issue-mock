package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

type virtualAccountRepository struct{ repository *SlashRepository }

var _ biz.SlashVirtualAccountRepository = (*virtualAccountRepository)(nil)

func NewVirtualAccountRepository(injector do.Injector) (biz.SlashVirtualAccountRepository, error) {
	return &virtualAccountRepository{repository: do.MustInvoke[*SlashRepository](injector)}, nil
}

func (r *virtualAccountRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.VirtualAccount.WithContext(ctx).
		Where(
			db.VirtualAccount.ID.Eq(id),
			db.VirtualAccount.Channel.Eq(string(enums.Channel_Slash)),
		).
		Count()
	return count > 0, err
}

func (r *virtualAccountRepository) FindByID(ctx context.Context, id model.ID) (*model.VirtualAccount, error) {
	db := r.repository.DB(ctx)
	return db.VirtualAccount.WithContext(ctx).
		Preload(db.VirtualAccount.Account).
		Where(
			db.VirtualAccount.ID.Eq(id),
			db.VirtualAccount.Channel.Eq(string(enums.Channel_Slash)),
		).
		First()
}

func (r *virtualAccountRepository) ExistByAccountID(ctx context.Context, req *biz.VirtualAccountExistByAccountIDRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.VirtualAccount.WithContext(ctx).Where(
		db.VirtualAccount.ID.Eq(req.ID),
		db.VirtualAccount.AccountID.Eq(*req.AccountID),
		db.VirtualAccount.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
	return count > 0, err
}

func (r *virtualAccountRepository) FindByAccountID(ctx context.Context, req *biz.VirtualAccountFindByAccountIDRequest) (*model.VirtualAccount, error) {
	db := r.repository.DB(ctx)
	return db.VirtualAccount.WithContext(ctx).
		Preload(db.VirtualAccount.Wallet).
		Preload(db.VirtualAccount.Account).
		Where(
			db.VirtualAccount.ID.Eq(req.ID),
			db.VirtualAccount.AccountID.Eq(*req.AccountID),
			db.VirtualAccount.Channel.Eq(string(enums.Channel_Slash)),
		).First()
}

func (r *virtualAccountRepository) ListVirtualAccounts(ctx context.Context, req *biz.VirtualAccountListRequest) ([]*model.VirtualAccount, error) {
	db := r.repository.DB(ctx)
	query := db.VirtualAccount.WithContext(ctx).
		Preload(db.VirtualAccount.Account).
		Preload(db.VirtualAccount.Wallet).
		Where(
			db.VirtualAccount.Channel.Eq(string(enums.Channel_Slash)),
		)
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.VirtualAccount.AccountID.In(req.AccountIDs...))
	}
	return query.Order(db.VirtualAccount.ID.Desc()).Find()
}

func (r *virtualAccountRepository) CreateVirtualAccount(ctx context.Context, item *model.VirtualAccount) error {
	return r.repository.DB(ctx).VirtualAccount.WithContext(ctx).Create(item)
}

func (r *virtualAccountRepository) Save(ctx context.Context, req *biz.SaveVirtualAccountRequest) error {
	db := r.repository.DB(ctx)
	_, err := db.VirtualAccount.WithContext(ctx).Where(
		db.VirtualAccount.ID.Eq(req.ID),
		db.VirtualAccount.AccountID.Eq(req.AccountID),
		db.VirtualAccount.Channel.Eq(string(enums.Channel_Slash)),
	).Update(db.VirtualAccount.Name, req.Name)
	return err
}
