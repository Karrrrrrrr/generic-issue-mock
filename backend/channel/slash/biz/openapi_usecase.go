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
	Source      model.ID
	Destination model.ID
	AmountCents int64
}

func (u *SlashOpenAPIUsecase) ListVirtualAccounts(ctx context.Context) ([]*model.VirtualAccount, error) {
	items, err := u.virtualAccountRepository.List(ctx)
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
			exists, err := u.virtualAccountRepository.ExistByID(txCtx, id)
			if err != nil {
				zap.S().Errorw("check slash virtual account", "error", err)
				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}
		}
		source, err := u.virtualAccountRepository.FindByID(txCtx, req.Source)
		if err != nil {
			zap.S().Errorw("find slash source account", "error", err)
			return ErrDatabaseOperation
		}
		destination, err := u.virtualAccountRepository.FindByID(txCtx, req.Destination)
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
			wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, id)
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
	Offset int
	Limit  int
	Status enums.CardStatus
}

func (u *SlashOpenAPIUsecase) ListCards(ctx context.Context, req *OpenAPIListCardsRequest) ([]*model.Card, error) {
	items, err := u.cardRepository.List(ctx, &ListCardsRequest{
		Offset: req.Offset,
		Limit:  req.Limit,
		Status: req.Status,
	})
	if err != nil {
		zap.S().Errorw("list slash openapi cards", "error", err)

		return nil, ErrDatabaseOperation
	}

	return items, nil
}

type OpenAPICreateCardRequest struct {
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

		productExists, err := u.cardProductRepository.ExistByID(txCtx, req.CardProductID)
		if err != nil {
			zap.S().Errorw("check slash openapi card product", "error", err)

			return ErrDatabaseOperation
		}
		if !productExists {
			return ErrResourceNotFound
		}

		product, err := u.cardProductRepository.FindByIDForUpdate(txCtx, req.CardProductID)
		if err != nil {
			zap.S().Errorw("lock slash openapi card product", "error", err)

			return ErrDatabaseOperation
		}

		product.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(product.Prefix, product.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance slash openapi card product sequence", "error", err)

			return ErrDatabaseOperation
		}

		card = &model.Card{
			Channel:                enums.Channel_Slash,
			CardProductID:          product.ID,
			CardBin:                product.Prefix,
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

func (u *SlashOpenAPIUsecase) GetCard(ctx context.Context, id model.ID) (*model.Card, error) {
	exists, err := u.cardRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash openapi card", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item, err := u.cardRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash openapi card", "error", err)

		return nil, ErrDatabaseOperation
	}

	return item, nil
}

type OpenAPIUpdateCardRequest struct {
	ID     model.ID
	Status enums.CardStatus
}

func (u *SlashOpenAPIUsecase) UpdateCard(ctx context.Context, req *OpenAPIUpdateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistByID(txCtx, req.ID)
		if err != nil {
			zap.S().Errorw("check slash openapi card", "error", err)

			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		card, err = u.cardRepository.FindByID(txCtx, req.ID)
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

func (u *SlashOpenAPIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	items, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list slash openapi card products", "error", err)

		return nil, ErrDatabaseOperation
	}

	return items, nil
}

type OpenAPIListTransactionsRequest struct {
	Offset          int
	Limit           int
	CardID          model.ID
	AuthorizationID model.ID
}

func (u *SlashOpenAPIUsecase) ListTransactions(ctx context.Context, req *OpenAPIListTransactionsRequest) ([]*model.CardTransaction, error) {
	items, err := u.cardTransactionRepository.List(ctx, &ListCardTransactionsRequest{
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

func (u *SlashOpenAPIUsecase) GetTransaction(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	exists, err := u.cardTransactionRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash openapi transaction", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	item, err := u.cardTransactionRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find slash openapi transaction", "error", err)

		return nil, ErrDatabaseOperation
	}

	return item, nil
}
