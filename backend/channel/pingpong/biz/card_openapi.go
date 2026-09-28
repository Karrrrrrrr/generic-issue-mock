package biz

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/cardwallet"
	"generic-mock/pkg/randomx"
	sharedbiz "generic-mock/shared/biz"

	"go.uber.org/zap"
)

type CreateCardRequest struct {
	Notificator      sharedbiz.Notificator
	AccountID        model.ID
	VirtualAccountID model.ID
	ProductID        model.ID
	Currency         common.Currency
	RequestID        string
	RawRequest       []byte
}

type GetCardRequest struct {
	AccountID model.ID
	ID        model.ID
}

type ChangeCardRequest struct {
	AccountID model.ID
	ID        model.ID
	Status    common.CardStatus
}

func (uc *PingPongOpenAPIUsecase) CreateCard(ctx context.Context, req *CreateCardRequest) (*model.Card, error) {
	if req.RequestID == "" || req.Currency != common.Currency_USD || req.ProductID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	var card *model.Card
	created := false
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		account, err := uc.lockAccount(ctx, req.AccountID)
		if err != nil {
			return err
		}
		exists, err := uc.cardRepo.ExistsRequest(ctx, &CardRequestExistsRequest{
			AccountID: req.AccountID,
			RequestID: req.RequestID,
		})
		if err != nil {
			zap.S().Errorw("check pingpong card request", "error", err)
			return pingerrors.ErrDatabase
		}
		if exists {
			card, err = uc.cardRepo.FindRequest(ctx, &CardRequestFindRequest{
				AccountID: req.AccountID,
				RequestID: req.RequestID,
			})
			if err != nil {
				zap.S().Errorw("find pingpong card request", "error", err)
				return pingerrors.ErrDatabase
			}
			var savedRequest, currentRequest any
			if json.Unmarshal(card.RawRequest, &savedRequest) != nil || json.Unmarshal(req.RawRequest, &currentRequest) != nil || !reflect.DeepEqual(savedRequest, currentRequest) {
				return pingerrors.ErrConflict
			}
			return nil
		}
		virtualAccount, err := uc.getVirtualAccount(ctx, &openAPIVirtualAccountReference{
			AccountID: req.AccountID,
			ID:        req.VirtualAccountID,
		})
		if err != nil {
			return err
		}
		exists, err = uc.cardProductRepo.Exists(ctx, &ProductExistsRequest{ID: req.ProductID})
		if err != nil {
			zap.S().Errorw("check pingpong product", "error", err)
			return pingerrors.ErrDatabase
		}
		if !exists {
			return pingerrors.ErrNotFound
		}
		product, err := uc.cardProductRepo.Lock(ctx, &ProductLockRequest{ID: req.ProductID})
		if err != nil {
			zap.S().Errorw("lock pingpong product", "error", err)
			return pingerrors.ErrDatabase
		}
		product.NextCardNumber++
		number, ok := cardnumber.Generate(cardnumber.GenerateRequest{
			Channel:  common.Channel_PingPong,
			Prefix:   product.Prefix,
			Sequence: product.NextCardNumber,
		})
		if !ok {
			return pingerrors.ErrInvalid
		}
		assignment, ok := cardwallet.Prepare(cardwallet.PrepareRequest{
			AccountID:      req.AccountID,
			Channel:        common.Channel_PingPong,
			CardType:       common.CardType_VirtualAccountSingle,
			Currency:       req.Currency,
			VirtualAccount: virtualAccount,
		})
		if !ok {
			return pingerrors.ErrInvalid
		}
		if err := uc.cardProductRepo.Save(ctx, product); err != nil {
			zap.S().Errorw("advance pingpong card sequence", "error", err)
			return pingerrors.ErrDatabase
		}
		if err := uc.walletRepo.Create(ctx, assignment.Wallet); err != nil {
			zap.S().Errorw("create pingpong card wallet", "error", err)
			return pingerrors.ErrDatabase
		}
		card = &model.Card{
			AccountID:        req.AccountID,
			Channel:          common.Channel_PingPong,
			CardProductID:    product.ID,
			CardBin:          number.Bin,
			CardNumber:       number.Number,
			Cvv:              randomx.Digits(3),
			ExpireAt:         time.Now().UTC().AddDate(2, 0, 0),
			Status:           common.CardStatus_Active,
			VirtualAccountID: assignment.VirtualAccountID,
			WalletID:         assignment.Wallet.ID,
			FormType:         common.CardFormType_Virtual,
			RequestID:        req.RequestID,
			CardCurrency:     req.Currency,
			CardScheme:       common.CardScheme_Visa,
			CardType:         common.CardType_VirtualAccountSingle,
			RawRequest:       req.RawRequest,
			Account:          account,
			Wallet:           assignment.Wallet,
			VirtualAccount:   virtualAccount,
		}
		if err := uc.cardRepo.Create(ctx, card); err != nil {
			zap.S().Errorw("create pingpong card", "error", err)
			return pingerrors.ErrDatabase
		}
		created = true
		return nil
	})
	if err == nil && created && req.Notificator != nil {
		_ = req.Notificator.NotifyIssueCard(ctx, &sharedbiz.NotifyIssueCardReq{AccountID: card.AccountID, Channel: common.Channel_PingPong, CardID: card.ID})
	}
	return card, err
}

func (uc *PingPongOpenAPIUsecase) GetCard(ctx context.Context, req *GetCardRequest) (*model.Card, error) {
	if req.AccountID <= 0 || req.ID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.cardRepo.Exists(ctx, &CardExistsRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong card", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	card, err := uc.cardRepo.Find(ctx, &CardFindRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong card", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return card, nil
}

func (uc *PingPongOpenAPIUsecase) ChangeCard(ctx context.Context, req *ChangeCardRequest) error {
	if req.Status != common.CardStatus_Active && req.Status != common.CardStatus_Frozen && req.Status != common.CardStatus_Deleted {
		return pingerrors.ErrInvalid
	}
	return uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockAccount(ctx, req.AccountID); err != nil {
			return err
		}
		card, err := uc.GetCard(ctx, &GetCardRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			return err
		}
		if card.Status == common.CardStatus_Deleted && req.Status != card.Status {
			return pingerrors.ErrClosed
		}
		card.Status = req.Status
		if err := uc.cardRepo.Save(ctx, card); err != nil {
			zap.S().Errorw("change pingpong card status", "error", err)
			return pingerrors.ErrDatabase
		}
		return nil
	})
}
