package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/model"

	"github.com/samber/do"
)

type webhookRecordRepository struct{ repository *Repository }

func NewWebhookRecordRepository(injector *do.Injector) (biz.WebhookRecordRepository, error) {
	return &webhookRecordRepository{repository: do.MustInvoke[*Repository](injector)}, nil
}

func (r *webhookRecordRepository) Create(ctx context.Context, item *model.WebhookRecord) error {
	return r.repository.DB(ctx).WebhookRecord.WithContext(ctx).Create(item)
}

func (r *webhookRecordRepository) Save(ctx context.Context, item *model.WebhookRecord) error {
	return r.repository.DB(ctx).WebhookRecord.WithContext(ctx).Save(item)
}

var _ biz.WebhookRecordRepository = (*webhookRecordRepository)(nil)
