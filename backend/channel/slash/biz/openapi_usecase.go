package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type SlashOpenAPIUsecase struct {
	transaction               SlashTransaction
	cardHolderRepository      SlashCardHolderRepository
	cardRepository            SlashCardRepository
	cardProductRepository     SlashCardProductRepository
	cardTransactionRepository SlashCardTransactionRepository
	virtualAccountRepository  SlashVirtualAccountRepository
	walletRepository          SlashWalletRepository
}

func NewSlashOpenAPIUsecase(injector *do.Injector) (*SlashOpenAPIUsecase, error) {
	return &SlashOpenAPIUsecase{
		transaction:               do.MustInvoke[SlashTransaction](injector),
		cardHolderRepository:      do.MustInvoke[SlashCardHolderRepository](injector),
		cardRepository:            do.MustInvoke[SlashCardRepository](injector),
		cardProductRepository:     do.MustInvoke[SlashCardProductRepository](injector),
		cardTransactionRepository: do.MustInvoke[SlashCardTransactionRepository](injector),
		virtualAccountRepository:  do.MustInvoke[SlashVirtualAccountRepository](injector),
		walletRepository:          do.MustInvoke[SlashWalletRepository](injector),
	}, nil
}

type OpenAPIVirtualAccountTransferRequest struct {
	AccountID   model.ID
	Source      model.ID
	Destination model.ID
	AmountCents int64
}

