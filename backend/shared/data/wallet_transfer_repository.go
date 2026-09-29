package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
)

type walletTransferRepository struct {
	*Repository
}

var _ biz.WalletTransferRepo = (*walletTransferRepository)(nil)

func NewWalletTransferRepository(injector do.Injector) (biz.WalletTransferRepo, error) {
	return &walletTransferRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *walletTransferRepository) List(ctx context.Context, req *biz.WalletTransferListRequest) ([]*model.WalletTransfer, error) {
	db := repo.DB(ctx)
	table := db.WalletTransfer
	query := table.WithContext(ctx).
		Preload(table.Account).
		Where(repo.buildPredicates(ctx, &req.WalletTransferFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset)
	if req.Limit != nil {
		query = query.Limit(*req.Limit)
	}
	return query.Find()
}

func (repo *walletTransferRepository) Count(ctx context.Context, req *biz.WalletTransferCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.WalletTransfer.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.WalletTransferFilters)...).
		Count()
}

func (repo *walletTransferRepository) buildPredicates(ctx context.Context, req *biz.WalletTransferFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.WalletTransfer
	predicates := []gen.Condition{table.Channel.Eq(string(req.Channel))}
	if len(req.IDs) != 0 {
		predicates = append(predicates, table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, table.AccountID.In(req.AccountIDs...))
	}
	if len(req.CardIDs) != 0 {
		predicates = append(predicates, table.CardID.In(req.CardIDs...))
	}
	if len(req.Kinds) != 0 {
		values := make([]string, 0, len(req.Kinds))
		for _, value := range req.Kinds {
			values = append(values, string(value))
		}
		predicates = append(predicates, table.Kind.In(values...))
	}
	if req.CreatedFrom != nil {
		predicates = append(predicates, table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		predicates = append(predicates, table.CreatedAt.Lte(*req.CreatedTo))
	}
	return predicates
}

func (repo *walletTransferRepository) Create(ctx context.Context, req *biz.WalletTransferCreateRequest) error {
	table := repo.DB(ctx).WalletTransfer
	return table.WithContext(ctx).Create(req.WalletTransfer)
}

func (repo *walletTransferRepository) ExistByRequestID(ctx context.Context, req *biz.WalletTransferExistByRequestIDRequest) (bool, error) {
	table := repo.DB(ctx).WalletTransfer
	count, err := table.WithContext(ctx).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.AccountID.Eq(req.AccountID),
			table.RequestID.Eq(req.RequestID),
		).
		Count()
	return count > 0, err
}

func (repo *walletTransferRepository) FindByRequestID(ctx context.Context, req *biz.WalletTransferFindByRequestIDRequest) (*model.WalletTransfer, error) {
	table := repo.DB(ctx).WalletTransfer
	return table.WithContext(ctx).
		Preload(table.Account).
		Where(
			table.Channel.Eq(string(req.Channel)),
			table.AccountID.Eq(req.AccountID),
			table.RequestID.Eq(req.RequestID),
		).
		First()
}

func (repo *walletTransferRepository) Exist(ctx context.Context, req *biz.WalletTransferExistRequest) (bool, error) {
	table := repo.DB(ctx).WalletTransfer
	count, err := table.WithContext(ctx).Where(table.AccountID.Eq(req.AccountID), table.Channel.Eq(string(req.Channel)), table.ID.Eq(req.ID)).Count()
	return count > 0, err
}

func (repo *walletTransferRepository) Find(ctx context.Context, req *biz.WalletTransferFindRequest) (*model.WalletTransfer, error) {
	table := repo.DB(ctx).WalletTransfer
	return table.WithContext(ctx).Where(table.AccountID.Eq(req.AccountID), table.Channel.Eq(string(req.Channel)), table.ID.Eq(req.ID)).First()
}
