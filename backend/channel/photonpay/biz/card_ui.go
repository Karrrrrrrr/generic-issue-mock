package biz

import (
	"context"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

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

type UIOpenCardRequest struct {
	AccountID    model.ID
	CardHolderID model.ID
	Currency     enums.Currency
	RequestID    string
}

type UIChangeCardStatusRequest struct {
	CardID model.ID
	Status enums.CardStatus
}

func (u *PhotonPayUIUsecase) OpenCard(ctx context.Context, req *UIOpenCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardHolderRepo.ExistCardHolderByAccountID(txCtx, &CardHolderExistCardHolderByAccountIDRequest{
			AccountID: &req.AccountID,
			ID:        req.CardHolderID,
		})
		if err != nil {
			zap.S().Errorw("check photonpay UI card holder", "error", err)

			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		productExists, err := u.cardProductRepo.ExistByPrefix(txCtx, photon.DefaultCardBin)
		if err != nil {
			zap.S().Errorw("check photonpay UI card product", "error", err)

			return ErrDatabaseOperation
		}
		if !productExists {
			return ErrResourceNotFound
		}

		product, err := u.cardProductRepo.FindByPrefixForUpdate(txCtx, photon.DefaultCardBin)
		if err != nil {
			zap.S().Errorw("lock photonpay UI card product", "error", err)

			return ErrDatabaseOperation
		}

		product.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(product.Prefix, product.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepo.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance photonpay UI card product sequence", "error", err)

			return ErrDatabaseOperation
		}
		virtualAccount, err := u.virtualAccountRepo.FindByAccountID(txCtx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay UI account virtual account", "error", err)
			return ErrDatabaseOperation
		}

		card = &model.Card{
			AccountID:              req.AccountID,
			Channel:                enums.Channel_PhotonPay,
			CardProductID:          product.ID,
			CardBin:                product.Prefix,
			CardNumber:             cardNumber,
			Cvv:                    randomx.Digits(3),
			ExpireAt:               time.Now().UTC().AddDate(0, 24, 0),
			Status:                 enums.CardStatus_Active,
			VirtualAccountID:       &virtualAccount.ID,
			WalletID:               virtualAccount.WalletID,
			CardHolderID:           req.CardHolderID,
			FormType:               enums.CardFormType_Virtual,
			RequestID:              req.RequestID,
			LastOperationRequestID: req.RequestID,
			LastOperationType:      enums.OperationType_OpenCard,
			LastOperationStatus:    enums.OperationStatus_Succeed,
			CardCurrency:           req.Currency,
			CardScheme:             photon.CardScheme,
			CardType:               enums.CardType_Share,
		}
		if err := u.cardRepo.CreateCard(txCtx, card); err != nil {
			zap.S().Errorw("create photonpay UI card", "error", err)

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

func (u *PhotonPayUIUsecase) ListCards(ctx context.Context, req *ListUICardsRequest) ([]*model.Card, int64, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, 0, ErrInvalidOperation
	}
	cards, err := u.cardRepo.ListCards(ctx, &CardListCardsRequest{
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
		zap.S().Errorw("list photonpay UI cards", "error", err)

		return nil, 0, ErrDatabaseOperation
	}

	total, err := u.cardRepo.Count(ctx, &CardCountRequest{
		AccountIDs:  types.PointerSlice(req.AccountID),
		IDs:         types.PointerSlice(req.ID),
		Statuses:    req.Statuses,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		CardNumber:  req.CardNumber,
	})
	if err != nil {
		zap.S().Errorw("count photonpay UI cards", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return cards, total, nil
}

func (u *PhotonPayUIUsecase) ChangeCardStatus(ctx context.Context, req *UIChangeCardStatusRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepo.ExistCardByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("check photonpay UI card", "error", err)

			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		card, err = u.cardRepo.FindCardByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("find photonpay UI card", "error", err)

			return ErrDatabaseOperation
		}

		card.Status = req.Status
		if err := u.cardRepo.SaveCard(txCtx, card); err != nil {
			zap.S().Errorw("update photonpay UI card status", "error", err)

			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PhotonPayUIUsecase) getCard(ctx context.Context, id model.ID) (*model.Card, error) {
	exists, err := u.cardRepo.ExistCardByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay UI card", "error", err)

		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	card, err := u.cardRepo.FindCardByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find photonpay UI card", "error", err)

		return nil, ErrDatabaseOperation
	}

	return card, nil
}
