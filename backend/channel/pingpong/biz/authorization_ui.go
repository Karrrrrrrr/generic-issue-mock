package biz

import (
	"context"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type SimulateAuthorizationRequest struct {
	AccountID    model.ID
	CardID       model.ID
	Amount       decimal.Decimal
	Currency     common.Currency
	RequestID    string
	MerchantName string
}

type AuthorizationStageRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
	RequestID string
	Kind      common.CardTransactionType
}

type ListAuthorizationsRequest struct {
	AccountID *model.ID
	CardID    *model.ID
	Offset    int
	Limit     int
}

func (uc *PingPongUIUsecase) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*model.Authorization, error) {
	if !req.Amount.IsPositive() || req.RequestID == "" {
		return nil, pingerrors.ErrInvalid
	}
	var authorization *model.Authorization
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockAccount(ctx, req.AccountID); err != nil {
			return err
		}
		card, err := uc.getCard(ctx, &uiCardReference{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			return err
		}
		previous, err := uc.previousStage(ctx, &stageLookup{
			AccountID: req.AccountID,
			RequestID: req.RequestID,
		})
		if err != nil {
			return err
		}
		if previous != nil {
			if previous.Type != common.CardTransactionType_AUTH || previous.CardID != card.ID || previous.Currency != req.Currency || !previous.TxAmount.Equal(req.Amount) || previous.MerchantName != req.MerchantName {
				return pingerrors.ErrConflict
			}
			authorization, err = uc.authorization(ctx, &authorizationReference{
				AccountID: req.AccountID,
				ID:        previous.AuthorizationID,
			})
			return err
		}
		if card.Status != common.CardStatus_Active || card.CardCurrency != req.Currency {
			return pingerrors.ErrInvalid
		}
		wallet, err := uc.walletRepo.Lock(ctx, &WalletLockRequest{
			AccountID: req.AccountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock pingpong authorization wallet", "error", err)
			return pingerrors.ErrDatabase
		}
		if wallet.Type != common.WalletType_Card || wallet.Currency != req.Currency {
			return pingerrors.ErrInvalid
		}
		if wallet.Amount.Sub(wallet.PendingOut).LessThan(req.Amount) {
			return pingerrors.ErrInsufficient
		}
		wallet.PendingOut = wallet.PendingOut.Add(req.Amount)
		authorization = &model.Authorization{
			AccountID:         req.AccountID,
			Channel:           common.Channel_PingPong,
			CardID:            card.ID,
			Currency:          req.Currency,
			Amount:            req.Amount,
			MerchantName:      req.MerchantName,
			AuthorizationCode: randomx.Digits(6),
			Status:            common.TransactionStatus_AUTHORIZED,
			Account:           card.Account,
		}
		if err := uc.authorizationRepo.Create(ctx, authorization); err != nil {
			zap.S().Errorw("create pingpong local authorization", "error", err)
			return pingerrors.ErrDatabase
		}
		stage := &model.CardTransaction{
			AccountID:       req.AccountID,
			Channel:         common.Channel_PingPong,
			CardID:          card.ID,
			AuthorizationID: authorization.ID,
			Type:            common.CardTransactionType_AUTH,
			Status:          common.TransactionStatus_AUTHORIZED,
			Currency:        req.Currency,
			TxCurrency:      req.Currency,
			TxAmount:        req.Amount,
			RequestID:       req.RequestID,
			MerchantName:    req.MerchantName,
		}
		if err := uc.cardTransactionRepo.Create(ctx, stage); err != nil {
			zap.S().Errorw("create pingpong authorization stage", "error", err)
			return pingerrors.ErrDatabase
		}
		if err := uc.walletRepo.Save(ctx, wallet); err != nil {
			zap.S().Errorw("reserve pingpong authorization balance", "error", err)
			return pingerrors.ErrDatabase
		}
		return nil
	})
	return authorization, err
}

type stageLookup struct {
	AccountID model.ID
	RequestID string
}

type authorizationReference struct {
	AccountID model.ID
	ID        model.ID
}

