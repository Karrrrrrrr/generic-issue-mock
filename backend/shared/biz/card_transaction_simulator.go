package biz

import (
	"context"
	"errors"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type SimulateAuthorizationReq struct {
	Channel         enums.Channel
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
	RequestID       *string
	Notificator     CardTransactionNotificator
}

func (req *SimulateAuthorizationReq) Validate() error {
	if req == nil || req.Channel == "" || req.CardID <= 0 ||
		!req.Amount.IsPositive() || req.Currency == "" || req.Notificator == nil {
		return sharederrors.ErrInvalidSimulationRequest
	}
	for _, value := range []*string{req.RequestID, req.MerchantName, req.MerchantCountry, req.MerchantMCC} {
		if value != nil && *value == "" {
			return sharederrors.ErrInvalidSimulationRequest
		}
	}
	return nil
}

type SimulateClearingReq struct {
	Channel         enums.Channel
	AuthorizationID model.ID
	Amount          decimal.Decimal
	RequestID       *string
	Notificator     CardTransactionNotificator
}

func (req *SimulateClearingReq) Validate() error {
	if req == nil || req.Channel == "" ||
		req.AuthorizationID <= 0 || !req.Amount.IsPositive() || req.Notificator == nil ||
		(req.RequestID != nil && *req.RequestID == "") {
		return sharederrors.ErrInvalidSimulationRequest
	}
	return nil
}

type SimulateRefundReq struct {
	Channel         enums.Channel
	CardID          *model.ID
	AuthorizationID *model.ID
	Amount          decimal.Decimal
	Currency        *enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
	RequestID       *string
	Notificator     CardTransactionNotificator
}

func (req *SimulateRefundReq) Validate() error {
	if req == nil || req.Channel == "" ||
		!req.Amount.IsPositive() || (req.CardID != nil && *req.CardID <= 0) ||
		(req.Currency != nil && *req.Currency == "") ||
		(req.AuthorizationID == nil && (req.CardID == nil || req.Currency == nil)) ||
		req.Notificator == nil ||
		(req.AuthorizationID != nil && *req.AuthorizationID <= 0) {
		return sharederrors.ErrInvalidSimulationRequest
	}
	for _, value := range []*string{req.RequestID, req.MerchantName, req.MerchantCountry, req.MerchantMCC} {
		if value != nil && *value == "" {
			return sharederrors.ErrInvalidSimulationRequest
		}
	}
	return nil
}

type SimulateReversalReq struct {
	Channel         enums.Channel
	AuthorizationID model.ID
	Amount          decimal.Decimal
	Status          enums.CardTransactionStatus
	RequestID       *string
	Notificator     CardTransactionNotificator
}

func (req *SimulateReversalReq) Validate() error {
	if req == nil || req.Channel == "" ||
		req.AuthorizationID <= 0 || !req.Amount.IsPositive() || req.Notificator == nil ||
		(req.Status != enums.TransactionStatus_VOID && req.Status != enums.TransactionStatus_SUCCEED) ||
		(req.RequestID != nil && *req.RequestID == "") {
		return sharederrors.ErrInvalidSimulationRequest
	}
	return nil
}

type CardTransactionSimulationResult struct {
	Authorization     *model.Authorization
	CardTransaction   *model.CardTransaction
	Remaining         decimal.Decimal
	Replayed          bool
	NotificationError error
}

type simulationCardReference struct {
	Channel enums.Channel
	CardID  model.ID
}

type simulationAuthorizationReference struct {
	Channel         enums.Channel
	AuthorizationID model.ID
}

type simulationAccountRequest struct {
	AccountID model.ID
	Channel   enums.Channel
}

type simulationCardRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	CardID    model.ID
}

type simulationWalletRequest struct {
	Card     *model.Card
	Currency enums.Currency
}

func (req *simulationWalletRequest) Validate() error {
	if req == nil || req.Card == nil {
		return sharederrors.ErrInvalidWallet
	}
	card := req.Card
	wallet := card.Wallet
	if wallet == nil || card.WalletID <= 0 || card.CardCurrency != req.Currency ||
		wallet.ID != card.WalletID || wallet.AccountID != card.AccountID || wallet.Channel != card.Channel ||
		wallet.Currency != req.Currency {
		return sharederrors.ErrInvalidWallet
	}
	switch card.CardType {
	case enums.CardType_Single:
		if card.VirtualAccountID != nil || wallet.Type != enums.WalletType_Card {
			return sharederrors.ErrInvalidWallet
		}
	case enums.CardType_Share:
		if card.VirtualAccountID == nil || *card.VirtualAccountID <= 0 || wallet.Type != enums.WalletType_VirtualAccount {
			return sharederrors.ErrInvalidWallet
		}
	case enums.CardType_VirtualAccountSingle:
		if card.VirtualAccountID == nil || *card.VirtualAccountID <= 0 || wallet.Type != enums.WalletType_Card {
			return sharederrors.ErrInvalidWallet
		}
	default:
		return sharederrors.ErrInvalidWallet
	}
	return nil
}

