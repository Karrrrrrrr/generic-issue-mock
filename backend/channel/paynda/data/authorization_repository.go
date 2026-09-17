package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
	"gorm.io/gorm/clause"
)

type authorizationRepository struct {
	repository *PayndaRepository
}

var _ biz.PayndaAuthorizationRepository = (*authorizationRepository)(nil)

func NewAuthorizationRepository(injector *do.Injector) (biz.PayndaAuthorizationRepository, error) {
	return &authorizationRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *authorizationRepository) Create(ctx context.Context, item *model.Authorization) error {
	return r.repository.DB(ctx).Authorization.WithContext(ctx).Create(item)
}

func (r *authorizationRepository) FindByID(
	ctx context.Context,
	req *biz.PayndaFindAuthorizationRequest,
) (*model.Authorization, error) {
	db := r.repository.DB(ctx)

	return db.Authorization.WithContext(ctx).
		Preload(db.Authorization.Account).
		Where(
			db.Authorization.ID.Eq(req.ID),
			db.Authorization.Channel.Eq(string(enums.Channel_Paynda)),
		).First()
}

func (r *authorizationRepository) List(ctx context.Context, req *biz.AuthorizationListRequest) ([]*model.Authorization, error) {
	db := r.repository.DB(ctx)
	query := db.Authorization.WithContext(ctx).
		Preload(db.Authorization.Account).
		Where(db.Authorization.Channel.Eq(string(enums.Channel_Paynda)))
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.Authorization.AccountID.In(req.AccountIDs...))
	}
	return query.Order(db.Authorization.ID.Desc()).Offset(req.Offset).Limit(req.Limit).Find()
}

func (r *authorizationRepository) ListByIDs(
	ctx context.Context,
	req *biz.PayndaListAuthorizationsByIDsRequest,
) ([]*model.Authorization, error) {
	db := r.repository.DB(ctx)

	query := db.Authorization.WithContext(ctx).
		Where(db.Authorization.Channel.Eq(string(enums.Channel_Paynda)))
	if len(req.IDs) != 0 {
		query = query.Where(db.Authorization.ID.In(req.IDs...))
	}
	return query.
		Preload(db.Authorization.Account).
		Order(db.Authorization.ID.Desc()).
		Find()
}

func (r *authorizationRepository) AuthorizationExists(ctx context.Context, req *biz.ExistAuthorizationRequest) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Authorization.WithContext(ctx).Where(
		db.Authorization.ID.Eq(req.ID),
		db.Authorization.AccountID.Eq(req.AccountID),
		db.Authorization.Channel.Eq(string(enums.Channel_Paynda)),
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
		db.Authorization.Channel.Eq(string(enums.Channel_Paynda)),
	).First()
}

func (r *authorizationRepository) ListAuthorizations(ctx context.Context, req *biz.AuthorizationListBalancesRequest) ([]*model.Authorization, error) {
	db := r.repository.DB(ctx)
	query := db.Authorization.WithContext(ctx).
		Preload(db.Authorization.Account).
		Where(
			db.Authorization.Channel.Eq(string(enums.Channel_Paynda)),
		)
	if len(req.AccountIDs) != 0 {
		query = query.Where(db.Authorization.AccountID.In(req.AccountIDs...))
	}
	return query.Order(db.Authorization.ID.Desc()).Find()
}
