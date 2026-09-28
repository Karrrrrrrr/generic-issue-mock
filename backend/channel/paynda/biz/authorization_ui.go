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
	Statuses     []enums.CardTransactionStatus
	AccountID    *model.ID
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	ID           *model.ID
	CardID       *model.ID
	MerchantName *string
}

type PayndaSimulateAuthorizationRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

type PayndaSimulateAuthorizationResult struct {
	Authorization   *model.Authorization
	CardTransaction *model.CardTransaction
}

type ClearAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
}

type ReverseAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
}

type RefundAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
}

type applyAuthorizationStepRequest struct {
	AccountID model.ID
	ID        model.ID
	Amount    decimal.Decimal
	Type      enums.CardTransactionType
}

type AuthorizationBalance struct {
	Authorization *model.Authorization
	Settled       decimal.Decimal
	Reversed      decimal.Decimal
	Refunded      decimal.Decimal
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
		switch stage.Type {
		case enums.CardTransactionType_CLEAR:
			if stage.Status == enums.TransactionStatus_SUCCEED {
				result.Settled = result.Settled.Add(stage.TxAmount)
			}
		case enums.CardTransactionType_VOID:
			if stage.Status == enums.TransactionStatus_VOID || stage.Status == enums.TransactionStatus_SUCCEED {
				result.Reversed = result.Reversed.Add(stage.TxAmount)
			}
		case enums.CardTransactionType_REFUND:
			if stage.Status == enums.TransactionStatus_SUCCEED {
				result.Refunded = result.Refunded.Add(stage.TxAmount)
			}
		}
	}
	result.Remaining = auth.Amount.Sub(result.Settled).Sub(result.Reversed)
	return result
}