func (u *SlashOpenAPIUsecase) ListVirtualAccounts(ctx context.Context, accountID model.ID) ([]*model.VirtualAccount, error) {
	items, err := u.virtualAccountRepository.ListByAccountID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("list slash virtual accounts", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

func (u *SlashOpenAPIUsecase) TransferVirtualAccount(ctx context.Context, req *OpenAPIVirtualAccountTransferRequest) error {
	if req.Source == req.Destination || req.AmountCents <= 0 {
		return ErrInvalidOperation
	}
	return u.transaction.InTx(ctx, func(txCtx context.Context) error {
		for _, id := range []model.ID{req.Source, req.Destination} {
			resource := &ResourceRequest{AccountID: &req.AccountID, ID: id}
			exists, err := u.virtualAccountRepository.ExistByAccountID(txCtx, resource)
			if err != nil {
				zap.S().Errorw("check slash virtual account", "error", err)
				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}
		}
		source, err := u.virtualAccountRepository.FindByAccountID(txCtx, &ResourceRequest{AccountID: &req.AccountID, ID: req.Source})
		if err != nil {
			zap.S().Errorw("find slash source account", "error", err)
			return ErrDatabaseOperation
		}
		destination, err := u.virtualAccountRepository.FindByAccountID(txCtx, &ResourceRequest{AccountID: &req.AccountID, ID: req.Destination})
		if err != nil {
			zap.S().Errorw("find slash destination account", "error", err)
			return ErrDatabaseOperation
		}
		first, second := source.WalletID, destination.WalletID
		if first > second {
			first, second = second, first
		}
		locked := make(map[model.ID]*model.Wallet, 2)
		for _, id := range []model.ID{first, second} {
			wallet, err := u.walletRepository.FindByAccountIDForUpdate(txCtx, &ResourceRequest{AccountID: &req.AccountID, ID: id})
			if err != nil {
				zap.S().Errorw("lock slash virtual account wallet", "error", err)
				return ErrDatabaseOperation
			}
			locked[id] = wallet
		}
		amount := decimal.NewFromInt(req.AmountCents).Div(decimal.NewFromInt(100))
		sourceWallet, destinationWallet := locked[source.WalletID], locked[destination.WalletID]
		if sourceWallet.Amount.LessThan(amount) {
			return ErrInvalidOperation
		}
		sourceWallet.Amount = sourceWallet.Amount.Sub(amount)
		destinationWallet.Amount = destinationWallet.Amount.Add(amount)
		if err := u.walletRepository.Save(txCtx, sourceWallet); err != nil {
			zap.S().Errorw("save slash source wallet", "error", err)
			return ErrDatabaseOperation
		}
		if err := u.walletRepository.Save(txCtx, destinationWallet); err != nil {
			zap.S().Errorw("save slash destination wallet", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
}

type OpenAPIListCardsRequest struct {
	AccountID model.ID
	Offset    int
	Limit     int
	Status    enums.CardStatus
}

func (u *SlashOpenAPIUsecase) ListCards(ctx context.Context, req *OpenAPIListCardsRequest) ([]*model.Card, error) {
	items, err := u.cardRepository.List(ctx, &ListCardsRequest{
		AccountID: req.AccountID,
		Offset:    req.Offset,
		Limit:     req.Limit,
		Status:    req.Status,
	})
	if err != nil {
		zap.S().Errorw("list slash openapi cards", "error", err)

		return nil, ErrDatabaseOperation
	}

	return items, nil
}

type OpenAPICreateCardRequest struct {
	AccountID     model.ID
	CardHolderID  model.ID
	CardProductID model.ID
	Currency      enums.Currency
	RequestID     string
}

func (u *SlashOpenAPIUsecase) CreateCard(ctx context.Context, req *OpenAPICreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if req.CardHolderID != 0 {
			holderExists, err := u.cardHolderRepository.ExistByID(txCtx, req.CardHolderID)
			if err != nil {
				zap.S().Errorw("check slash openapi card holder", "error", err)

				return ErrDatabaseOperation
			}
			if !holderExists {
				return ErrResourceNotFound
			}
		}

		product := &ResourceRequest{AccountID: &req.AccountID, ID: req.CardProductID}
		productExists, err := u.cardProductRepository.ExistByAccountID(txCtx, product)
		if err != nil {
			zap.S().Errorw("check slash openapi card product", "error", err)

			return ErrDatabaseOperation
		}
		if !productExists {
			return ErrResourceNotFound
		}

		cardProduct, err := u.cardProductRepository.FindByAccountIDForUpdate(txCtx, product)
		if err != nil {
			zap.S().Errorw("lock slash openapi card product", "error", err)

			return ErrDatabaseOperation
		}

		cardProduct.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(cardProduct.Prefix, cardProduct.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, cardProduct); err != nil {
			zap.S().Errorw("advance slash openapi card product sequence", "error", err)

			return ErrDatabaseOperation
		}

		card = &model.Card{
			Channel:                enums.Channel_Slash,
			AccountID:              req.AccountID,
			CardProductID:          cardProduct.ID,
			CardBin:                cardProduct.Prefix,
			CardNumber:             cardNumber,
			Cvv:                    randomx.Digits(3),
			ExpireAt:               time.Now().UTC().AddDate(2, 0, 0),
			Status:                 enums.CardStatus_Active,
			CardHolderID:           req.CardHolderID,
			FormType:               enums.CardFormType_Virtual,
			CardCurrency:           req.Currency,
			CardScheme:             enums.CardScheme_Visa,
			CardType:               enums.CardType_Single,
			RequestID:              req.RequestID,
			LastOperationRequestID: req.RequestID,
		}
		if err := u.cardRepository.Create(txCtx, card); err != nil {
			zap.S().Errorw("create slash openapi card", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *SlashOpenAPIUsecase) GetCard(ctx context.Context, req *ResourceRequest) (*model.Card, error) {
	exists, err := u.cardRepository.ExistByAccountID(ctx, req)
	if err != nil {
		zap.S().Errorw("check slash openapi card", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item, err := u.cardRepository.FindByAccountID(ctx, req)
	if err != nil {
		zap.S().Errorw("find slash openapi card", "error", err)

		return nil, ErrDatabaseOperation
	}

	return item, nil
}

type OpenAPIUpdateCardRequest struct {
	AccountID model.ID
	ID        model.ID
	Status    enums.CardStatus
}

func (u *SlashOpenAPIUsecase) UpdateCard(ctx context.Context, req *OpenAPIUpdateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		resource := &ResourceRequest{AccountID: &req.AccountID, ID: req.ID}
		exists, err := u.cardRepository.ExistByAccountID(txCtx, resource)
		if err != nil {
			zap.S().Errorw("check slash openapi card", "error", err)

			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		card, err = u.cardRepository.FindByAccountID(txCtx, resource)
		if err != nil {
			zap.S().Errorw("find slash openapi card", "error", err)

			return ErrDatabaseOperation
		}

		card.Status = req.Status
		if err := u.cardRepository.Save(txCtx, card); err != nil {
			zap.S().Errorw("update slash openapi card", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *SlashOpenAPIUsecase) ListCardProducts(ctx context.Context, accountID model.ID) ([]*model.CardProduct, error) {
	items, err := u.cardProductRepository.ListByAccountID(ctx, accountID)
	if err != nil {
		zap.S().Errorw("list slash openapi card products", "error", err)

		return nil, ErrDatabaseOperation
	}

	return items, nil
}

type OpenAPIListTransactionsRequest struct {
	AccountID       model.ID
	Offset          int
	Limit           int
	CardID          model.ID
	AuthorizationID model.ID
}

func (u *SlashOpenAPIUsecase) ListTransactions(ctx context.Context, req *OpenAPIListTransactionsRequest) ([]*model.CardTransaction, error) {
	items, err := u.cardTransactionRepository.List(ctx, &ListCardTransactionsRequest{
		AccountID:       req.AccountID,
		Offset:          req.Offset,
		Limit:           req.Limit,
		CardID:          req.CardID,
		AuthorizationID: req.AuthorizationID,
	})
	if err != nil {
		zap.S().Errorw("list slash openapi transactions", "error", err)

		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *SlashOpenAPIUsecase) GetTransaction(ctx context.Context, req *ResourceRequest) (*model.CardTransaction, error) {
	exists, err := u.cardTransactionRepository.ExistByAccountID(ctx, req)
	if err != nil {
		zap.S().Errorw("check slash openapi transaction", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item, err := u.cardTransactionRepository.FindByAccountID(ctx, req)
	if err != nil {
		zap.S().Errorw("find slash openapi transaction", "error", err)

		return nil, ErrDatabaseOperation
	}

	return item, nil
}
