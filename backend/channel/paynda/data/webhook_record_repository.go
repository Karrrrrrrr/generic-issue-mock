package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do/v2"
	"gorm.io/gen/field"
)

type webhookRecordRepository struct{ repository *PayndaRepository }

func NewWebhookRecordRepository(injector do.Injector) (biz.PayndaWebhookRecordRepository, error) {
	return &webhookRecordRepository{repository: do.MustInvoke[*PayndaRepository](injector)}, nil
}

func (r *webhookRecordRepository) Create(ctx context.Context, item *model.WebhookRecord) error {
	return r.repository.DB(ctx).WebhookRecord.WithContext(ctx).Create(item)
}

func (r *webhookRecordRepository) Exist(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.WebhookRecord.WithContext(ctx).Where(
		db.WebhookRecord.ID.Eq(id),
		db.WebhookRecord.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *webhookRecordRepository) Find(ctx context.Context, id model.ID) (*model.WebhookRecord, error) {
	db := r.repository.DB(ctx)

	return db.WebhookRecord.WithContext(ctx).
		Preload(db.WebhookRecord.Account).
		Where(
			db.WebhookRecord.ID.Eq(id),
			db.WebhookRecord.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *webhookRecordRepository) Count(ctx context.Context, req *biz.WebhookRecordCountRequest) (int64, error) {
	db := r.repository.DB(ctx)
	query := db.WebhookRecord.WithContext(ctx).Where(db.WebhookRecord.Channel.Eq(string(enums.Channel_Paynda)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.WebhookRecord.AccountID.In(req.AccountIDs...))
	}
	return query.Count()
}

func (r *webhookRecordRepository) List(
	ctx context.Context,
	req *biz.WebhookRecordListRequest,
) ([]*model.WebhookRecord, error) {
	db := r.repository.DB(ctx)
	query := db.WebhookRecord.WithContext(ctx).
		Preload(db.WebhookRecord.Account).
		Where(
			db.WebhookRecord.Channel.Eq(string(enums.Channel_Paynda)),
		)
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.WebhookRecord.AccountID.In(req.AccountIDs...))
	}

	return query.Order(db.WebhookRecord.ID.Desc()).Offset(req.Offset).Limit(req.Limit).Find()
}

func (r *webhookRecordRepository) Save(ctx context.Context, item *model.WebhookRecord) error {
	db := r.repository.DB(ctx)
	table := db.WebhookRecord
	assigns := []field.AssignExpr{
		table.RequestHeaders.Value(item.RequestHeaders),
		table.ResponseHeaders.Value(item.ResponseHeaders),
		table.ResponseBody.Value(item.ResponseBody),
		table.StatusCode.Value(item.StatusCode),
		table.Status.Value(string(item.Status)),
		table.ErrorMessage.Value(item.ErrorMessage),
	}
	if item.DeliveredAt == nil {
		assigns = append(assigns, table.DeliveredAt.Null())
	} else {
		assigns = append(assigns, table.DeliveredAt.Value(*item.DeliveredAt))
	}
	_, err := table.WithContext(ctx).
		Where(
			table.ID.Eq(item.ID),
			table.Channel.Eq(string(enums.Channel_Paynda)),
		).
		UpdateSimple(assigns...)
	return err
}

var _ biz.PayndaWebhookRecordRepository = (*webhookRecordRepository)(nil)
