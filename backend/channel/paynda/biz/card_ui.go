package biz

import (
	"context"
	"time"

	payndaerrors "generic-mock/channel/paynda/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type ListUICardsRequest struct {
	AccountID   *model.ID
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Offset      int
	Limit       int
	ID          *model.ID
	CardNumber  *string
	Statuses    []enums.CardStatus
}

type PayndaUICreateCardRequest struct {
	AccountID    model.ID
	CardHolderID model.ID
	Currency     enums.Currency
}

type PayndaUIUpdateCardStatusRequest struct {
	AccountID model.ID
	CardID    model.ID
	Status    enums.CardStatus
}

func (u *PayndaUIUsecase) CreateCard(ctx context.Context, req *PayndaUICreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.accountRepository.ExistByID(txCtx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check paynda UI card account", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		if !exists {
			return payndaerrors.ErrResourceNotFound
		}
		exists, err = u.cardHolderRepository.ExistByAccountID(txCtx, &CardHolderExistByAccountIDRequest{
			AccountID: req.AccountID,
			ID:        req.CardHolderID,
		})
		if err != nil {
			zap.S().Errorw("check paynda UI card holder", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		if !exists {
			return payndaerrors.ErrResourceNotFound
		}
		products, err := u.cardProductRepository.List(txCtx)
		if err != nil {
			zap.S().Errorw("list paynda UI card products", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		if len(products) == 0 {
			return payndaerrors.ErrResourceNotFound
		}

		var defaultProduct *model.CardProduct
		for _, item := range products {
			if item.IsDefault {
				defaultProduct = item
				break
			}
		}
		if defaultProduct == nil {
			return payndaerrors.ErrResourceNotFound
		}

		product, err := u.cardProductRepository.FindByIDForUpdate(txCtx, defaultProduct.ID)
		if err != nil {
			zap.S().Errorw("lock paynda UI card product", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		product.NextCardNumber++
		generatedCard, ok := cardnumber.Generate(cardnumber.GenerateRequest{
			Channel:  enums.Channel_Paynda,
			Prefix:   product.Prefix,
			Sequence: product.NextCardNumber,
		})
		if !ok {
			return payndaerrors.ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance paynda UI card product sequence", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: req.AccountID,
			Channel:   enums.Channel_Paynda,
			Available: decimal.Zero,
			Type:      enums.WalletType_Card,
			Currency:  req.Currency,
		}
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create paynda UI card wallet", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		card = &model.Card{
			AccountID:     req.AccountID,
			Channel:       enums.Channel_Paynda,
			CardProductID: product.ID,
			CardBin:       generatedCard.Bin,
			CardNumber:    generatedCard.Number,
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
			return payndaerrors.ErrDatabaseOperation
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
	if req.AccountID <= 0 || req.CardID <= 0 {
		return nil, payndaerrors.ErrInvalidOperation
	}
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistForStatusChange(txCtx, &CardStatusExistsRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("check paynda card status change", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		if !exists {
			return payndaerrors.ErrResourceNotFound
		}
		card, err = u.cardRepository.LockForStatusChange(txCtx, &CardStatusLockRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("lock paynda card status change", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		nextStatus := req.Status
		if (card.Status == enums.CardStatus_Deleted && nextStatus != enums.CardStatus_Deleted) ||
			(card.Status == enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleted) {
			return payndaerrors.ErrCardClosed
		}
		card.Status = nextStatus
		if err := u.cardRepository.SaveStatus(txCtx, &CardStatusSaveRequest{
			AccountID:              card.AccountID,
			ID:                     card.ID,
			Status:                 card.Status,
			LastOperationRequestID: card.LastOperationRequestID,
			LastOperationType:      card.LastOperationType,
			LastOperationStatus:    card.LastOperationStatus,
		}); err != nil {
			zap.S().Errorw("update paynda UI card status", "error", err)
			return payndaerrors.ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (u *PayndaUIUsecase) ListCards(ctx context.Context, req *ListUICardsRequest) ([]*model.Card, int64, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, 0, payndaerrors.ErrInvalidOperation
	}
	items, err := u.cardRepository.List(ctx, &CardListRequest{
		AccountIDs:  types.PointerSlice(req.AccountID),
		IDs:         types.PointerSlice(req.ID),
		Statuses:    req.Statuses,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		CardNumber:  req.CardNumber,
		Limit:       req.Limit,
		Offset:      req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list paynda UI cards", "error", err)
		return nil, 0, payndaerrors.ErrDatabaseOperation
	}

	total, err := u.cardRepository.Count(ctx, &CardCountRequest{
		AccountIDs:  types.PointerSlice(req.AccountID),
		IDs:         types.PointerSlice(req.ID),
		Statuses:    req.Statuses,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		CardNumber:  req.CardNumber,
	})
	if err != nil {
		zap.S().Errorw("count paynda UI cards", "error", err)
		return nil, 0, payndaerrors.ErrDatabaseOperation
	}

	return items, total, nil
}
