package biz

import (
	"context"
	"crypto/rand"
	"time"

	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Transaction interface {
	InTx(context.Context, func(context.Context) error) error
}

const defaultCardBinPrefix = "424242"

type CardHolderRepository interface {
	Create(context.Context, *model.CardHolder) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardHolder, error)
	Count(context.Context, *ListCardHoldersRequest) (int64, error)
	List(context.Context, *ListCardHoldersRequest) ([]*model.CardHolder, error)
	Save(context.Context, *model.CardHolder) error
}

type CardRepository interface {
	Create(context.Context, *model.Card) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.Card, error)
	Count(context.Context, *ListCardsRequest) (int64, error)
	List(context.Context, *ListCardsRequest) ([]*model.Card, error)
	Save(context.Context, *model.Card) error
}

type CardProductRepository interface {
	Create(context.Context, *model.CardProduct) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardProduct, error)
	ExistDefault(context.Context) (bool, error)
	FindDefault(context.Context) (*model.CardProduct, error)
	List(context.Context) ([]*model.CardProduct, error)
}

type AuthorizationRepository interface {
	Create(context.Context, *model.Authorization) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.Authorization, error)
	Count(context.Context, *ListAuthorizationsRequest) (int64, error)
	List(context.Context, *ListAuthorizationsRequest) ([]*model.Authorization, error)
	Save(context.Context, *model.Authorization) error
}

type CardTransactionRepository interface {
	Create(context.Context, *model.CardTransaction) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardTransaction, error)
	Count(context.Context, *ListCardTransactionsRequest) (int64, error)
	List(context.Context, *ListCardTransactionsRequest) ([]*model.CardTransaction, error)
}

type Usecase struct {
	transaction               Transaction
	cardHolderRepository      CardHolderRepository
	cardRepository            CardRepository
	cardProductRepository     CardProductRepository
	authorizationRepository   AuthorizationRepository
	cardTransactionRepository CardTransactionRepository
}

func NewUsecase(injector *do.Injector) (*Usecase, error) {
	return &Usecase{
		transaction:               do.MustInvokeNamed[Transaction](injector, "slash.transaction"),
		cardHolderRepository:      do.MustInvokeNamed[CardHolderRepository](injector, "slash.card-holder-repository"),
		cardRepository:            do.MustInvokeNamed[CardRepository](injector, "slash.card-repository"),
		cardProductRepository:     do.MustInvokeNamed[CardProductRepository](injector, "slash.card-product-repository"),
		authorizationRepository:   do.MustInvokeNamed[AuthorizationRepository](injector, "slash.authorization-repository"),
		cardTransactionRepository: do.MustInvokeNamed[CardTransactionRepository](injector, "slash.card-transaction-repository"),
	}, nil
}

type ListCardHoldersRequest struct {
	Offset int
	Limit  int
}

type ListCardsRequest struct {
	Offset     int
	Limit      int
	IDContains string
	CardNumber string
	Status     enums.CardStatus
}

type ListAuthorizationsRequest struct {
	Offset int
	Limit  int
	ID     model.ID
	CardID model.ID
	Status enums.CardTransactionStatus
}

type ListCardTransactionsRequest struct {
	Offset          int
	Limit           int
	ID              model.ID
	CardID          model.ID
	AuthorizationID model.ID
	Type            enums.CardTransactionType
	Status          enums.CardTransactionStatus
}

