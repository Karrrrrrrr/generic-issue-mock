package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/model"

	"github.com/samber/do"
)

type webhookRecordRepository struct {
	repository *SlashRepository
}

func NewWebhookRecordRepository(injector *do.Injector) (biz.SlashWebhookRecordRepository, error) {
	return &webhookRecordRepository{
		repository: do.MustInvoke[*SlashRepository](injector),
	}, nil
}

func (r *webhookRecordRepository) Create(ctx context.Context, item *model.WebhookRecord) error {
	return r.repository.DB(ctx).WebhookRecord.WithContext(ctx).Create(item)
}

func (r *webhookRecordRepository) Save(ctx context.Context, item *model.WebhookRecord) error {
	return r.repository.DB(ctx).WebhookRecord.WithContext(ctx).Save(item)
}

var _ biz.SlashWebhookRecordRepository = (*webhookRecordRepository)(nil)
