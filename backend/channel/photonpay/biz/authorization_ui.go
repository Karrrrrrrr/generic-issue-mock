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

type UISimulateAuthorizationRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

type UISimulateAuthorizationResult struct {
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

func (u *PhotonPayUIUsecase) ListAuthorizations(ctx context.Context, req *ListRequest) ([]*model.Authorization, error) {
	items, err := u.authorizationRepo.List(ctx, &AuthorizationListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list photonpay UI authorizations", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

func (u *PhotonPayUIUsecase) SimulateAuthorization(ctx context.Context, req *UISimulateAuthorizationRequest) (*UISimulateAuthorizationResult, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var result *UISimulateAuthorizationResult
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.getCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}
		if card.WalletID == 0 {
			return ErrResourceNotFound
		}
		accountID := card.AccountID
		wallet, err := u.walletRepo.FindByIDForUpdate(txCtx, &WalletFindByIDForUpdateRequest{
			AccountID: &accountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay UI authorization wallet", "error", err)

			return ErrDatabaseOperation
		}

		if wallet.Currency != req.Currency {
			return ErrInvalidOperation
		}
		authorization := &model.Authorization{
			Account:           card.Account,
			AccountID:         card.AccountID,
			Channel:           enums.Channel_PhotonPay,
			CardID:            card.ID,
			Currency:          req.Currency,
			Amount:            req.Amount,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: randomx.Digits(6),
			Status:            enums.TransactionStatus_AUTHORIZED,
		}
		if err := u.authorizationRepo.Create(txCtx, authorization); err != nil {
			zap.S().Errorw("create photonpay UI authorization", "error", err)

			return ErrDatabaseOperation
		}

		transaction := &model.CardTransaction{
			Account:           card.Account,
			AccountID:         card.AccountID,
			Channel:           enums.Channel_PhotonPay,
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
		if err := u.cardTransactionRepo.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create photonpay UI authorization transaction", "error", err)

			return ErrDatabaseOperation
		}

		result = &UISimulateAuthorizationResult{
			Authorization:   authorization,
			CardTransaction: transaction,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	u.dispatchTransaction(ctx, result.CardTransaction)

	return result, nil
}

func (u *PhotonPayUIUsecase) ListAuthorizationBalances(ctx context.Context, req *ListAuthorizationBalancesRequest) ([]*AuthorizationBalance, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, ErrInvalidOperation
	}
	items, err := u.authorizationRepo.ListAuthorizations(ctx, &AuthorizationListBalancesRequest{
		AccountIDs:   types.PointerSlice(req.AccountID),
		IDs:          types.PointerSlice(req.ID),
		CardIDs:      types.PointerSlice(req.CardID),
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
	})
	if err != nil {
		zap.S().Errorw("list photonpay authorization balances", "error", err)
		return nil, ErrDatabaseOperation
	}
	results := make([]*AuthorizationBalance, 0, len(items))
	for _, item := range items {
		stages, err := u.cardTransactionRepo.ListStages(ctx, &ListAuthorizationStagesRequest{
			AccountID: item.AccountID,
			ID:        item.ID,
		})
		if err != nil {
			zap.S().Errorw("list photonpay authorization stages", "error", err)
			return nil, ErrDatabaseOperation
		}
		results = append(results, authorizationBalance(&authorizationBalanceRequest{
			Authorization: item,
			Stages:        stages,
		}))
	}
	return results, nil
}

func (u *PhotonPayUIUsecase) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*model.CardTransaction, error) {
	if req.ID <= 0 || req.AccountID <= 0 || !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var result *model.CardTransaction
	err := u.transaction.InTx(ctx, func(ctx context.Context) error {
		exists, err := u.authorizationRepo.AuthorizationExists(ctx, &ExistAuthorizationRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			zap.S().Errorw("check photonpay clearing authorization", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		auth, err := u.authorizationRepo.LockAuthorization(ctx, &LockAuthorizationRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay clearing authorization", "error", err)
			return ErrDatabaseOperation
		}
		stages, err := u.cardTransactionRepo.ListStages(ctx, &ListAuthorizationStagesRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			zap.S().Errorw("list photonpay clearing stages", "error", err)
			return ErrDatabaseOperation
		}
		card, err := u.cardRepo.FindCard(ctx, &FindCardRequest{
			AccountID: req.AccountID,
			ID:        auth.CardID,
		})
		if err != nil {
			zap.S().Errorw("find photonpay clearing card", "error", err)
			return ErrDatabaseOperation
		}
		wallet, err := u.walletRepo.LockWallet(ctx, &LockWalletRequest{
			AccountID: req.AccountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay clearing wallet", "error", err)
			return ErrDatabaseOperation
		}
		if wallet.Currency != auth.Currency {
			return ErrInvalidOperation
		}
		wallet.Amount = wallet.Amount.Sub(req.Amount)
		wallet.Out = wallet.Out.Add(req.Amount)
		if err := u.walletRepo.SaveWallet(ctx, wallet); err != nil {
			zap.S().Errorw("save photonpay clearing wallet", "error", err)
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
			Channel:                 enums.Channel_PhotonPay,
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
		if err := u.cardTransactionRepo.Create(ctx, result); err != nil {
			zap.S().Errorw("create photonpay clearing stage", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	u.dispatchTransaction(ctx, result)
	return result, nil
}
