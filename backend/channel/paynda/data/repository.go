package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	"generic-mock/internal/query"
	"generic-mock/model"
	"generic-mock/pkg/gormx"

	"github.com/samber/do"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PayndaRepository struct {
	db *gorm.DB
}

func NewPayndaRepository(injector *do.Injector) (*PayndaRepository, error) {
	return &PayndaRepository{
		db: do.MustInvoke[*gorm.DB](injector),
	}, nil
}

func (r *PayndaRepository) DB(ctx context.Context) *query.Query {
	return query.Use(gormx.DB(ctx, r.db))
}

type cardHolderRepository struct {
	repository *PayndaRepository
}

func NewCardHolderRepository(injector *do.Injector) (biz.PayndaCardHolderRepository, error) {
	return &cardHolderRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *cardHolderRepository) Create(ctx context.Context, item *model.CardHolder) error {
	return r.repository.DB(ctx).CardHolder.WithContext(ctx).Create(item)
}

func (r *cardHolderRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardHolder.WithContext(ctx).Where(
		db.CardHolder.ID.Eq(id),
		db.CardHolder.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *cardHolderRepository) FindByID(ctx context.Context, id model.ID) (*model.CardHolder, error) {
	db := r.repository.DB(ctx)

	return db.CardHolder.WithContext(ctx).Where(
		db.CardHolder.ID.Eq(id),
		db.CardHolder.Channel.Eq(string(enums.Channel_Paynda)),
	).First()
}

func (r *cardHolderRepository) List(
	ctx context.Context,
	req *biz.PayndaListRequest,
) ([]*model.CardHolder, error) {
	db := r.repository.DB(ctx)

	return db.CardHolder.WithContext(ctx).
		Where(db.CardHolder.Channel.Eq(string(enums.Channel_Paynda))).
		Order(db.CardHolder.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *cardHolderRepository) Save(ctx context.Context, item *model.CardHolder) error {
	return r.repository.DB(ctx).CardHolder.WithContext(ctx).Save(item)
}

var _ biz.PayndaCardHolderRepository = (*cardHolderRepository)(nil)

type cardProductRepository struct {
	repository *PayndaRepository
}

func NewCardProductRepository(injector *do.Injector) (biz.PayndaCardProductRepository, error) {
	return &cardProductRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *cardProductRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardProduct.WithContext(ctx).Where(
		db.CardProduct.ID.Eq(id),
		db.CardProduct.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *cardProductRepository) FindByIDForUpdate(
	ctx context.Context,
	id model.ID,
) (*model.CardProduct, error) {
	db := r.repository.DB(ctx)

	return db.CardProduct.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			db.CardProduct.ID.Eq(id),
			db.CardProduct.Channel.Eq(string(enums.Channel_Paynda)),
		).
		First()
}

func (r *cardProductRepository) List(ctx context.Context) ([]*model.CardProduct, error) {
	db := r.repository.DB(ctx)

	return db.CardProduct.WithContext(ctx).
		Where(db.CardProduct.Channel.Eq(string(enums.Channel_Paynda))).
		Order(db.CardProduct.ID.Desc()).
		Find()
}

func (r *cardProductRepository) Save(ctx context.Context, item *model.CardProduct) error {
	return r.repository.DB(ctx).CardProduct.WithContext(ctx).Save(item)
}

var _ biz.PayndaCardProductRepository = (*cardProductRepository)(nil)

type cardRepository struct {
	repository *PayndaRepository
}

func NewCardRepository(injector *do.Injector) (biz.PayndaCardRepository, error) {
	return &cardRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *cardRepository) Create(ctx context.Context, item *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Create(item)
}

func (r *cardRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(id),
		db.Card.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *cardRepository) ExistByRequestID(ctx context.Context, requestID string) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).Where(
		db.Card.RequestID.Eq(requestID),
		db.Card.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *cardRepository) ExistByLastOperationRequestID(
	ctx context.Context,
	requestID string,
) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).Where(
		db.Card.LastOperationRequestID.Eq(requestID),
		db.Card.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *cardRepository) FindByID(ctx context.Context, id model.ID) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).Where(
		db.Card.ID.Eq(id),
		db.Card.Channel.Eq(string(enums.Channel_Paynda)),
	).First()
}

func (r *cardRepository) FindByRequestID(ctx context.Context, requestID string) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).Where(
		db.Card.RequestID.Eq(requestID),
		db.Card.Channel.Eq(string(enums.Channel_Paynda)),
	).First()
}

func (r *cardRepository) FindByLastOperationRequestID(
	ctx context.Context,
	requestID string,
) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).Where(
		db.Card.LastOperationRequestID.Eq(requestID),
		db.Card.Channel.Eq(string(enums.Channel_Paynda)),
	).First()
}