func (u *PayndaUIUsecase) ListAuthorizations(ctx context.Context, req *PayndaListRequest) ([]*model.Authorization, error) {
	items, err := u.authorizationRepository.List(ctx, &AuthorizationListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		zap.S().Errorw("list paynda UI authorizations", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

func (u *PayndaUIUsecase) SimulateAuthorization(
	ctx context.Context,
	req *PayndaSimulateAuthorizationRequest,
) (*PayndaSimulateAuthorizationResult, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var result *PayndaSimulateAuthorizationResult
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistByID(txCtx, &CardExistByIDRequest{ID: req.CardID})
		if err != nil {
			zap.S().Errorw("check paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		card, err := u.cardRepository.FindByID(txCtx, &CardFindByIDRequest{ID: req.CardID})
		if err != nil {
			zap.S().Errorw("find paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}
		if card.WalletID == 0 {
			return ErrResourceNotFound
		}
		wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, &WalletFindByIDForUpdateRequest{
			AccountID: &card.AccountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock paynda UI authorization wallet", "error", err)
			return ErrDatabaseOperation
		}

		if wallet.Currency != req.Currency {
			return ErrInvalidOperation
		}
		authorization := &model.Authorization{
			Account:           card.Account,
			AccountID:         card.AccountID,
			Channel:           enums.Channel_Paynda,
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
			zap.S().Errorw("create paynda UI authorization", "error", err)
			return ErrDatabaseOperation
		}
		transaction := &model.CardTransaction{
			Account:           card.Account,
			AccountID:         card.AccountID,
			Channel:           enums.Channel_Paynda,
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
			zap.S().Errorw("create paynda UI transaction", "error", err)
			return ErrDatabaseOperation
		}
		result = &PayndaSimulateAuthorizationResult{
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

func (u *PayndaUIUsecase) ListAuthorizationBalances(ctx context.Context, req *ListAuthorizationBalancesRequest) ([]*AuthorizationBalance, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, ErrInvalidOperation
	}
	items, err := u.authorizationRepository.ListAuthorizations(ctx, &AuthorizationListBalancesRequest{
		AccountIDs:   types.PointerSlice(req.AccountID),
		Statuses:     req.Statuses,
		IDs:          types.PointerSlice(req.ID),
		CardIDs:      types.PointerSlice(req.CardID),
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
	})
	if err != nil {
		zap.S().Errorw("list paynda authorization balances", "error", err)
		return nil, ErrDatabaseOperation
	}
	results := make([]*AuthorizationBalance, 0, len(items))
	for _, item := range items {
		stages, err := u.cardTransactionRepository.ListStages(ctx, &ListAuthorizationStagesRequest{
			AccountID: item.AccountID,
			ID:        item.ID,
		})
		if err != nil {
			zap.S().Errorw("list paynda authorization stages", "error", err)
			return nil, ErrDatabaseOperation
		}
		results = append(results, authorizationBalance(&authorizationBalanceRequest{
			Authorization: item,
			Stages:        stages,
		}))
	}
	return results, nil
}

func (u *PayndaUIUsecase) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
		Amount:    req.Amount,
		Type:      enums.CardTransactionType_CLEAR,
	})
}

func (u *PayndaUIUsecase) ReverseAuthorization(ctx context.Context, req *ReverseAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
		Amount:    req.Amount,
		Type:      enums.CardTransactionType_VOID,
	})
}

func (u *PayndaUIUsecase) RefundAuthorization(ctx context.Context, req *RefundAuthorizationRequest) (*model.CardTransaction, error) {
	return u.applyAuthorizationStep(ctx, &applyAuthorizationStepRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
		Amount:    req.Amount,
		Type:      enums.CardTransactionType_REFUND,
	})
}

func (u *PayndaUIUsecase) applyAuthorizationStep(ctx context.Context, req *applyAuthorizationStepRequest) (*model.CardTransaction, error) {
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
			zap.S().Errorw("check paynda authorization for operation", "error", err)
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
			zap.S().Errorw("lock paynda authorization for operation", "error", err)
			return ErrDatabaseOperation
		}
		stages, err := u.cardTransactionRepository.ListStages(ctx, &ListAuthorizationStagesRequest{
			AccountID: req.AccountID,
			ID:        req.ID,
		})
		if err != nil {
			zap.S().Errorw("list paynda authorization operation stages", "error", err)
			return ErrDatabaseOperation
		}
		if req.Type != enums.CardTransactionType_VOID {
			card, err := u.cardRepository.FindCard(ctx, &FindCardRequest{
				AccountID: req.AccountID,
				ID:        auth.CardID,
			})
			if err != nil {
				zap.S().Errorw("find paynda authorization operation card", "error", err)
				return ErrDatabaseOperation
			}
			wallet, err := u.walletRepository.LockWallet(ctx, &LockWalletRequest{
				AccountID: req.AccountID,
				ID:        card.WalletID,
			})
			if err != nil {
				zap.S().Errorw("lock paynda authorization operation wallet", "error", err)
				return ErrDatabaseOperation
			}
			if wallet.Currency != auth.Currency {
				return ErrInvalidOperation
			}
			if req.Type == enums.CardTransactionType_REFUND {
				wallet.Available = wallet.Available.Add(req.Amount)
				wallet.In = wallet.In.Add(req.Amount)
			} else {
				wallet.Available = wallet.Available.Sub(req.Amount)
				wallet.Out = wallet.Out.Add(req.Amount)
			}
			if err := u.walletRepository.SaveWallet(ctx, wallet); err != nil {
				zap.S().Errorw("save paynda authorization operation wallet", "error", err)
				return ErrDatabaseOperation
			}
		}
		var originID model.ID
		for _, stage := range stages {
			if stage.Type == enums.CardTransactionType_AUTH {
				originID = stage.ID
				break
			}
		}
		status := enums.TransactionStatus_SUCCEED
		if req.Type == enums.CardTransactionType_VOID {
			status = enums.TransactionStatus_VOID
		}
		result = &model.CardTransaction{
			AccountID:               req.AccountID,
			Channel:                 enums.Channel_Paynda,
			AuthorizationID:         auth.ID,
			OriginCardTransactionID: originID,
			CardID:                  auth.CardID,
			Type:                    req.Type,
			Status:                  status,
			Currency:                auth.Currency,
			TxCurrency:              auth.Currency,
			TxAmount:                req.Amount,
			MerchantName:            auth.MerchantName,
			MerchantCountry:         auth.MerchantCountry,
			MerchantMCC:             auth.MerchantMCC,
			AuthorizationCode:       auth.AuthorizationCode,
		}
		if err := u.cardTransactionRepository.Create(ctx, result); err != nil {
			zap.S().Errorw("create paynda authorization operation stage", "error", err)
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