func (uc *PingPongUIUsecase) previousStage(ctx context.Context, req *stageLookup) (*model.CardTransaction, error) {
	exists, err := uc.cardTransactionRepo.ExistsRequest(ctx, &CardTransactionRequestExistsRequest{
		AccountID: req.AccountID,
		RequestID: req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong stage request", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, nil
	}
	stage, err := uc.cardTransactionRepo.FindRequest(ctx, &CardTransactionRequestFindRequest{
		AccountID: req.AccountID,
		RequestID: req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong stage request", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return stage, nil
}

func (uc *PingPongUIUsecase) authorization(ctx context.Context, req *authorizationReference) (*model.Authorization, error) {
	if req.ID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.authorizationRepo.Exists(ctx, &AuthorizationExistsRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong authorization", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	item, err := uc.authorizationRepo.Find(ctx, &AuthorizationFindRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong authorization", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return item, nil
}

func (uc *PingPongUIUsecase) Stage(ctx context.Context, req *AuthorizationStageRequest) error {
	if !req.Amount.IsPositive() || req.RequestID == "" {
		return pingerrors.ErrInvalid
	}
	if req.Kind != common.CardTransactionType_CLEAR && req.Kind != common.CardTransactionType_VOID && req.Kind != common.CardTransactionType_REFUND {
		return pingerrors.ErrInvalid
	}
	return uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockAccount(ctx, req.AccountID); err != nil {
			return err
		}
		auth, err := uc.authorization(ctx, &authorizationReference{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			return err
		}
		previous, err := uc.previousStage(ctx, &stageLookup{
			AccountID: req.AccountID,
			RequestID: req.RequestID,
		})
		if err != nil {
			return err
		}
		if previous != nil {
			if previous.AuthorizationID != auth.ID || previous.Type != req.Kind || !previous.TxAmount.Equal(req.Amount) {
				return pingerrors.ErrConflict
			}
			return nil
		}
		card, err := uc.getCard(ctx, &uiCardReference{
			AccountID: req.AccountID,
			ID:        auth.CardID,
		})
		if err != nil {
			return err
		}
		wallet, err := uc.walletRepo.Lock(ctx, &WalletLockRequest{
			AccountID: req.AccountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock pingpong stage wallet", "error", err)
			return pingerrors.ErrDatabase
		}
		remaining := CalculateRemainingAuthorization(auth)
		if req.Kind != common.CardTransactionType_REFUND {
			release := decimal.Min(req.Amount, decimal.Max(remaining, decimal.Zero))
			wallet.PendingOut = wallet.PendingOut.Sub(release)
		}
		if req.Kind == common.CardTransactionType_CLEAR {
			wallet.Amount = wallet.Amount.Sub(req.Amount)
			wallet.Out = wallet.Out.Add(req.Amount)
		}
		if req.Kind == common.CardTransactionType_REFUND {
			wallet.Amount = wallet.Amount.Add(req.Amount)
			wallet.In = wallet.In.Add(req.Amount)
		}
		stage := &model.CardTransaction{
			AccountID:       req.AccountID,
			Channel:         common.Channel_PingPong,
			CardID:          card.ID,
			AuthorizationID: auth.ID,
			Type:            req.Kind,
			Status:          common.TransactionStatus_SUCCEED,
			Currency:        auth.Currency,
			TxCurrency:      auth.Currency,
			TxAmount:        req.Amount,
			RequestID:       req.RequestID,
		}
		if err := uc.cardTransactionRepo.Create(ctx, stage); err != nil {
			zap.S().Errorw("create pingpong card stage", "error", err)
			return pingerrors.ErrDatabase
		}
		if err := uc.walletRepo.Save(ctx, wallet); err != nil {
			zap.S().Errorw("save pingpong stage wallet", "error", err)
			return pingerrors.ErrDatabase
		}
		return nil
	})
}

func CalculateRemainingAuthorization(auth *model.Authorization) decimal.Decimal {
	remaining := auth.Amount
	for _, stage := range auth.CardTransactions {
		if stage.Status == common.TransactionStatus_SUCCEED && (stage.Type == common.CardTransactionType_CLEAR || stage.Type == common.CardTransactionType_VOID) {
			remaining = remaining.Sub(stage.TxAmount)
		}
	}
	return remaining
}

func (uc *PingPongUIUsecase) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) ([]*model.Authorization, int64, error) {
	items, err := uc.authorizationRepo.List(ctx, &AuthorizationListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		CardIDs:    types.PointerSlice(req.CardID),
		Offset:     req.Offset,
		Limit:      &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong authorizations", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.authorizationRepo.Count(ctx, &AuthorizationCountRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		CardIDs:    types.PointerSlice(req.CardID),
	})
	if err != nil {
		zap.S().Errorw("count pingpong authorizations", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}
