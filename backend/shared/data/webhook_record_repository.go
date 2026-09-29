package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"gorm.io/gen"
)

type webhookRecordRepository struct {
	*Repository
}

var _ biz.WebhookRecordRepo = (*webhookRecordRepository)(nil)

func NewWebhookRecordRepository(injector do.Injector) (biz.WebhookRecordRepo, error) {
	return &webhookRecordRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *webhookRecordRepository) List(ctx context.Context, req *biz.WebhookRecordListRequest) ([]*model.WebhookRecord, error) {
	db := repo.DB(ctx)
	table := db.WebhookRecord
	query := table.WithContext(ctx).
		Preload(table.Account).
		Where(repo.buildPredicates(ctx, &req.WebhookRecordFilters)...).
		Order(table.ID.Desc()).
		Offset(req.Offset)
	if req.Limit != nil {
		query = query.Limit(*req.Limit)
	}
	return query.Find()
}

func (repo *webhookRecordRepository) Count(ctx context.Context, req *biz.WebhookRecordCountRequest) (int64, error) {
	db := repo.DB(ctx)
	return db.WebhookRecord.WithContext(ctx).
		Where(repo.buildPredicates(ctx, &req.WebhookRecordFilters)...).
		Count()
}

func (repo *webhookRecordRepository) buildPredicates(ctx context.Context, req *biz.WebhookRecordFilters) []gen.Condition {
	db := repo.DB(ctx)
	table := db.WebhookRecord
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
	if len(req.Statuses) != 0 {
		statuses := make([]string, 0, len(req.Statuses))
		for _, status := range req.Statuses {
			statuses = append(statuses, string(status))
		}
		predicates = append(predicates, table.Status.In(statuses...))
	}
	if req.CreatedFrom != nil {
		predicates = append(predicates, table.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		predicates = append(predicates, table.CreatedAt.Lte(*req.CreatedTo))
	}
	return predicates
}

func (repo *webhookRecordRepository) Exist(ctx context.Context, req *biz.WebhookRecordExistRequest) (bool, error) {
	table := repo.DB(ctx).WebhookRecord
	count, err := table.WithContext(ctx).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		Count()
	return count > 0, err
}

func (repo *webhookRecordRepository) Find(ctx context.Context, req *biz.WebhookRecordFindRequest) (*model.WebhookRecord, error) {
	table := repo.DB(ctx).WebhookRecord
	return table.WithContext(ctx).
		Preload(table.Account).
		Where(
			table.AccountID.Eq(req.AccountID),
			table.Channel.Eq(string(req.Channel)),
			table.ID.Eq(req.ID),
		).
		First()
}

func (repo *webhookRecordRepository) Create(ctx context.Context, req *biz.WebhookRecordCreateRequest) error {
	return repo.DB(ctx).WebhookRecord.WithContext(ctx).Create(req.Record)
}

func (repo *webhookRecordRepository) ExistBySource(ctx context.Context, req *biz.WebhookRecordExistBySourceRequest) (bool, error) {
	table := repo.DB(ctx).WebhookRecord
	count, err := table.WithContext(ctx).Where(
		table.AccountID.Eq(req.AccountID), table.Channel.Eq(string(req.Channel)),
		table.WebhookConfigID.Eq(req.WebhookConfigID), table.Event.Eq(req.Event), table.SourceID.Eq(req.SourceID),
	).Count()
	return count > 0, err
}

func (repo *webhookRecordRepository) UpdateDelivery(ctx context.Context, req *biz.WebhookRecordUpdateDeliveryRequest) error {
	table := repo.DB(ctx).WebhookRecord
	_, err := table.WithContext(ctx).Where(
		table.AccountID.Eq(req.AccountID), table.Channel.Eq(string(req.Channel)), table.ID.Eq(req.ID),
	).Updates(map[string]interface{}{
		"status": req.Status, "status_code": req.StatusCode, "response_body": req.ResponseBody,
		"response_headers": req.ResponseHeaders, "delivered_at": req.DeliveredAt, "error_message": req.ErrorMessage,
	})
	return err
}
