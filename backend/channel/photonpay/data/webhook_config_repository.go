package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type webhookConfigRepository struct{ repository *Repository }

func NewWebhookConfigRepository(injector *do.Injector) (biz.WebhookConfigRepository, error) {
	return &webhookConfigRepository{repository: do.MustInvoke[*Repository](injector)}, nil
}
func (r *webhookConfigRepository) Create(ctx context.Context, item *model.WebhookConfig) error {
	return r.repository.DB(ctx).WebhookConfig.WithContext(ctx).Create(item)
}
func (r *webhookConfigRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.WebhookConfig.WithContext(ctx).Where(db.WebhookConfig.ID.Eq(id), db.WebhookConfig.Channel.Eq(string(enums.Channel_PhotonPay))).Count()
	return count > 0, err
}
func (r *webhookConfigRepository) FindByID(ctx context.Context, id model.ID) (*model.WebhookConfig, error) {
	db := r.repository.DB(ctx)
	return db.WebhookConfig.WithContext(ctx).Where(db.WebhookConfig.ID.Eq(id), db.WebhookConfig.Channel.Eq(string(enums.Channel_PhotonPay))).First()
}
func (r *webhookConfigRepository) List(ctx context.Context) ([]*model.WebhookConfig, error) {
	db := r.repository.DB(ctx)
	return db.WebhookConfig.WithContext(ctx).Where(db.WebhookConfig.Channel.Eq(string(enums.Channel_PhotonPay))).Order(db.WebhookConfig.ID.Desc()).Find()
}
func (r *webhookConfigRepository) Save(ctx context.Context, item *model.WebhookConfig) error {
	return r.repository.DB(ctx).WebhookConfig.WithContext(ctx).Save(item)
}
func (r *webhookConfigRepository) Delete(ctx context.Context, id model.ID) error {
	db := r.repository.DB(ctx)
	_, err := db.WebhookConfig.WithContext(ctx).Where(db.WebhookConfig.ID.Eq(id), db.WebhookConfig.Channel.Eq(string(enums.Channel_PhotonPay))).Delete()
	return err
}