type simulationAuthorizationRequest struct {
	AccountID       model.ID
	Channel         enums.Channel
	CardID          model.ID
	AuthorizationID model.ID
}

type simulationRequestReference struct {
	AccountID model.ID
	Channel   enums.Channel
	RequestID *string
}

type simulationAuthorizationState struct {
	Authorization       *model.Authorization
	Remaining           decimal.Decimal
	OriginTransactionID model.ID
}

type notifySimulationRequest struct {
	Result      *CardTransactionSimulationResult
	Notificator CardTransactionNotificator
}

type CardTransactionSimulator interface {
	SimulateAuthorization(context.Context, *SimulateAuthorizationReq) (*CardTransactionSimulationResult, error)
	SimulateClearing(context.Context, *SimulateClearingReq) (*CardTransactionSimulationResult, error)
	SimulateRefund(context.Context, *SimulateRefundReq) (*CardTransactionSimulationResult, error)
	SimulateReversal(context.Context, *SimulateReversalReq) (*CardTransactionSimulationResult, error)
}

type cardTransactionSimulator struct {
	accountRepo         AccountRepo
	cardRepo            CardRepo
	balanceChanger      BalanceChanger
	authorizationRepo   AuthorizationRepo
	cardTransactionRepo CardTransactionRepo
	tx                  Transaction
}

var _ CardTransactionSimulator = (*cardTransactionSimulator)(nil)

func NewCardTransactionSimulator(injector *do.Injector) (CardTransactionSimulator, error) {
	return &cardTransactionSimulator{
		accountRepo:         do.MustInvoke[AccountRepo](injector),
		cardRepo:            do.MustInvoke[CardRepo](injector),
		balanceChanger:      do.MustInvoke[BalanceChanger](injector),
		authorizationRepo:   do.MustInvoke[AuthorizationRepo](injector),
		cardTransactionRepo: do.MustInvoke[CardTransactionRepo](injector),
		tx:                  do.MustInvoke[Transaction](injector),
	}, nil
}