func (r *cardRepository) List(ctx context.Context, req *biz.PayndaListRequest) ([]*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Where(db.Card.Channel.Eq(string(enums.Channel_Paynda))).
		Order(db.Card.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func (r *cardRepository) Save(ctx context.Context, item *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Save(item)
}

var _ biz.PayndaCardRepository = (*cardRepository)(nil)

type walletRepository struct {
	repository *PayndaRepository
}

func NewWalletRepository(injector *do.Injector) (biz.PayndaWalletRepository, error) {
	return &walletRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *walletRepository) Create(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Create(item)
}

func (r *walletRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Wallet.WithContext(ctx).Where(db.Wallet.ID.Eq(id)).Count()

	return count > 0, err
}

func (r *walletRepository) FindByID(ctx context.Context, id model.ID) (*model.Wallet, error) {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Where(
		r.repository.DB(ctx).Wallet.ID.Eq(id),
	).First()
}

func (r *walletRepository) FindByIDForUpdate(ctx context.Context, id model.ID) (*model.Wallet, error) {
	db := r.repository.DB(ctx)

	return db.Wallet.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(db.Wallet.ID.Eq(id)).
		First()
}

func (r *walletRepository) Save(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Save(item)
}

var _ biz.PayndaWalletRepository = (*walletRepository)(nil)

type accountRepository struct {
	repository *PayndaRepository
}

func NewAccountRepository(injector *do.Injector) (biz.PayndaAccountRepository, error) {
	return &accountRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *accountRepository) Create(ctx context.Context, item *model.Account) error {
	return r.repository.DB(ctx).Account.WithContext(ctx).Create(item)
}

func (r *accountRepository) ExistByChannel(ctx context.Context) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Account.WithContext(ctx).Where(
		db.Account.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *accountRepository) FindByChannel(ctx context.Context) (*model.Account, error) {
	db := r.repository.DB(ctx)

	return db.Account.WithContext(ctx).Where(
		db.Account.Channel.Eq(string(enums.Channel_Paynda)),
	).Order(db.Account.ID.Desc()).First()
}

var _ biz.PayndaAccountRepository = (*accountRepository)(nil)

type cardTransactionRepository struct {
	repository *PayndaRepository
}

func NewCardTransactionRepository(injector *do.Injector) (biz.PayndaCardTransactionRepository, error) {
	return &cardTransactionRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *cardTransactionRepository) Create(ctx context.Context, item *model.CardTransaction) error {
	return r.repository.DB(ctx).CardTransaction.WithContext(ctx).Create(item)
}

func (r *cardTransactionRepository) Save(ctx context.Context, item *model.CardTransaction) error {
	return r.repository.DB(ctx).CardTransaction.WithContext(ctx).Save(item)
}

func (r *cardTransactionRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(id),
		db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *cardTransactionRepository) ExistByRequestID(ctx context.Context, requestID string) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.RequestID.Eq(requestID),
		db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
	).Count()

	return count > 0, err
}

func (r *cardTransactionRepository) FindByID(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	db := r.repository.DB(ctx)

	return db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(id),
		db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
	).First()
}

func (r *cardTransactionRepository) FindByRequestID(
	ctx context.Context,
	requestID string,
) (*model.CardTransaction, error) {
	db := r.repository.DB(ctx)

	return db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.RequestID.Eq(requestID),
		db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
	).First()
}

func (r *cardTransactionRepository) List(
	ctx context.Context,
	req *biz.PayndaListTransactionsRequest,
) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	predicates := payndaTransactionPredicates(db, req)

	return db.CardTransaction.WithContext(ctx).
		Where(predicates...).
		Order(db.CardTransaction.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

func payndaTransactionPredicates(
	db *query.Query,
	req *biz.PayndaListTransactionsRequest,
) []gen.Condition {
	predicates := []gen.Condition{
		db.CardTransaction.Channel.Eq(string(enums.Channel_Paynda)),
	}
	if req.CardID != 0 {
		predicates = append(predicates, db.CardTransaction.CardID.Eq(req.CardID))
	}
	if req.OccurredAtGTE != nil {
		predicates = append(predicates, db.CardTransaction.OccurredAt.Gte(*req.OccurredAtGTE))
	}
	if req.OccurredAtLTE != nil {
		predicates = append(predicates, db.CardTransaction.OccurredAt.Lte(*req.OccurredAtLTE))
	}
	if len(req.Types) > 0 {
		types := make([]string, 0, len(req.Types))
		for _, item := range req.Types {
			types = append(types, string(item))
		}
		predicates = append(predicates, db.CardTransaction.Type.In(types...))
	}

	return predicates
}

var _ biz.PayndaCardTransactionRepository = (*cardTransactionRepository)(nil)

type authorizationRepository struct {
	repository *PayndaRepository
}

func NewAuthorizationRepository(injector *do.Injector) (biz.PayndaAuthorizationRepository, error) {
	return &authorizationRepository{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (r *authorizationRepository) Create(ctx context.Context, item *model.Authorization) error {
	return r.repository.DB(ctx).Authorization.WithContext(ctx).Create(item)
}

var _ biz.PayndaAuthorizationRepository = (*authorizationRepository)(nil)

type transaction struct {
	repository *PayndaRepository
}

func NewPayndaTransaction(injector *do.Injector) (biz.PayndaTransaction, error) {
	return &transaction{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (t *transaction) InTx(ctx context.Context, fn func(context.Context) error) error {
	return gormx.InTx(ctx, t.repository.db, fn)
}

var _ biz.PayndaTransaction = (*transaction)(nil)
