package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type ListAuthorizationBalancesRequest struct {
	AccountID    *model.ID
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	ID           *model.ID
	CardID       *model.ID
	MerchantName *string
}

type ListAuthorizationsRequest struct {
	AccountID *model.ID
	Offset    int
	Limit     int
	ID        *model.ID
	CardID    *model.ID
	Status    *enums.CardTransactionStatus
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

type ClearAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
}

type AuthorizationBalance struct {
	Authorization *model.Authorization
	Settled       decimal.Decimal
	Remaining     decimal.Decimal
}

type authorizationBalanceRequest struct {
	Authorization *model.Authorization
	Stages        []*model.CardTransaction
}

func authorizationBalance(req *authorizationBalanceRequest) *AuthorizationBalance {
	auth := req.Authorization
	stages := req.Stages
	result := &AuthorizationBalance{
		Authorization: auth,
		Remaining:     auth.Amount,
	}
	for _, stage := range stages {
		if stage.Type == enums.CardTransactionType_CLEAR && stage.Status != enums.TransactionStatus_FAILED {
			result.Settled = result.Settled.Add(stage.TxAmount)
			result.Remaining = result.Remaining.Sub(stage.TxAmount)
		}
		if stage.Type == enums.CardTransactionType_VOID {
			result.Remaining = result.Remaining.Sub(stage.TxAmount)
		}
	}
	return result
}

func (u *SlashUIUsecase) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*SimulateAuthorizationResult, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var result *SimulateAuthorizationResult
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.GetCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}
		walletID := card.WalletID
		if card.VirtualAccountID != nil {
			virtualAccount, err := u.virtualAccountRepository.FindByID(txCtx, *card.VirtualAccountID)
			if err != nil {
				zap.S().Errorw("find slash UI authorization virtual account", "error", err)
				return ErrDatabaseOperation
			}
			walletID = virtualAccount.WalletID
		}
		if walletID == 0 {
			return ErrResourceNotFound
		}
		wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, walletID)
		if err != nil {
			zap.S().Errorw("lock slash UI authorization wallet", "error", err)
			return ErrDatabaseOperation
		}
		if wallet.Currency != req.Currency {
			return ErrInvalidOperation
		}
		authorization := &model.Authorization{
			Account:           card.Account,
			AccountID:         card.AccountID,
			Channel:           enums.Channel_Slash,
			CardID:            card.ID,
			Currency:          req.Currency,
			Amount:            req.Amount,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: randomx.Digits(6),
			Status:            enums.TransactionStatus_AUTHORIZED,
		}
		if err := u.authorizationRepository.Create(txCtx, authorization); err != nil {
			zap.S().Errorw("create slash authorization", "error", err)
			return ErrDatabaseOperation
		}
		transaction := &model.CardTransaction{
			Account:           card.Account,
			AccountID:         card.AccountID,
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

func (u *SlashUIUsecase) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) ([]*model.Authorization, int64, error) {
	items, err := u.authorizationRepository.List(ctx, &AuthorizationListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		CardIDs:    types.PointerSlice(req.CardID),
		IDs:        types.PointerSlice(req.ID),
		Limit:      req.Limit,
		Offset:     req.Offset,
		Statuses:   types.PointerSlice(req.Status),
	})
	if err != nil {
		zap.S().Errorw("list slash authorizations", "error", err)
		return nil, 0, ErrDatabaseOperation
	}
	total, err := u.authorizationRepository.Count(ctx, &AuthorizationCountRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		CardIDs:    types.PointerSlice(req.CardID),
		IDs:        types.PointerSlice(req.ID),
		Limit:      req.Limit,
		Offset:     req.Offset,
		Statuses:   types.PointerSlice(req.Status),
	})
	if err != nil {
		zap.S().Errorw("count slash authorizations", "error", err)
		return nil, 0, ErrDatabaseOperation
	}

	return items, total, nil
}

