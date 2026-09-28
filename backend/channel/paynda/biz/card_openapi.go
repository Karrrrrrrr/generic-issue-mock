package biz

import (
	"context"
	"time"

	paynda "generic-mock/channel/paynda/enums"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type PayndaRequestLookup struct {
	AccountID model.ID
	RequestID string
}

type PayndaCreateCardRequest struct {
	AccountID     model.ID
	CardHolderID  model.ID
	CardProductID model.ID
	Currency      enums.Currency
	ExpireAt      time.Time
	RequestID     string
}

type PayndaUpdateCardStatusRequest struct {
	AccountID model.ID
	CardID    model.ID
	RequestID string
	Status    paynda.CardStatus
}

type PayndaRequestResult struct {
	Card         *model.Card
	Transaction  *model.CardTransaction
	IsCardCreate bool
}

func (u *PayndaOpenAPIUsecase) CreateCard(ctx context.Context, req *PayndaCreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.accountRepository.ExistByID(txCtx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check paynda card account", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		holderResource := &PayndaResourceRequest{
			AccountID: req.AccountID,
			ID:        req.CardHolderID,
		}
		if err := u.requireCardHolder(txCtx, holderResource); err != nil {
			return err
		}
		if err := u.requireCardProduct(txCtx, req.CardProductID); err != nil {
			return err
		}

		product, err := u.cardProductRepository.FindByIDForUpdate(txCtx, req.CardProductID)
		if err != nil {
			zap.S().Errorw("lock paynda card product", "error", err)
			return ErrDatabaseOperation
		}
		product.NextCardNumber++
		generatedCard, ok := cardnumber.Generate(cardnumber.GenerateRequest{
			Channel:  enums.Channel_Paynda,
			Prefix:   product.Prefix,
			Sequence: product.NextCardNumber,
		})
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance paynda card product sequence", "error", err)
			return ErrDatabaseOperation
		}

		wallet := &model.Wallet{
			AccountID: req.AccountID,
			Channel:   enums.Channel_Paynda,
			Available: decimal.Zero,
			Type:      enums.WalletType_Card,
			Currency:  req.Currency,
		}
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create paynda card wallet", "error", err)
			return ErrDatabaseOperation
		}

		card = &model.Card{
			Channel:                enums.Channel_Paynda,
			AccountID:              req.AccountID,
			CardProductID:          product.ID,
			CardBin:                generatedCard.Bin,
			CardNumber:             generatedCard.Number,
			Cvv:                    randomx.Digits(3),
			ExpireAt:               req.ExpireAt,
			Status:                 enums.CardStatus_Active,
			WalletID:               wallet.ID,
			CardHolderID:           req.CardHolderID,
			FormType:               enums.CardFormType_Virtual,
			RequestID:              req.RequestID,
			LastOperationRequestID: req.RequestID,
			LastOperationType:      enums.OperationType_OpenCard,
			LastOperationStatus:    enums.OperationStatus_Succeed,
			CardCurrency:           req.Currency,
			CardScheme:             enums.CardScheme_MasterCard,
			CardType:               enums.CardType_Single,
		}
		if err := u.cardRepository.Create(txCtx, card); err != nil {
			zap.S().Errorw("create paynda card", "error", err)
			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PayndaOpenAPIUsecase) GetCard(ctx context.Context, req *PayndaResourceRequest) (*model.Card, error) {
	if err := u.requireCard(ctx, req); err != nil {
		return nil, err
	}

	card, err := u.cardRepository.FindByID(ctx, &CardFindByIDRequest{
		AccountID: &req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find paynda card", "error", err)
		return nil, ErrDatabaseOperation
	}

	return card, nil
}

func (u *PayndaOpenAPIUsecase) ListCards(ctx context.Context, req *PayndaListRequest) ([]*model.Card, error) {
	items, err := u.cardRepository.List(ctx, &CardListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list paynda cards", "error", err)
		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaOpenAPIUsecase) GetCardBalance(ctx context.Context, req *PayndaResourceRequest) (*model.Wallet, error) {
	card, err := u.GetCard(ctx, req)
	if err != nil {
		return nil, err
	}
	if card.WalletID == 0 {
		return nil, ErrResourceNotFound
	}

	walletRequest := &PayndaResourceRequest{
		AccountID: req.AccountID,
		ID:        card.WalletID,
	}
	exists, err := u.walletRepository.ExistByID(ctx, &WalletExistByIDRequest{
		AccountID: &walletRequest.AccountID,
		ID:        walletRequest.ID,
	})
	if err != nil {
		zap.S().Errorw("check paynda card wallet", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	wallet, err := u.walletRepository.FindByID(ctx, &WalletFindByIDRequest{
		AccountID: &walletRequest.AccountID,
		ID:        walletRequest.ID,
	})
	if err != nil {
		zap.S().Errorw("find paynda card wallet", "error", err)
		return nil, ErrDatabaseOperation
	}

	return wallet, nil
}

func (u *PayndaOpenAPIUsecase) UpdateCardStatus(
	ctx context.Context,
	req *PayndaUpdateCardStatusRequest,
) (*model.Card, error) {
	if req.AccountID <= 0 || req.CardID <= 0 {
		return nil, ErrInvalidOperation
	}
	if req.RequestID == "" {
		req.RequestID = randomx.Digits(20)
	}

	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistForStatusChange(txCtx, &CardStatusExistsRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("check paynda card status change", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		card, err = u.cardRepository.LockForStatusChange(txCtx, &CardStatusLockRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			zap.S().Errorw("lock paynda card status change", "error", err)
			return ErrDatabaseOperation
		}
		nextStatus := paynda.CardStatusToGeneric(req.Status)
		if (card.Status == enums.CardStatus_Deleted && nextStatus != enums.CardStatus_Deleted) ||
			(card.Status == enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleteing && nextStatus != enums.CardStatus_Deleted) {
			return ErrCardClosed
		}
		card.Status = nextStatus
		card.LastOperationRequestID = req.RequestID
		card.LastOperationType = enums.OperationType_UpdateCard
		card.LastOperationStatus = enums.OperationStatus_Succeed
		if err := u.cardRepository.SaveStatus(txCtx, &CardStatusSaveRequest{
			AccountID:              card.AccountID,
			ID:                     card.ID,
			Status:                 card.Status,
			LastOperationRequestID: card.LastOperationRequestID,
			LastOperationType:      card.LastOperationType,
			LastOperationStatus:    card.LastOperationStatus,
		}); err != nil {
			zap.S().Errorw("update paynda card status", "error", err)
			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PayndaOpenAPIUsecase) ReleaseCard(ctx context.Context, req *PayndaUpdateCardStatusRequest) (*model.Card, error) {
	req.Status = paynda.CardStatus_Deleted
	return u.UpdateCardStatus(ctx, req)
}

func (u *PayndaOpenAPIUsecase) ListCardBalanceUpdates(
	ctx context.Context,
	req *PayndaListRequest,
) ([]*model.CardTransaction, error) {
	items, err := u.cardTransactionRepository.List(ctx, &CardTransactionListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Offset:     req.Offset,
		Limit:      req.Limit,
		Types: []enums.CardTransactionType{
			enums.CardTransactionType_FundIn,
			enums.CardTransactionType_FundOut,
		},
	})
	if err != nil {
		zap.S().Errorw("list paynda card balance updates", "error", err)
		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaOpenAPIUsecase) FindRequestResult(
	ctx context.Context,
	req *PayndaRequestLookup,
) (*PayndaRequestResult, error) {
	exists, err := u.cardRepository.ExistByRequestID(ctx, (*CardExistByRequestIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check paynda create card request", "error", err)
		return nil, ErrDatabaseOperation
	}
	if exists {
		card, err := u.cardRepository.FindByRequestID(ctx, (*CardFindByRequestIDRequest)(req))
		if err != nil {
			zap.S().Errorw("find paynda create card request", "error", err)
			return nil, ErrDatabaseOperation
		}
		return &PayndaRequestResult{
			Card:         card,
			IsCardCreate: true,
		}, nil
	}

	exists, err = u.cardRepository.ExistByLastOperationRequestID(ctx, (*CardExistByLastOperationRequestIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check paynda card operation request", "error", err)
		return nil, ErrDatabaseOperation
	}
	if exists {
		card, err := u.cardRepository.FindByLastOperationRequestID(ctx, (*CardFindByLastOperationRequestIDRequest)(req))
		if err != nil {
			zap.S().Errorw("find paynda card operation request", "error", err)
			return nil, ErrDatabaseOperation
		}
		return &PayndaRequestResult{Card: card}, nil
	}

	exists, err = u.cardTransactionRepository.ExistByRequestID(ctx, (*CardTransactionExistByRequestIDRequest)(req))
	if err != nil {
		zap.S().Errorw("check paynda card balance transfer request", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	transaction, err := u.cardTransactionRepository.FindByRequestID(ctx, (*CardTransactionFindByRequestIDRequest)(req))
	if err != nil {
		zap.S().Errorw("find paynda card balance transfer request", "error", err)
		return nil, ErrDatabaseOperation
	}

	return &PayndaRequestResult{Transaction: transaction}, nil
}

func (u *PayndaOpenAPIUsecase) requireCard(ctx context.Context, req *PayndaResourceRequest) error {
	exists, err := u.cardRepository.ExistByID(ctx, &CardExistByIDRequest{
		AccountID: &req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check paynda card", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}
