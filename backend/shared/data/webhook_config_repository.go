package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
)

type webhookConfigRepository struct {
	*Repository
}

var _ biz.WebhookConfigRepo = (*webhookConfigRepository)(nil)

func NewWebhookConfigRepository(injector do.Injector) (biz.WebhookConfigRepo, error) {
	return &webhookConfigRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *webhookConfigRepository) List(ctx context.Context, req *biz.WebhookConfigListRequest) ([]*model.WebhookConfig, error) {
	db := repo.DB(ctx)
	table := db.WebhookConfig
	query := table.WithContext(ctx).
		Preload(table.Account).
		Where(repo.buildPredicates(ctx, &req.WebhookConfigFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset)
	if req.Limit != nil {
		query = query.Limit(*req.Limit)
	}
	return query.Find()
}

func (repo *webhookConfigRepository) Count(ctx context.Context, req *biz.WebhookConfigCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.WebhookConfig.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.WebhookConfigFilters)...).
		Count()
}

func (repo *webhookConfigRepository) buildPredicates(ctx context.Context, req *biz.WebhookConfigFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.WebhookConfig
	predicates := []gen.Condition{table.Channel.Eq(string(req.Channel))}
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, table.AccountID.In(req.AccountIDs...))
	}
	if len(req.IDs) != 0 {
		predicates = append(predicates, table.ID.In(req.IDs...))
	}
	if len(req.Events) != 0 {
		predicates = append(predicates, table.Event.In(req.Events...))
	}
	return predicates
}

func (repo *webhookConfigRepository) Exist(ctx context.Context, req *biz.WebhookConfigExistRequest) (bool, error) {
	table := repo.DB(ctx).WebhookConfig
	count, err := table.WithContext(ctx).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		Count()
	return count > 0, err
}

func (repo *webhookConfigRepository) Find(ctx context.Context, req *biz.WebhookConfigFindRequest) (*model.WebhookConfig, error) {
	table := repo.DB(ctx).WebhookConfig
	return table.WithContext(ctx).
		Preload(table.Account).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		First()
}

func (repo *webhookConfigRepository) Delete(ctx context.Context, req *biz.WebhookConfigDeleteRequest) error {
	table := repo.DB(ctx).WebhookConfig
	_, err := table.WithContext(ctx).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		Delete()
	return err
}

func (repo *webhookConfigRepository) Update(ctx context.Context, req *biz.WebhookConfigUpdateRequest) error {
	table := repo.DB(ctx).WebhookConfig
	_, err := table.WithContext(ctx).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		UpdateSimple(
			table.TargetURL.Value(req.TargetURL),
			table.Enabled.Value(req.Enabled),
		)
	return err
}

func (repo *webhookConfigRepository) Create(ctx context.Context, req *biz.WebhookConfigCreateRequest) error {
	return repo.DB(ctx).WebhookConfig.WithContext(ctx).Create(req.WebhookConfig)
}