func (u *SlashUIUsecase) GetAuthorization(ctx context.Context, id model.ID) (*model.Authorization, error) {
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

func (u *SlashUIUsecase) requireAuthorization(ctx context.Context, id model.ID) error {
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

func (u *SlashUIUsecase) ListAuthorizationBalances(ctx context.Context, req *ListAuthorizationBalancesRequest) ([]*AuthorizationBalance, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, ErrInvalidOperation
	}
	items, err := u.authorizationRepository.ListAuthorizations(ctx, &AuthorizationListBalancesRequest{
		AccountIDs:   types.PointerSlice(req.AccountID),
		IDs:          types.PointerSlice(req.ID),
		CardIDs:      types.PointerSlice(req.CardID),
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
	})
	if err != nil {
		zap.S().Errorw("list slash authorization balances", "error", err)
		return nil, ErrDatabaseOperation
	}
	results := make([]*AuthorizationBalance, 0, len(items))
	for _, item := range items {
		stages, err := u.cardTransactionRepository.ListStages(ctx, &ListAuthorizationStagesRequest{
			AccountID: item.AccountID,
			ID:        item.ID,
		})
		if err != nil {
			zap.S().Errorw("list slash authorization stages", "error", err)
			return nil, ErrDatabaseOperation
		}
		results = append(results, authorizationBalance(&authorizationBalanceRequest{
			Authorization: item,
			Stages:        stages,
		}))
	}
	return results, nil
}

func (u *SlashUIUsecase) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*model.CardTransaction, error) {
	if req.ID <= 0 || req.AccountID <= 0 || !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var result *model.CardTransaction
	err := u.transaction.InTx(ctx, func(ctx context.Context) error {
		exists, err := u.authorizationRepository.AuthorizationExists(ctx, &ExistAuthorizationRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			zap.S().Errorw("check slash clearing authorization", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		auth, err := u.authorizationRepository.LockAuthorization(ctx, &LockAuthorizationRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			zap.S().Errorw("lock slash clearing authorization", "error", err)
			return ErrDatabaseOperation
		}
		stages, err := u.cardTransactionRepository.ListStages(ctx, &ListAuthorizationStagesRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			zap.S().Errorw("list slash clearing stages", "error", err)
			return ErrDatabaseOperation
		}
		card, err := u.cardRepository.FindCard(ctx, &FindCardRequest{
			AccountID: req.AccountID,
			ID:        auth.CardID,
		})
		if err != nil {
			zap.S().Errorw("find slash clearing card", "error", err)
			return ErrDatabaseOperation
		}
		wallet, err := u.walletRepository.LockWallet(ctx, &LockWalletRequest{
			AccountID: req.AccountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock slash clearing wallet", "error", err)
			return ErrDatabaseOperation
		}
		if wallet.Currency != auth.Currency {
			return ErrInvalidOperation
		}
		wallet.Amount = wallet.Amount.Sub(req.Amount)
		wallet.Out = wallet.Out.Add(req.Amount)
		if err := u.walletRepository.SaveWallet(ctx, wallet); err != nil {
			zap.S().Errorw("save slash clearing wallet", "error", err)
			return ErrDatabaseOperation
		}
		var originID model.ID
		for _, stage := range stages {
			if stage.Type == enums.CardTransactionType_AUTH {
				originID = stage.ID
				break
			}
		}
		result = &model.CardTransaction{
			AccountID:               req.AccountID,
			Channel:                 enums.Channel_Slash,
			AuthorizationID:         auth.ID,
			OriginCardTransactionID: originID,
			CardID:                  auth.CardID,
			Type:                    enums.CardTransactionType_CLEAR,
			Status:                  enums.TransactionStatus_SUCCEED,
			Currency:                auth.Currency,
			TxCurrency:              auth.Currency,
			TxAmount:                req.Amount,
			MerchantName:            auth.MerchantName,
			MerchantCountry:         auth.MerchantCountry,
			MerchantMCC:             auth.MerchantMCC,
			AuthorizationCode:       auth.AuthorizationCode,
		}
		if err := u.cardTransactionRepository.Create(ctx, result); err != nil {
			zap.S().Errorw("create slash clearing stage", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
