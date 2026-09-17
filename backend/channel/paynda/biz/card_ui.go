package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type PayndaUICreateCardRequest struct {
	AccountID    model.ID
	CardHolderID model.ID
	Currency     enums.Currency
}

type PayndaUIUpdateCardStatusRequest struct {
	CardID model.ID
	Status enums.CardStatus
}

func (u *PayndaUIUsecase) CreateCard(ctx context.Context, req *PayndaUICreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.accountRepository.ExistByID(txCtx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check paynda UI card account", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		exists, err = u.cardHolderRepository.ExistByAccountID(txCtx, &CardHolderExistByAccountIDRequest{
			AccountID: req.AccountID,
			ID:        req.CardHolderID,
		})
		if err != nil {
			zap.S().Errorw("check paynda UI card holder", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		products, err := u.cardProductRepository.List(txCtx)
		if err != nil {
			zap.S().Errorw("list paynda UI card products", "error", err)
			return ErrDatabaseOperation
		}
		if len(products) == 0 {
			return ErrResourceNotFound
		}

		var defaultProduct *model.CardProduct
		for _, item := range products {
			if item.IsDefault {
				defaultProduct = item
				break
			}
		}
		if defaultProduct == nil {
			return ErrResourceNotFound
		}

		product, err := u.cardProductRepository.FindByIDForUpdate(txCtx, defaultProduct.ID)
		if err != nil {
			zap.S().Errorw("lock paynda UI card product", "error", err)
			return ErrDatabaseOperation
		}
		product.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(product.Prefix, product.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance paynda UI card product sequence", "error", err)
			return ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: req.AccountID,
			Channel:   enums.Channel_Paynda,
			Amount:    decimal.Zero,
			Type:      enums.WalletType_Card,
			Currency:  req.Currency,
		}
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create paynda UI card wallet", "error", err)
			return ErrDatabaseOperation
		}
		card = &model.Card{
			AccountID:     req.AccountID,
			Channel:       enums.Channel_Paynda,
			CardProductID: product.ID,
			CardBin:       product.Prefix,
			CardNumber:    cardNumber,
			Cvv:           randomx.Digits(3),
			ExpireAt:      time.Now().UTC().AddDate(2, 0, 0),
			Status:        enums.CardStatus_Active,
			WalletID:      wallet.ID,
			CardHolderID:  req.CardHolderID,
			FormType:      enums.CardFormType_Virtual,
			CardCurrency:  req.Currency,
			CardScheme:    enums.CardScheme_MasterCard,
			CardType:      enums.CardType_Single,
		}
		if err := u.cardRepository.Create(txCtx, card); err != nil {
			zap.S().Errorw("create paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	u.dispatchCardStatus(ctx, card)
	return card, nil
}

func (u *PayndaUIUsecase) UpdateCardStatus(ctx context.Context, req *PayndaUIUpdateCardStatusRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistByID(txCtx, &CardExistByIDRequest{ID: req.CardID})
		if err != nil {
			zap.S().Errorw("check paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		card, err = u.cardRepository.FindByID(txCtx, &CardFindByIDRequest{ID: req.CardID})
		if err != nil {
			zap.S().Errorw("find paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		card.Status = req.Status
		if err := u.cardRepository.Save(txCtx, card); err != nil {
			zap.S().Errorw("update paynda UI card status", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (u *PayndaUIUsecase) ListCards(ctx context.Context, req *PayndaListRequest) ([]*model.Card, error) {
	items, err := u.cardRepository.List(ctx, &CardListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list paynda UI cards", "error", err)
		return nil, ErrDatabaseOperation
	}

	return items, nil
}
