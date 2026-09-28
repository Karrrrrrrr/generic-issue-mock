package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

var _ biz.PingPongAuthorizationRepository = (*authorizationRepository)(nil)

type authorizationRepository struct{ *PingPongRepository }

func NewAuthorizationRepository(injector do.Injector) (biz.PingPongAuthorizationRepository, error) {
	return &authorizationRepository{PingPongRepository: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (repo *authorizationRepository) Exists(ctx context.Context, req *biz.AuthorizationExistsRequest) (bool, error) {
	table := repo.DB(ctx).Authorization
	count, err := repo.DB(ctx).Authorization.WithContext(ctx).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
			table.AccountID.Eq(req.AccountID),
		).Count()
	return count > 0, err
}

func (repo *authorizationRepository) Find(ctx context.Context, req *biz.AuthorizationFindRequest) (*model.Authorization, error) {
	table := repo.DB(ctx).Authorization
	return repo.DB(ctx).Authorization.WithContext(ctx).
		Preload(table.Account).
		Preload(table.CardTransactions).
		Where(
			table.ID.Eq(req.ID),
			table.Channel.Eq(string(common.Channel_PingPong)),
			table.AccountID.Eq(req.AccountID),
		).First()
}

func (repo *authorizationRepository) Create(ctx context.Context, item *model.Authorization) error {
	return repo.DB(ctx).Authorization.WithContext(ctx).Create(item)
}

func (repo *authorizationRepository) List(ctx context.Context, req *biz.AuthorizationListRequest) ([]*model.Authorization, error) {
	table := repo.DB(ctx).Authorization
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if req.MerchantName != nil {
		statement = statement.Where(table.MerchantName.Like("%" + *req.MerchantName + "%"))
	}
	if req.CreatedFrom != nil {
		statement = statement.Where(table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		statement = statement.Where(table.CreatedAt.Lte(*req.CreatedTo))
	}
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Status.In(values...))
	}
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		statement = statement.Where(table.AccountID.In(req.AccountIDs...))
	}
	if len(req.CardIDs) != 0 {
		statement = statement.Where(table.CardID.In(req.CardIDs...))
	}
	statement = statement.Preload(table.Account)
	statement = statement.Preload(table.CardTransactions)
	if req.Limit != nil {
		statement = statement.Limit(*req.Limit)
	}
	return statement.Order(table.ID.Desc()).Offset(req.Offset).Find()
}

func (repo *authorizationRepository) Count(ctx context.Context, req *biz.AuthorizationCountRequest) (int64, error) {
	table := repo.DB(ctx).Authorization
	statement := table.WithContext(ctx).Where(table.Channel.Eq(string(common.Channel_PingPong)))
	if req.MerchantName != nil {
		statement = statement.Where(table.MerchantName.Like("%" + *req.MerchantName + "%"))
	}
	if req.CreatedFrom != nil {
		statement = statement.Where(table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		statement = statement.Where(table.CreatedAt.Lte(*req.CreatedTo))
	}
	if len(req.Statuses) != 0 {
		values := make([]string, 0, len(req.Statuses))
		for _, value := range req.Statuses {
			values = append(values, string(value))
		}
		statement = statement.Where(table.Status.In(values...))
	}
	if len(req.IDs) != 0 {
		statement = statement.Where(table.ID.In(req.IDs...))
	}
	if len(req.AccountIDs) != 0 {
		statement = statement.Where(table.AccountID.In(req.AccountIDs...))
	}
	if len(req.CardIDs) != 0 {
		statement = statement.Where(table.CardID.In(req.CardIDs...))
	}
	return statement.Count()
}
