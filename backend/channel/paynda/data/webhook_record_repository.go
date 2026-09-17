package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/model"

	"github.com/samber/do"
)

type webhookRecordRepository struct{ repository *PayndaRepository }

func NewWebhookRecordRepository(injector *do.Injector) (biz.PayndaWebhookRecordRepository, error) {
	return &webhookRecordRepository{repository: do.MustInvoke[*PayndaRepository](injector)}, nil
}

func (r *webhookRecordRepository) Create(ctx context.Context, item *model.WebhookRecord) error {
	return r.repository.DB(ctx).WebhookRecord.WithContext(ctx).Create(item)
}

func (r *webhookRecordRepository) Save(ctx context.Context, item *model.WebhookRecord) error {
	return r.repository.DB(ctx).WebhookRecord.WithContext(ctx).Save(item)
}

var _ biz.PayndaWebhookRecordRepository = (*webhookRecordRepository)(nil)
