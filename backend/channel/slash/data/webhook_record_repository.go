package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/internal/query"
	"generic-mock/model"

	"github.com/samber/do"
	"gorm.io/gen"
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

type webhookRecordPredicatesRequest struct {
	DB      *query.Query
	Filters biz.WebhookRecordFilters
}

func webhookRecordPredicates(req *webhookRecordPredicatesRequest) []gen.Condition {
	table := req.DB.WebhookRecord
	filters := req.Filters
	predicates := []gen.Condition{table.Channel.Eq(string(enums.Channel_Slash))}
	if len(filters.AccountIDs) != 0 {
		predicates = append(predicates, table.AccountID.In(filters.AccountIDs...))
	}
	if len(filters.Events) != 0 {
		predicates = append(predicates, table.Event.In(filters.Events...))
	}
	if len(filters.Statuses) != 0 {
		statuses := make([]string, 0, len(filters.Statuses))
		for _, status := range filters.Statuses {
			statuses = append(statuses, string(status))
		}
		predicates = append(predicates, table.Status.In(statuses...))
	}
	if filters.CreatedFrom != nil {
		predicates = append(predicates, table.CreatedAt.Gte(*filters.CreatedFrom))
	}
	if filters.CreatedTo != nil {
		predicates = append(predicates, table.CreatedAt.Lte(*filters.CreatedTo))
	}
	return predicates
}

func (r *webhookRecordRepository) List(ctx context.Context, req *biz.WebhookRecordListRequest) ([]*model.WebhookRecord, error) {
	db := r.repository.DB(ctx)
	conditions := webhookRecordPredicates(&webhookRecordPredicatesRequest{
		DB:      db,
		Filters: req.WebhookRecordFilters,
	})
	return db.WebhookRecord.WithContext(ctx).
		Preload(db.WebhookRecord.Account).
		Where(conditions...).
		Order(db.WebhookRecord.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *webhookRecordRepository) Count(ctx context.Context, req *biz.WebhookRecordCountRequest) (int64, error) {
	db := r.repository.DB(ctx)
	conditions := webhookRecordPredicates(&webhookRecordPredicatesRequest{
		DB:      db,
		Filters: req.WebhookRecordFilters,
	})
	return db.WebhookRecord.WithContext(ctx).Where(conditions...).Count()
}

func (r *webhookRecordRepository) Exist(ctx context.Context, req *biz.WebhookRecordExistRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.WebhookRecord.WithContext(ctx).
		Where(
			db.WebhookRecord.Channel.Eq(string(enums.Channel_Slash)),
			db.WebhookRecord.AccountID.Eq(req.AccountID),
			db.WebhookRecord.ID.Eq(req.ID),
		).
		Count()
	return count > 0, err
}

func (r *webhookRecordRepository) Find(ctx context.Context, req *biz.WebhookRecordFindRequest) (*model.WebhookRecord, error) {
	db := r.repository.DB(ctx)
	return db.WebhookRecord.WithContext(ctx).
		Preload(db.WebhookRecord.Account).
		Where(
			db.WebhookRecord.Channel.Eq(string(enums.Channel_Slash)),
			db.WebhookRecord.AccountID.Eq(req.AccountID),
			db.WebhookRecord.ID.Eq(req.ID),
		).
		First()
}