func (simulator *cardTransactionSimulator) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationReq) (*CardTransactionSimulationResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var accountID model.ID
	var result *CardTransactionSimulationResult
	var operationErr error
	err := simulator.tx.InTx(ctx, func(ctx context.Context) (err error) {
		defer func() {
			operationErr = err
		}()
		reference, err := simulator.findSimulationCard(ctx, &simulationCardReference{
			Channel: req.Channel,
			CardID:  req.CardID,
		})
		if err != nil {
			return err
		}
		accountID = reference.AccountID
		if err := simulator.lockAccount(ctx, &simulationAccountRequest{
			AccountID: accountID,
			Channel:   req.Channel,
		}); err != nil {
			return err
		}
		previous, err := simulator.findPreviousTransaction(ctx, &simulationRequestReference{
			AccountID: accountID,
			Channel:   req.Channel,
			RequestID: req.RequestID,
		})
		if err != nil {
			return err
		}
		if previous != nil {
			if previous.Type != enums.CardTransactionType_AUTH || previous.Status != enums.TransactionStatus_AUTHORIZED ||
				previous.CardID != req.CardID || previous.Currency != req.Currency || !previous.TxAmount.Equal(req.Amount) ||
				previous.MerchantName != types.Value(req.MerchantName) ||
				previous.MerchantCountry != types.Value(req.MerchantCountry) || previous.MerchantMCC != types.Value(req.MerchantMCC) {
				return sharederrors.ErrSimulationRequestConflict
			}
			result, err = simulator.replayTransaction(ctx, previous)
			return err
		}
		card, err := simulator.lockCard(ctx, &simulationCardRequest{
			AccountID: accountID,
			Channel:   req.Channel,
			CardID:    req.CardID,
		})
		if err != nil {
			return err
		}
		if card.Status != enums.CardStatus_Active {
			return sharederrors.ErrCardNotActive
		}
		walletReq := &simulationWalletRequest{
			Card:     card,
			Currency: req.Currency,
		}
		if err := walletReq.Validate(); err != nil {
			return err
		}
		if err := simulator.balanceChanger.TryBalanceChange(ctx, &TccBalanceChangeReq{
			AccountID:      accountID,
			Channel:        req.Channel,
			WalletID:       card.WalletID,
			Currency:       req.Currency,
			Amount:         req.Amount,
			CheckAvailable: true,
		}); err != nil {
			if errors.Is(err, sharederrors.ErrInsufficientAvailableBalance) {
				return sharederrors.ErrInsufficientCardBalance
			}
			return err
		}
		authorization := &model.Authorization{
			Account:           card.Account,
			AccountID:         accountID,
			Channel:           req.Channel,
			CardID:            card.ID,
			Currency:          req.Currency,
			Amount:            req.Amount,
			MerchantName:      types.Value(req.MerchantName),
			MerchantCountry:   types.Value(req.MerchantCountry),
			MerchantMCC:       types.Value(req.MerchantMCC),
			AuthorizationCode: randomx.Digits(6),
			Status:            enums.TransactionStatus_AUTHORIZED,
		}
		if err := simulator.authorizationRepo.Create(ctx, authorization); err != nil {
			zap.S().Errorw("create shared simulated authorization",
				"account_id", accountID,
				"channel", req.Channel,
				"card_id", req.CardID,
				"error", err,
			)
			return sharederrors.ErrDatabaseOperation
		}
		transaction := &model.CardTransaction{
			Account:           card.Account,
			AccountID:         accountID,
			Channel:           req.Channel,
			CardID:            card.ID,
			AuthorizationID:   authorization.ID,
			Type:              enums.CardTransactionType_AUTH,
			Status:            enums.TransactionStatus_AUTHORIZED,
			Currency:          req.Currency,
			TxCurrency:        req.Currency,
			TxAmount:          req.Amount,
			RequestID:         types.Value(req.RequestID),
			MerchantName:      authorization.MerchantName,
			MerchantCountry:   authorization.MerchantCountry,
			MerchantMCC:       authorization.MerchantMCC,
			AuthorizationCode: authorization.AuthorizationCode,
		}
		if err := simulator.cardTransactionRepo.Create(ctx, transaction); err != nil {
			zap.S().Errorw("create shared simulated authorization transaction",
				"account_id", accountID,
				"channel", req.Channel,
				"authorization_id", authorization.ID,
				"error", err,
			)
			return sharederrors.ErrDatabaseOperation
		}
		result = &CardTransactionSimulationResult{
			Authorization:   authorization,
			CardTransaction: transaction,
			Remaining:       req.Amount,
		}
		return nil
	})
	if err != nil {
		if operationErr != nil {
			return nil, operationErr
		}
		if errors.Is(err, sharederrors.ErrNestedTransaction) {
			return nil, err
		}
		zap.S().Errorw("run shared authorization simulation transaction",
			"account_id", accountID,
			"channel", req.Channel,
			"card_id", req.CardID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	simulator.notifyTransaction(ctx, &notifySimulationRequest{
		Result:      result,
		Notificator: req.Notificator,
	})
	return result, nil
}

func (simulator *cardTransactionSimulator) SimulateClearing(ctx context.Context, req *SimulateClearingReq) (*CardTransactionSimulationResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var accountID model.ID
	var cardID model.ID
	var result *CardTransactionSimulationResult
	var operationErr error
	err := simulator.tx.InTx(ctx, func(ctx context.Context) (err error) {
		defer func() {
			operationErr = err
		}()
		reference, err := simulator.findSimulationAuthorization(ctx, &simulationAuthorizationReference{
			Channel:         req.Channel,
			AuthorizationID: req.AuthorizationID,
		})
		if err != nil {
			return err
		}
		accountID = reference.AccountID
		cardID = reference.CardID
		if err := simulator.lockAccount(ctx, &simulationAccountRequest{
			AccountID: accountID,
			Channel:   req.Channel,
		}); err != nil {
			return err
		}
		previous, err := simulator.findPreviousTransaction(ctx, &simulationRequestReference{
			AccountID: accountID,
			Channel:   req.Channel,
			RequestID: req.RequestID,
		})
		if err != nil {
			return err
		}
		if previous != nil {
			if previous.Type != enums.CardTransactionType_CLEAR || previous.Status != enums.TransactionStatus_SUCCEED ||
				previous.CardID != cardID || previous.AuthorizationID != req.AuthorizationID || !previous.TxAmount.Equal(req.Amount) {
				return sharederrors.ErrSimulationRequestConflict
			}
			result, err = simulator.replayTransaction(ctx, previous)
			return err
		}
		state, err := simulator.loadAuthorization(ctx, &simulationAuthorizationRequest{
			AccountID:       accountID,
			Channel:         req.Channel,
			CardID:          cardID,
			AuthorizationID: req.AuthorizationID,
		})
		if err != nil {
			return err
		}
		card, err := simulator.lockCard(ctx, &simulationCardRequest{
			AccountID: accountID,
			Channel:   req.Channel,
			CardID:    cardID,
		})
		if err != nil {
			return err
		}
		authorization := state.Authorization
		walletReq := &simulationWalletRequest{
			Card:     card,
			Currency: authorization.Currency,
		}
		if err := walletReq.Validate(); err != nil {
			return err
		}
		release := decimal.Min(req.Amount, decimal.Max(state.Remaining, decimal.Zero))
		if err := simulator.balanceChanger.ConfirmBalanceChange(ctx, &ConfirmBalanceChangeReq{
			AccountID:      accountID,
			Channel:        req.Channel,
			WalletID:       card.WalletID,
			Currency:       authorization.Currency,
			Amount:         req.Amount,
			ReleaseAmount:  release,
			CheckAvailable: false,
		}); err != nil {
			return err
		}
		transaction := &model.CardTransaction{
			Account:                 card.Account,
			AccountID:               accountID,
			Channel:                 req.Channel,
			CardID:                  card.ID,
			AuthorizationID:         authorization.ID,
			OriginCardTransactionID: state.OriginTransactionID,
			Type:                    enums.CardTransactionType_CLEAR,
			Status:                  enums.TransactionStatus_SUCCEED,
			Currency:                authorization.Currency,
			TxCurrency:              authorization.Currency,
			TxAmount:                req.Amount,
			RequestID:               types.Value(req.RequestID),
			MerchantName:            authorization.MerchantName,
			MerchantCountry:         authorization.MerchantCountry,
			MerchantMCC:             authorization.MerchantMCC,
			AuthorizationCode:       authorization.AuthorizationCode,
		}
		if err := simulator.cardTransactionRepo.Create(ctx, transaction); err != nil {
			zap.S().Errorw("create shared simulated clearing transaction",
				"account_id", accountID,
				"channel", req.Channel,
				"authorization_id", authorization.ID,
				"error", err,
			)
			return sharederrors.ErrDatabaseOperation
		}
		result = &CardTransactionSimulationResult{
			Authorization:   authorization,
			CardTransaction: transaction,
			Remaining:       state.Remaining.Sub(req.Amount),
		}
		return nil
	})
	if err != nil {
		if operationErr != nil {
			return nil, operationErr
		}
		if errors.Is(err, sharederrors.ErrNestedTransaction) {
			return nil, err
		}
		zap.S().Errorw("run shared clearing simulation transaction",
			"account_id", accountID,
			"channel", req.Channel,
			"authorization_id", req.AuthorizationID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	simulator.notifyTransaction(ctx, &notifySimulationRequest{
		Result:      result,
		Notificator: req.Notificator,
	})
	return result, nil
}

func (simulator *cardTransactionSimulator) SimulateRefund(ctx context.Context, req *SimulateRefundReq) (*CardTransactionSimulationResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var accountID model.ID
	var cardID model.ID
	var currency enums.Currency
	var result *CardTransactionSimulationResult
	var operationErr error
	err := simulator.tx.InTx(ctx, func(ctx context.Context) (err error) {
		defer func() {
			operationErr = err
		}()
		if req.AuthorizationID != nil {
			reference, err := simulator.findSimulationAuthorization(ctx, &simulationAuthorizationReference{
				Channel:         req.Channel,
				AuthorizationID: *req.AuthorizationID,
			})
			if err != nil {
				return err
			}
			if (req.CardID != nil && *req.CardID != reference.CardID) ||
				(req.Currency != nil && *req.Currency != reference.Currency) {
				return sharederrors.ErrInvalidAuthorization
			}
			accountID = reference.AccountID
			cardID = reference.CardID
			currency = reference.Currency
		} else {
			reference, err := simulator.findSimulationCard(ctx, &simulationCardReference{
				Channel: req.Channel,
				CardID:  *req.CardID,
			})
			if err != nil {
				return err
			}
			accountID = reference.AccountID
			cardID = reference.ID
			currency = *req.Currency
		}
		if err := simulator.lockAccount(ctx, &simulationAccountRequest{
			AccountID: accountID,
			Channel:   req.Channel,
		}); err != nil {
			return err
		}
		authorizationID := types.Value(req.AuthorizationID)
		previous, err := simulator.findPreviousTransaction(ctx, &simulationRequestReference{
			AccountID: accountID,
			Channel:   req.Channel,
			RequestID: req.RequestID,
		})
		if err != nil {
			return err
		}
		if previous != nil && (previous.Type != enums.CardTransactionType_REFUND ||
			previous.Status != enums.TransactionStatus_SUCCEED || previous.CardID != cardID ||
			previous.AuthorizationID != authorizationID || previous.Currency != currency || !previous.TxAmount.Equal(req.Amount)) {
			return sharederrors.ErrSimulationRequestConflict
		}
		merchantName := types.Value(req.MerchantName)
		merchantCountry := types.Value(req.MerchantCountry)
		merchantMCC := types.Value(req.MerchantMCC)
		var authorization *model.Authorization
		var originTransactionID model.ID
		remaining := decimal.Zero
		if req.AuthorizationID != nil {
			state, err := simulator.loadAuthorization(ctx, &simulationAuthorizationRequest{
				AccountID:       accountID,
				Channel:         req.Channel,
				CardID:          cardID,
				AuthorizationID: authorizationID,
			})
			if err != nil {
				return err
			}
			authorization = state.Authorization
			if authorization.Currency != currency {
				return sharederrors.ErrInvalidAuthorization
			}
			originTransactionID = state.OriginTransactionID
			remaining = state.Remaining
			if req.MerchantName == nil {
				merchantName = authorization.MerchantName
			}
			if req.MerchantCountry == nil {
				merchantCountry = authorization.MerchantCountry
			}
			if req.MerchantMCC == nil {
				merchantMCC = authorization.MerchantMCC
			}
		}
		if previous != nil {
			if previous.MerchantName != merchantName ||
				previous.MerchantCountry != merchantCountry || previous.MerchantMCC != merchantMCC {
				return sharederrors.ErrSimulationRequestConflict
			}
			result = &CardTransactionSimulationResult{
				Authorization:   authorization,
				CardTransaction: previous,
				Remaining:       remaining,
				Replayed:        true,
			}
			return nil
		}
		card, err := simulator.lockCard(ctx, &simulationCardRequest{
			AccountID: accountID,
			Channel:   req.Channel,
			CardID:    cardID,
		})
		if err != nil {
			return err
		}
		if req.AuthorizationID == nil && card.Status != enums.CardStatus_Active {
			return sharederrors.ErrCardNotActive
		}
		walletReq := &simulationWalletRequest{
			Card:     card,
			Currency: currency,
		}
		if err := walletReq.Validate(); err != nil {
			return err
		}
		if err := simulator.balanceChanger.ChangeBalanceSimple(ctx, &ChangeBalanceSimpleReq{
			AccountID:      accountID,
			Channel:        req.Channel,
			WalletID:       card.WalletID,
			Currency:       currency,
			Amount:         req.Amount,
			CheckAvailable: false,
		}); err != nil {
			return err
		}
		authorizationCode := randomx.Digits(6)
		if authorization != nil {
			authorizationCode = authorization.AuthorizationCode
		}
		transaction := &model.CardTransaction{
			Account:                 card.Account,
			AccountID:               accountID,
			Channel:                 req.Channel,
			CardID:                  card.ID,
			AuthorizationID:         authorizationID,
			OriginCardTransactionID: originTransactionID,
			Type:                    enums.CardTransactionType_REFUND,
			Status:                  enums.TransactionStatus_SUCCEED,
			Currency:                currency,
			TxCurrency:              currency,
			TxAmount:                req.Amount,
			RequestID:               types.Value(req.RequestID),
			MerchantName:            merchantName,
			MerchantCountry:         merchantCountry,
			MerchantMCC:             merchantMCC,
			AuthorizationCode:       authorizationCode,
		}
		if err := simulator.cardTransactionRepo.Create(ctx, transaction); err != nil {
			zap.S().Errorw("create shared simulated refund transaction",
				"account_id", accountID,
				"channel", req.Channel,
				"card_id", cardID,
				"authorization_id", authorizationID,
				"error", err,
			)
			return sharederrors.ErrDatabaseOperation
		}
		result = &CardTransactionSimulationResult{
			Authorization:   authorization,
			CardTransaction: transaction,
			Remaining:       remaining,
		}
		return nil
	})
	if err != nil {
		if operationErr != nil {
			return nil, operationErr
		}
		if errors.Is(err, sharederrors.ErrNestedTransaction) {
			return nil, err
		}
		zap.S().Errorw("run shared refund simulation transaction",
			"account_id", accountID,
			"channel", req.Channel,
			"card_id", cardID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	simulator.notifyTransaction(ctx, &notifySimulationRequest{
		Result:      result,
		Notificator: req.Notificator,
	})
	return result, nil
}

func (simulator *cardTransactionSimulator) SimulateReversal(ctx context.Context, req *SimulateReversalReq) (*CardTransactionSimulationResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var accountID model.ID
	var cardID model.ID
	var result *CardTransactionSimulationResult
	var operationErr error
	err := simulator.tx.InTx(ctx, func(ctx context.Context) (err error) {
		defer func() {
			operationErr = err
		}()
		reference, err := simulator.findSimulationAuthorization(ctx, &simulationAuthorizationReference{
			Channel:         req.Channel,
			AuthorizationID: req.AuthorizationID,
		})
		if err != nil {
			return err
		}
		accountID = reference.AccountID
		cardID = reference.CardID
		if err := simulator.lockAccount(ctx, &simulationAccountRequest{
			AccountID: accountID,
			Channel:   req.Channel,
		}); err != nil {
			return err
		}
		previous, err := simulator.findPreviousTransaction(ctx, &simulationRequestReference{
			AccountID: accountID,
			Channel:   req.Channel,
			RequestID: req.RequestID,
		})
		if err != nil {
			return err
		}
		if previous != nil {
			if previous.Type != enums.CardTransactionType_VOID || previous.Status != req.Status ||
				previous.CardID != cardID || previous.AuthorizationID != req.AuthorizationID || !previous.TxAmount.Equal(req.Amount) {
				return sharederrors.ErrSimulationRequestConflict
			}
			result, err = simulator.replayTransaction(ctx, previous)
			return err
		}
		state, err := simulator.loadAuthorization(ctx, &simulationAuthorizationRequest{
			AccountID:       accountID,
			Channel:         req.Channel,
			CardID:          cardID,
			AuthorizationID: req.AuthorizationID,
		})
		if err != nil {
			return err
		}
		card, err := simulator.lockCard(ctx, &simulationCardRequest{
			AccountID: accountID,
			Channel:   req.Channel,
			CardID:    cardID,
		})
		if err != nil {
			return err
		}
		authorization := state.Authorization
		walletReq := &simulationWalletRequest{
			Card:     card,
			Currency: authorization.Currency,
		}
		if err := walletReq.Validate(); err != nil {
			return err
		}
		release := decimal.Min(req.Amount, decimal.Max(state.Remaining, decimal.Zero))
		if release.IsPositive() {
			if err := simulator.balanceChanger.CancelBalanceChange(ctx, &TccBalanceChangeReq{
				AccountID:      accountID,
				Channel:        req.Channel,
				WalletID:       card.WalletID,
				Currency:       authorization.Currency,
				Amount:         release,
				CheckAvailable: false,
			}); err != nil {
				return err
			}
		}
		transaction := &model.CardTransaction{
			Account:                 card.Account,
			AccountID:               accountID,
			Channel:                 req.Channel,
			CardID:                  card.ID,
			AuthorizationID:         authorization.ID,
			OriginCardTransactionID: state.OriginTransactionID,
			Type:                    enums.CardTransactionType_VOID,
			Status:                  req.Status,
			Currency:                authorization.Currency,
			TxCurrency:              authorization.Currency,
			TxAmount:                req.Amount,
			RequestID:               types.Value(req.RequestID),
			MerchantName:            authorization.MerchantName,
			MerchantCountry:         authorization.MerchantCountry,
			MerchantMCC:             authorization.MerchantMCC,
			AuthorizationCode:       authorization.AuthorizationCode,
		}
		if err := simulator.cardTransactionRepo.Create(ctx, transaction); err != nil {
			zap.S().Errorw("create shared simulated reversal transaction",
				"account_id", accountID,
				"channel", req.Channel,
				"authorization_id", authorization.ID,
				"error", err,
			)
			return sharederrors.ErrDatabaseOperation
		}
		result = &CardTransactionSimulationResult{
			Authorization:   authorization,
			CardTransaction: transaction,
			Remaining:       state.Remaining.Sub(req.Amount),
		}
		return nil
	})
	if err != nil {
		if operationErr != nil {
			return nil, operationErr
		}
		if errors.Is(err, sharederrors.ErrNestedTransaction) {
			return nil, err
		}
		zap.S().Errorw("run shared reversal simulation transaction",
			"account_id", accountID,
			"channel", req.Channel,
			"authorization_id", req.AuthorizationID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	simulator.notifyTransaction(ctx, &notifySimulationRequest{
		Result:      result,
		Notificator: req.Notificator,
	})
	return result, nil
}

func (simulator *cardTransactionSimulator) lockAccount(ctx context.Context, req *simulationAccountRequest) error {
	exists, err := simulator.accountRepo.Exist(ctx, &AccountExistRequest{
		ID:      req.AccountID,
		Channel: req.Channel,
	})
	if err != nil {
		zap.S().Errorw("check shared simulation account",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"error", err,
		)
		return sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return sharederrors.ErrAccountNotFound
	}
	if _, err := simulator.accountRepo.FindByIDWithLock(ctx, &AccountFindByIDWithLockRequest{
		ID:      req.AccountID,
		Channel: req.Channel,
	}); err != nil {
		zap.S().Errorw("lock shared simulation account",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"error", err,
		)
		return sharederrors.ErrDatabaseOperation
	}
	return nil
}

func (simulator *cardTransactionSimulator) lockCard(ctx context.Context, req *simulationCardRequest) (*model.Card, error) {
	exists, err := simulator.cardRepo.Exist(ctx, &CardExistRequest{
		ID:        req.CardID,
		AccountID: req.AccountID,
		Channel:   req.Channel,
	})
	if err != nil {
		zap.S().Errorw("check shared simulation card",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"card_id", req.CardID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrCardNotFound
	}
	card, err := simulator.cardRepo.FindByIDWithLock(ctx, &CardFindByIDWithLockRequest{
		ID:        req.CardID,
		AccountID: req.AccountID,
		Channel:   req.Channel,
	})
	if err != nil {
		zap.S().Errorw("lock shared simulation card",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"card_id", req.CardID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return card, nil
}

func (simulator *cardTransactionSimulator) loadAuthorization(ctx context.Context, req *simulationAuthorizationRequest) (*simulationAuthorizationState, error) {
	if req.AuthorizationID <= 0 {
		return nil, sharederrors.ErrInvalidAuthorization
	}
	exists, err := simulator.authorizationRepo.Exist(ctx, &AuthorizationExistRequest{
		ID:        req.AuthorizationID,
		AccountID: req.AccountID,
		Channel:   req.Channel,
		CardID:    req.CardID,
	})
	if err != nil {
		zap.S().Errorw("check shared simulation authorization",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"authorization_id", req.AuthorizationID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrAuthorizationNotFound
	}
	authorization, err := simulator.authorizationRepo.FindByIDWithLock(ctx, &AuthorizationFindByIDWithLockRequest{
		ID:        req.AuthorizationID,
		AccountID: req.AccountID,
		Channel:   req.Channel,
		CardID:    req.CardID,
	})
	if err != nil {
		zap.S().Errorw("lock shared simulation authorization",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"authorization_id", req.AuthorizationID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if authorization.Status != enums.TransactionStatus_AUTHORIZED || !authorization.Amount.IsPositive() || authorization.Currency == "" {
		return nil, sharederrors.ErrInvalidAuthorization
	}
	stages, err := simulator.cardTransactionRepo.ListStages(ctx, &CardTransactionListStagesRequest{
		AccountID:       req.AccountID,
		Channel:         req.Channel,
		CardID:          req.CardID,
		AuthorizationID: req.AuthorizationID,
	})
	if err != nil {
		zap.S().Errorw("list shared simulation authorization stages",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"authorization_id", req.AuthorizationID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	state := &simulationAuthorizationState{
		Authorization: authorization,
		Remaining:     authorization.Amount,
	}
	authorization.CardTransactions = stages
	for _, stage := range stages {
		switch stage.Type {
		case enums.CardTransactionType_AUTH:
			if stage.Status == enums.TransactionStatus_AUTHORIZED && state.OriginTransactionID == 0 {
				state.OriginTransactionID = stage.ID
			}
		case enums.CardTransactionType_CLEAR:
			if stage.Status == enums.TransactionStatus_SUCCEED {
				state.Remaining = state.Remaining.Sub(stage.TxAmount)
			}
		case enums.CardTransactionType_VOID:
			if stage.Status == enums.TransactionStatus_SUCCEED || stage.Status == enums.TransactionStatus_VOID {
				state.Remaining = state.Remaining.Sub(stage.TxAmount)
			}
		}
	}
	return state, nil
}

func (simulator *cardTransactionSimulator) findPreviousTransaction(ctx context.Context, req *simulationRequestReference) (*model.CardTransaction, error) {
	if req.RequestID == nil {
		return nil, nil
	}
	exists, err := simulator.cardTransactionRepo.ExistByRequestID(ctx, &CardTransactionExistByRequestIDRequest{
		AccountID: req.AccountID,
		Channel:   req.Channel,
		RequestID: *req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("check shared simulation request",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"request_id", *req.RequestID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, nil
	}
	transaction, err := simulator.cardTransactionRepo.FindByRequestID(ctx, &CardTransactionFindByRequestIDRequest{
		AccountID: req.AccountID,
		Channel:   req.Channel,
		RequestID: *req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("find shared simulation request",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"request_id", *req.RequestID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return transaction, nil
}

func (simulator *cardTransactionSimulator) replayTransaction(ctx context.Context, transaction *model.CardTransaction) (*CardTransactionSimulationResult, error) {
	state, err := simulator.loadAuthorization(ctx, &simulationAuthorizationRequest{
		AccountID:       transaction.AccountID,
		Channel:         transaction.Channel,
		CardID:          transaction.CardID,
		AuthorizationID: transaction.AuthorizationID,
	})
	if err != nil {
		return nil, err
	}
	return &CardTransactionSimulationResult{
		Authorization:   state.Authorization,
		CardTransaction: transaction,
		Remaining:       state.Remaining,
		Replayed:        true,
	}, nil
}

func (simulator *cardTransactionSimulator) notifyTransaction(ctx context.Context, req *notifySimulationRequest) {
	if req.Result.Replayed {
		return
	}
	transaction := req.Result.CardTransaction
	req.Result.NotificationError = req.Notificator.NotifyCardTransaction(ctx, &NotifyCardTransactionReq{
		AccountID:         transaction.AccountID,
		Channel:           transaction.Channel,
		CardID:            transaction.CardID,
		AuthorizationID:   transaction.AuthorizationID,
		CardTransactionID: transaction.ID,
		Type:              transaction.Type,
	})
	if req.Result.NotificationError != nil {
		zap.S().Errorw("notify shared simulated card transaction",
			"account_id", transaction.AccountID,
			"channel", transaction.Channel,
			"card_transaction_id", transaction.ID,
			"error", req.Result.NotificationError,
		)
	}
}

func (simulator *cardTransactionSimulator) findSimulationCard(ctx context.Context, req *simulationCardReference) (*model.Card, error) {
	exists, err := simulator.cardRepo.ExistForSimulation(ctx, &CardSimulationExistRequest{
		ID:      req.CardID,
		Channel: req.Channel,
	})
	if err != nil {
		zap.S().Errorw("check shared simulation card ownership",
			"channel", req.Channel,
			"card_id", req.CardID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrCardNotFound
	}
	item, err := simulator.cardRepo.FindForSimulation(ctx, &CardSimulationFindRequest{
		ID:      req.CardID,
		Channel: req.Channel,
	})
	if err != nil {
		zap.S().Errorw("find shared simulation card ownership",
			"channel", req.Channel,
			"card_id", req.CardID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return item, nil
}

func (simulator *cardTransactionSimulator) findSimulationAuthorization(ctx context.Context, req *simulationAuthorizationReference) (*model.Authorization, error) {
	exists, err := simulator.authorizationRepo.ExistForSimulation(ctx, &AuthorizationSimulationExistRequest{
		ID:      req.AuthorizationID,
		Channel: req.Channel,
	})
	if err != nil {
		zap.S().Errorw("check shared simulation authorization ownership",
			"channel", req.Channel,
			"authorization_id", req.AuthorizationID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrAuthorizationNotFound
	}
	item, err := simulator.authorizationRepo.FindForSimulation(ctx, &AuthorizationSimulationFindRequest{
		ID:      req.AuthorizationID,
		Channel: req.Channel,
	})
	if err != nil {
		zap.S().Errorw("find shared simulation authorization ownership",
			"channel", req.Channel,
			"authorization_id", req.AuthorizationID,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return item, nil
}
