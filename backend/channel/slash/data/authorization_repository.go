package data

import (
	"context"
	"gorm.io/gorm/clause"

	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	"generic-mock/internal/query"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do/v2"
	"gorm.io/gen"
)

type authorizationRepository struct {
	repository *SlashRepository
}

func NewAuthorizationRepository(injector do.Injector) (biz.SlashAuthorizationRepository, error) {
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
	return db.Authorization.WithContext(ctx).
		Preload(db.Authorization.Account).
		Where(
			db.Authorization.ID.Eq(id),
			db.Authorization.Channel.Eq(string(enums.Channel_Slash)),
		).First()
}

func (r *authorizationRepository) Count(ctx context.Context, req *biz.AuthorizationCountRequest) (int64, error) {
	db := r.repository.DB(ctx)
	return db.Authorization.WithContext(ctx).
		Where(authorizationPredicates(db, (*biz.AuthorizationListRequest)(req))...).
		Count()
}

func (r *authorizationRepository) List(ctx context.Context, req *biz.AuthorizationListRequest) ([]*model.Authorization, error) {
	db := r.repository.DB(ctx)
	return db.Authorization.WithContext(ctx).
		Preload(db.Authorization.Account).
		Where(authorizationPredicates(db, req)...).
		Order(db.Authorization.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *authorizationRepository) Save(ctx context.Context, item *model.Authorization) error {
	return r.repository.DB(ctx).Authorization.WithContext(ctx).Save(item)
}

func authorizationPredicates(db *query.Query, req *biz.AuthorizationListRequest) []gen.Condition {
	predicates := make([]gen.Condition, 0, 4)
	predicates = append(predicates, db.Authorization.Channel.Eq(string(enums.Channel_Slash)))
	if len(req.AccountIDs) != 0 {
		predicates = append(predicates, db.Authorization.AccountID.In(req.AccountIDs...))
	}
	if len(req.IDs) != 0 {
		predicates = append(predicates, db.Authorization.ID.In(req.IDs...))
	}
	if len(req.CardIDs) != 0 {
		predicates = append(predicates, db.Authorization.CardID.In(req.CardIDs...))
	}
	if len(req.Statuses) != 0 {
		predicates = append(predicates, db.Authorization.Status.In(types.BulkConvertSlice(req.Statuses, func(value enums.CardTransactionStatus) string {
			return string(value)
		})...))
	}
	return predicates
}

var _ biz.SlashAuthorizationRepository = (*authorizationRepository)(nil)

func (r *authorizationRepository) AuthorizationExists(ctx context.Context, req *biz.ExistAuthorizationRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Authorization.WithContext(ctx).Where(
		db.Authorization.ID.Eq(req.ID),
		db.Authorization.AccountID.Eq(req.AccountID),
		db.Authorization.Channel.Eq(string(enums.Channel_Slash)),
	).Count()
	return count > 0, err
}

func (r *authorizationRepository) LockAuthorization(ctx context.Context, req *biz.LockAuthorizationRequest) (*model.Authorization, error) {
	db := r.repository.DB(ctx)
	return db.Authorization.WithContext(ctx).
		Preload(db.Authorization.Account).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).Where(
		db.Authorization.ID.Eq(req.ID),
		db.Authorization.AccountID.Eq(req.AccountID),
		db.Authorization.Channel.Eq(string(enums.Channel_Slash)),
	).First()
}

func (r *authorizationRepository) ListAuthorizations(ctx context.Context, req *biz.AuthorizationListBalancesRequest) ([]*model.Authorization, error) {
	db := r.repository.DB(ctx)
	query := db.Authorization.WithContext(ctx).
		Preload(db.Authorization.Account).
		Where(
			db.Authorization.Channel.Eq(string(enums.Channel_Slash)),
		)
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.Authorization.AccountID.In(req.AccountIDs...))
	}
	if len(req.IDs) != 0 {
		query = query.Where(db.Authorization.ID.In(req.IDs...))
	}
	if len(req.CardIDs) != 0 {
		query = query.Where(db.Authorization.CardID.In(req.CardIDs...))
	}
	if len(req.Statuses) != 0 {
		query = query.Where(db.Authorization.Status.In(types.BulkConvertSlice(req.Statuses, func(value enums.CardTransactionStatus) string {
			return string(value)
		})...))
	}
	if req.MerchantName != nil {
		query = query.Where(db.Authorization.MerchantName.Like("%" + *req.MerchantName + "%"))
	}
	if req.CreatedFrom != nil {
		query = query.Where(db.Authorization.CreatedAt.Gte(*req.CreatedFrom))
	}
	if req.CreatedTo != nil {
		query = query.Where(db.Authorization.CreatedAt.Lte(*req.CreatedTo))
	}
	return query.Order(db.Authorization.ID.Desc()).Find()
}

func (r *authorizationRepository) FindAuthorizationDetail(ctx context.Context, req *biz.FindAuthorizationDetailRequest) (*model.Authorization, error) {
	db := r.repository.DB(ctx)
	return db.Authorization.WithContext(ctx).
		Preload(db.Authorization.Account).
		Where(
			db.Authorization.ID.Eq(req.ID),
			db.Authorization.AccountID.Eq(req.AccountID),
			db.Authorization.Channel.Eq(string(enums.Channel_Slash)),
		).First()
}
