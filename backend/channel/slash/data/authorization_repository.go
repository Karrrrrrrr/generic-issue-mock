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

type authorizationRepository struct {
	repository *SlashRepository
}

func NewAuthorizationRepository(injector *do.Injector) (biz.SlashAuthorizationRepository, error) {
	return &authorizationRepository{
		repository: do.MustInvoke[*SlashRepository](injector),
	}, nil
}

func (r *authorizationRepository) Create(ctx context.Context, item *model.Authorization) error {
	return r.repository.DB(ctx).Authorization.WithContext(ctx).Create(item)
}

func (r *authorizationRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Authorization.WithContext(ctx).Where(
		db.Authorization.ID.Eq(id),
		db.Authorization.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
	return count > 0, err
}

func (r *authorizationRepository) FindByID(ctx context.Context, id model.ID) (*model.Authorization, error) {
	db := r.repository.DB(ctx)
	return db.Authorization.WithContext(ctx).Where(
		db.Authorization.ID.Eq(id),
		db.Authorization.Channel.Eq(string(enums.Channel_Slash)),
	).First()
}

func (r *authorizationRepository) Count(ctx context.Context, req *biz.ListAuthorizationsRequest) (int64, error) {
	db := r.repository.DB(ctx)
	return db.Authorization.WithContext(ctx).
		Where(authorizationPredicates(db, req)...).
		Count()
}

func (r *authorizationRepository) List(ctx context.Context, req *biz.ListAuthorizationsRequest) ([]*model.Authorization, error) {
	db := r.repository.DB(ctx)
	return db.Authorization.WithContext(ctx).
		Where(authorizationPredicates(db, req)...).
		Order(db.Authorization.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *authorizationRepository) Save(ctx context.Context, item *model.Authorization) error {
	return r.repository.DB(ctx).Authorization.WithContext(ctx).Save(item)
}

func authorizationPredicates(db *query.Query, req *biz.ListAuthorizationsRequest) []gen.Condition {
	predicates := make([]gen.Condition, 0, 4)
	predicates = append(predicates, db.Authorization.Channel.Eq(string(enums.Channel_Slash)))
	if req.ID != 0 {
		predicates = append(predicates, db.Authorization.ID.Eq(req.ID))
	}
	if req.CardID != 0 {
		predicates = append(predicates, db.Authorization.CardID.Eq(req.CardID))
	}
	if req.Status != "" {
		predicates = append(predicates, db.Authorization.Status.Eq(string(req.Status)))
	}
	return predicates
}

var _ biz.SlashAuthorizationRepository = (*authorizationRepository)(nil)
