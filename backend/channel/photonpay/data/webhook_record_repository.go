package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
)

type webhookRecordRepository struct{ repository *Repository }

func NewWebhookRecordRepository(injector do.Injector) (biz.WebhookRecordRepository, error) {
	return &webhookRecordRepository{repository: do.MustInvoke[*Repository](injector)}, nil
}

func (r *webhookRecordRepository) Create(ctx context.Context, item *model.WebhookRecord) error {
	return r.repository.DB(ctx).WebhookRecord.WithContext(ctx).Create(item)
}

func (r *webhookRecordRepository) Exist(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.WebhookRecord.WithContext(ctx).Where(
		db.WebhookRecord.ID.Eq(id),
		db.WebhookRecord.Channel.Eq(string(enums.Channel_PhotonPay)),
	).Count()

	return count > 0, err
}

func (r *webhookRecordRepository) Find(ctx context.Context, id model.ID) (*model.WebhookRecord, error) {
	db := r.repository.DB(ctx)

	return db.WebhookRecord.WithContext(ctx).
		Preload(db.WebhookRecord.Account).
		Where(
			db.WebhookRecord.ID.Eq(id),
			db.WebhookRecord.Channel.Eq(string(enums.Channel_PhotonPay)),
		).First()
}

func (r *webhookRecordRepository) Count(ctx context.Context, req *biz.WebhookRecordCountRequest) (int64, error) {
	db := r.repository.DB(ctx)
	query := db.WebhookRecord.WithContext(ctx).Where(db.WebhookRecord.Channel.Eq(string(enums.Channel_PhotonPay)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.WebhookRecord.AccountID.In(req.AccountIDs...))
	}
	return query.Count()
}

func (r *webhookRecordRepository) List(ctx context.Context, req *biz.WebhookRecordListRequest) ([]*model.WebhookRecord, error) {
	db := r.repository.DB(ctx)
	query := db.WebhookRecord.WithContext(ctx).Where(db.WebhookRecord.Channel.Eq(string(enums.Channel_PhotonPay)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.WebhookRecord.AccountID.In(req.AccountIDs...))
	}
	return query.
		Preload(db.WebhookRecord.Account).
		Order(db.WebhookRecord.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *webhookRecordRepository) Save(ctx context.Context, item *model.WebhookRecord) error {
	return r.repository.DB(ctx).WebhookRecord.WithContext(ctx).Save(item)
}

var _ biz.WebhookRecordRepository = (*webhookRecordRepository)(nil)