type CreateCardHolderRequest struct {
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (u *Usecase) CreateCardHolder(ctx context.Context, req *CreateCardHolderRequest) (*model.CardHolder, error) {
	holder := &model.CardHolder{
		Channel:      enums.Channel_Slash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Mobile:       req.Mobile,
		Status:       enums.CardHolderStatus_Normal,
		ReviewStatus: enums.CardHolderReviewStatus_Approved,
		Shared:       true,
	}
	if err := u.cardHolderRepository.Create(ctx, holder); err != nil {
		zap.S().Errorw("create slash card holder", "error", err)
		return nil, ErrDatabaseOperation
	}

	return holder, nil
}

func (u *Usecase) ListCardHolders(ctx context.Context, req *ListCardHoldersRequest) ([]*model.CardHolder, int64, error) {
	items, err := u.cardHolderRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list slash card holders", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.cardHolderRepository.Count(ctx, req)
	if err != nil {
		zap.S().Errorw("count slash card holders", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

type CreateCardRequest struct {
	CardHolderID  model.ID
	CardProductID model.ID
	Currency      enums.Currency
}

func (u *Usecase) CreateCard(ctx context.Context, req *CreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCardHolder(txCtx, req.CardHolderID); err != nil {
			return err
		}

		product, err := u.getCardProduct(txCtx, req.CardProductID)
		if err != nil {
			return err
		}
		if len(product.Prefix) >= 16 {
			return ErrInvalidOperation
		}

		card = &model.Card{
			Channel:                enums.Channel_Slash,
			CardProductID:          product.ID,
			CardBin:                product.Prefix,
			CardNumber:             product.Prefix + randomDigits(16-len(product.Prefix)),
			Cvv:                    randomDigits(3),
			ExpireTime:             time.Now().UTC().AddDate(2, 0, 0).Format("01/06"),
			Status:                 enums.CardStatus_Active,
			CardHolderID:           req.CardHolderID,
			FormType:               enums.CardFormType_Virtual,
			CardCurrency:           req.Currency,
			CardScheme:             "VISA",
			CardType:               enums.CardType_Single,
			RequestID:              randomDigits(20),
			LastOperationRequestID: randomDigits(20),
		}
		if err := u.cardRepository.Create(txCtx, card); err != nil {
			zap.S().Errorw("create slash card", "error", err)
			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *Usecase) ListCards(ctx context.Context, req *ListCardsRequest) ([]*model.Card, int64, error) {
	items, err := u.cardRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list slash cards", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.cardRepository.Count(ctx, req)
	if err != nil {
		zap.S().Errorw("count slash cards", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

type CardProductInfo struct {
	Product *model.CardProduct
}

func (u *Usecase) ListCardProducts(ctx context.Context) ([]*CardProductInfo, error) {
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		_, err := u.getCardProduct(txCtx, "")
		return err
	})
	if err != nil {
		return nil, err
	}
	products, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list slash card products", "error", err)
		return nil, ErrDatabaseOperation
	}
	items := make([]*CardProductInfo, 0, len(products))
	for _, product := range products {
		items = append(items, &CardProductInfo{
			Product: product,
		})
	}

	return items, nil
}

func (u *Usecase) GetCard(ctx context.Context, id model.ID) (*model.Card, error) {
	if err := u.requireCard(ctx, id); err != nil {
		return nil, err
	}
	card, err := u.cardRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash card", "error", err)
		return nil, ErrDatabaseOperation
	}

	return card, nil
}

type UpdateCardStatusRequest struct {
	ID     model.ID
	Status enums.CardStatus
}

func (u *Usecase) UpdateCardStatus(ctx context.Context, req *UpdateCardStatusRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCard(txCtx, req.ID); err != nil {
			return err
		}
		var err error
		card, err = u.cardRepository.FindByID(txCtx, req.ID)
		if err != nil {
			zap.S().Errorw("find slash card", "error", err)
			return ErrDatabaseOperation
		}
		card.Status = req.Status
		if err := u.cardRepository.Save(txCtx, card); err != nil {
			zap.S().Errorw("update slash card status", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

type SimulateAuthorizationRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

type SimulateAuthorizationResult struct {
	Authorization   *model.Authorization
	CardTransaction *model.CardTransaction
}

func (u *Usecase) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*SimulateAuthorizationResult, error) {
	var result *SimulateAuthorizationResult
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.GetCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}
		now := time.Now().UTC()
		authorization := &model.Authorization{
			Channel:           enums.Channel_Slash,
			CardID:            card.ID,
			Currency:          req.Currency,
			Amount:            req.Amount,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: randomDigits(6),
			Status:            enums.TransactionStatus_AUTHORIZED,
			OccurredAt:        now,
		}
		if err := u.authorizationRepository.Create(txCtx, authorization); err != nil {
			zap.S().Errorw("create slash authorization", "error", err)
			return ErrDatabaseOperation
		}
		transaction := &model.CardTransaction{
			Channel:           enums.Channel_Slash,
			AuthorizationID:   authorization.ID,
			CardID:            card.ID,
			Status:            enums.TransactionStatus_AUTHORIZED,
			Type:              enums.CardTransactionType_AUTH,
			Currency:          req.Currency,
			TxAmount:          req.Amount,
			TxCurrency:        req.Currency,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: authorization.AuthorizationCode,
			OccurredAt:        now,
		}
		if err := u.cardTransactionRepository.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create slash authorization transaction", "error", err)
			return ErrDatabaseOperation
		}
		result = &SimulateAuthorizationResult{
			Authorization:   authorization,
			CardTransaction: transaction,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *Usecase) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) ([]*model.Authorization, int64, error) {
	items, err := u.authorizationRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list slash authorizations", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.authorizationRepository.Count(ctx, req)
	if err != nil {
		zap.S().Errorw("count slash authorizations", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *Usecase) GetAuthorization(ctx context.Context, id model.ID) (*model.Authorization, error) {
	if err := u.requireAuthorization(ctx, id); err != nil {
		return nil, err
	}
	item, err := u.authorizationRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash authorization", "error", err)
		return nil, ErrDatabaseOperation
	}

	return item, nil
}

func (u *Usecase) ListCardTransactions(ctx context.Context, req *ListCardTransactionsRequest) ([]*model.CardTransaction, int64, error) {
	items, err := u.cardTransactionRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list slash card transactions", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.cardTransactionRepository.Count(ctx, req)
	if err != nil {
		zap.S().Errorw("count slash card transactions", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *Usecase) GetCardTransaction(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	if err := u.requireCardTransaction(ctx, id); err != nil {
		return nil, err
	}
	item, err := u.cardTransactionRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash card transaction", "error", err)
		return nil, ErrDatabaseOperation
	}

	return item, nil
}

type ApplyTransactionStepRequest struct {
	CardTransactionID model.ID
	Type              enums.CardTransactionType
	Amount            decimal.Decimal
}

func (u *Usecase) ApplyTransactionStep(ctx context.Context, req *ApplyTransactionStepRequest) (*model.CardTransaction, error) {
	var next *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		origin, err := u.GetCardTransaction(txCtx, req.CardTransactionID)
		if err != nil {
			return err
		}
		amount := req.Amount
		if amount.IsZero() {
			amount = origin.TxAmount
		}
		if !amount.IsPositive() || amount.GreaterThan(origin.TxAmount) {
			return ErrInvalidOperation
		}
		status := enums.TransactionStatus_SUCCEED
		if req.Type == enums.CardTransactionType_VOID {
			status = enums.TransactionStatus_VOID
		}
		next = &model.CardTransaction{
			Channel:                 enums.Channel_Slash,
			OriginCardTransactionID: origin.ID,
			AuthorizationID:         origin.AuthorizationID,
			CardID:                  origin.CardID,
			Status:                  status,
			Type:                    req.Type,
			Currency:                origin.Currency,
			TxAmount:                amount,
			TxCurrency:              origin.TxCurrency,
			MerchantName:            origin.MerchantName,
			MerchantCountry:         origin.MerchantCountry,
			MerchantMCC:             origin.MerchantMCC,
			AuthorizationCode:       origin.AuthorizationCode,
			OccurredAt:              time.Now().UTC(),
		}
		if err := u.cardTransactionRepository.Create(txCtx, next); err != nil {
			zap.S().Errorw("create slash card transaction step", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return next, nil
}

func (u *Usecase) requireCardHolder(ctx context.Context, id model.ID) error {
	exists, err := u.cardHolderRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card holder", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	return nil
}

func (u *Usecase) requireCard(ctx context.Context, id model.ID) error {
	exists, err := u.cardRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	return nil
}

func (u *Usecase) requireAuthorization(ctx context.Context, id model.ID) error {
	exists, err := u.authorizationRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash authorization", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	return nil
}

func (u *Usecase) requireCardTransaction(ctx context.Context, id model.ID) error {
	exists, err := u.cardTransactionRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card transaction", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	return nil
}

func (u *Usecase) getCardProduct(ctx context.Context, id model.ID) (*model.CardProduct, error) {
	if id == "" {
		exists, err := u.cardProductRepository.ExistDefault(ctx)
		if err != nil {
			zap.S().Errorw("check slash default card product", "error", err)
			return nil, ErrDatabaseOperation
		}
		if !exists {
			product := &model.CardProduct{
				Channel:   enums.Channel_Slash,
				Prefix:    defaultCardBinPrefix,
				IsDefault: true,
			}
			if err := u.cardProductRepository.Create(ctx, product); err != nil {
				zap.S().Errorw("create slash default card product", "error", err)
				return nil, ErrDatabaseOperation
			}

			return product, nil
		}

		product, err := u.cardProductRepository.FindDefault(ctx)
		if err != nil {
			zap.S().Errorw("find slash default card product", "error", err)
			return nil, ErrDatabaseOperation
		}

		return product, nil
	}

	if err := u.requireCardProduct(ctx, id); err != nil {
		return nil, err
	}
	product, err := u.cardProductRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash card product", "error", err)
		return nil, ErrDatabaseOperation
	}

	return product, nil
}

func (u *Usecase) requireCardProduct(ctx context.Context, id model.ID) error {
	exists, err := u.cardProductRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card product", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}

func randomDigits(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	for index := range bytes {
		bytes[index] = '0' + bytes[index]%10
	}
	return string(bytes)
}
