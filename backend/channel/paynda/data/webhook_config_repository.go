package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type webhookConfigRepository struct{ repository *PayndaRepository }

func NewWebhookConfigRepository(injector *do.Injector) (biz.PayndaWebhookConfigRepository, error) {
	return &webhookConfigRepository{repository: do.MustInvoke[*PayndaRepository](injector)}, nil
}

func (r *webhookConfigRepository) Create(ctx context.Context, item *model.WebhookConfig) error {
	return r.repository.DB(ctx).WebhookConfig.WithContext(ctx).Create(item)
}

func (r *webhookConfigRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.WebhookConfig.WithContext(ctx).
		Where(
			db.WebhookConfig.ID.Eq(id),
			db.WebhookConfig.Channel.Eq(string(enums.Channel_Paynda)),
		).
		Count()
	return count > 0, err
}

func (r *webhookConfigRepository) FindByID(ctx context.Context, id model.ID) (*model.WebhookConfig, error) {
	db := r.repository.DB(ctx)
	return db.WebhookConfig.WithContext(ctx).
		Preload(db.WebhookConfig.Account).
		Where(
			db.WebhookConfig.ID.Eq(id),
			db.WebhookConfig.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *webhookConfigRepository) ListByAccountIDs(ctx context.Context, req *biz.WebhookConfigListByAccountIDsRequest) ([]*model.WebhookConfig, error) {
	db := r.repository.DB(ctx)
	query := db.WebhookConfig.WithContext(ctx).
		Preload(db.WebhookConfig.Account).
		Where(db.WebhookConfig.Channel.Eq(string(enums.Channel_Paynda)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.WebhookConfig.AccountID.In(req.AccountIDs...))
	}

	return query.Order(db.WebhookConfig.ID.Desc()).Find()
}

func (r *webhookConfigRepository) Save(ctx context.Context, item *model.WebhookConfig) error {
	return r.repository.DB(ctx).WebhookConfig.WithContext(ctx).Save(item)
}

func (r *webhookConfigRepository) Delete(ctx context.Context, id model.ID) error {
	db := r.repository.DB(ctx)
	_, err := db.WebhookConfig.WithContext(ctx).Where(
		db.WebhookConfig.ID.Eq(id),
		db.WebhookConfig.Channel.Eq(string(enums.Channel_Paynda)),
	).Delete()
	return err
}
