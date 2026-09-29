package biz

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/cardwallet"
	"generic-mock/pkg/randomx"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"github.com/samber/do/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type IssueCardHolder struct {
	ID           *model.ID
	LastName     *string
	FirstName    *string
	MobilePrefix *string
	Mobile       *string
	Email        *string
}

type IssueCardReq struct {
	IssueCardHolder  *IssueCardHolder
	Channel          enums.Channel
	AccountID        model.ID
	CardType         enums.CardType
	VirtualAccountID *model.ID
	CardProductID    model.ID
	Currency         enums.Currency
	CardScheme       enums.CardScheme
	FormType         enums.CardFormType
	Status           enums.CardStatus
	ExpireAt         time.Time
	RequestID        *string
	InitialAvailable *decimal.Decimal
	RawRequest       json.RawMessage
	Notificator      Notificator
}

type IssueCardResult struct {
	Card              *model.Card
	NotificationError error
}

type CardIssuer struct {
	accountRepo        AccountRepo
	cardRepo           CardRepo
	cardHolderRepo     CardHolderRepo
	walletRepo         WalletRepo
	virtualAccountRepo VirtualAccountRepo
	cardProductRepo    CardProductRepo
	tx                 Transaction
}

func NewCardIssuer(injector do.Injector) (*CardIssuer, error) {
	return &CardIssuer{
		accountRepo:        do.MustInvoke[AccountRepo](injector),
		cardRepo:           do.MustInvoke[CardRepo](injector),
		cardHolderRepo:     do.MustInvoke[CardHolderRepo](injector),
		walletRepo:         do.MustInvoke[WalletRepo](injector),
		virtualAccountRepo: do.MustInvoke[VirtualAccountRepo](injector),
		cardProductRepo:    do.MustInvoke[CardProductRepo](injector),
		tx:                 do.MustInvoke[Transaction](injector),
	}, nil
}

func (issuer *CardIssuer) Issue(ctx context.Context, req *IssueCardReq) (*IssueCardResult, error) {
	if err := validateIssueCardRequest(req); err != nil {
		return nil, err
	}
	var card *model.Card
	var operationErr error
	err := issuer.tx.InTx(ctx, func(txCtx context.Context) error {
		card, operationErr = issuer.issueCard(txCtx, req)
		return operationErr
	})
	if err != nil {
		if operationErr != nil {
			return nil, operationErr
		}
		if errors.Is(err, sharederrors.ErrNestedTransaction) {
			return nil, err
		}
		zap.S().Errorw(
			"run shared card issuance transaction",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}

	result := &IssueCardResult{Card: card}
	result.NotificationError = req.Notificator.NotifyIssueCard(ctx, &NotifyIssueCardReq{
		AccountID: card.AccountID,
		Channel:   card.Channel,
		CardID:    card.ID,
	})
	if result.NotificationError != nil {
		zap.S().Errorw(
			"notify shared card issuance",
			"account_id", card.AccountID,
			"channel", card.Channel,
			"card_id", card.ID,
			"error", result.NotificationError,
		)
	}
	return result, nil
}

func validateIssueCardRequest(req *IssueCardReq) error {
	if req == nil || req.AccountID <= 0 || req.CardProductID <= 0 || req.Channel == "" ||
		req.Currency == "" || req.CardScheme == "" || req.Notificator == nil || !req.ExpireAt.After(time.Now()) {
		return sharederrors.ErrInvalidIssueRequest
	}
	if req.FormType != enums.CardFormType_Virtual && req.FormType != enums.CardFormType_Physical {
		return sharederrors.ErrInvalidIssueRequest
	}
	if req.Status != enums.CardStatus_Active && req.Status != enums.CardStatus_Inactive {
		return sharederrors.ErrInvalidIssueRequest
	}
	if req.RequestID != nil && *req.RequestID == "" {
		return sharederrors.ErrInvalidIssueRequest
	}
	if req.InitialAvailable != nil {
		if req.CardType != enums.CardType_Single || req.InitialAvailable.IsNegative() {
			return sharederrors.ErrInvalidWallet
		}
	}
	if req.RawRequest != nil && !json.Valid(req.RawRequest) {
		return sharederrors.ErrInvalidIssueRequest
	}
	switch req.CardType {
	case enums.CardType_Single:
		if req.VirtualAccountID != nil {
			return sharederrors.ErrInvalidWallet
		}
	case enums.CardType_Share, enums.CardType_VirtualAccountSingle:
		if req.VirtualAccountID == nil || *req.VirtualAccountID <= 0 {
			return sharederrors.ErrInvalidWallet
		}
	default:
		return sharederrors.ErrInvalidIssueRequest
	}
	if req.IssueCardHolder == nil {
		return nil
	}
	holder := req.IssueCardHolder
	if holder.ID != nil && *holder.ID <= 0 {
		return sharederrors.ErrInvalidCardHolder
	}
	hasDetails := false
	for _, value := range []*string{holder.FirstName, holder.LastName, holder.MobilePrefix, holder.Mobile, holder.Email} {
		if value != nil {
			if holder.ID != nil || *value == "" {
				return sharederrors.ErrInvalidCardHolder
			}
			hasDetails = true
		}
	}
	if holder.ID == nil && !hasDetails {
		return sharederrors.ErrInvalidCardHolder
	}
	return nil
}

func (issuer *CardIssuer) issueCard(ctx context.Context, req *IssueCardReq) (*model.Card, error) {
	exists, err := issuer.accountRepo.Exist(ctx, &AccountExistRequest{
		ID:      req.AccountID,
		Channel: req.Channel,
	})
	if err != nil {
		zap.S().Errorw(
			"check shared card account",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrAccountNotFound
	}

	exists, err = issuer.cardProductRepo.Exist(ctx, &CardProductExistRequest{
		ID:      req.CardProductID,
		Channel: req.Channel,
	})
	if err != nil {
		zap.S().Errorw(
			"check shared card product",
			"card_product_id", req.CardProductID,
			"channel", req.Channel,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrCardProductNotFound
	}
	product, err := issuer.cardProductRepo.FindByIDWithLock(ctx, &CardProductFindByIDWithLockRequest{
		ID:      req.CardProductID,
		Channel: req.Channel,
	})
	if err != nil {
		zap.S().Errorw(
			"lock shared card product",
			"card_product_id", req.CardProductID,
			"channel", req.Channel,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if product.ID != req.CardProductID || product.Channel != req.Channel || product.NextCardNumber < 0 || product.NextCardNumber == math.MaxInt64 {
		return nil, sharederrors.ErrInvalidCardProduct
	}
	nextSequence := product.NextCardNumber + 1
	generated, ok := cardnumber.Generate(cardnumber.GenerateRequest{
		Channel:  req.Channel,
		Prefix:   product.Prefix,
		Sequence: nextSequence,
	})
	if !ok {
		return nil, sharederrors.ErrInvalidCardProduct
	}

	assignment, err := issuer.prepareCardWallet(ctx, req)
	if err != nil {
		return nil, err
	}
	if req.InitialAvailable != nil {
		assignment.Wallet.Available = *req.InitialAvailable
		assignment.Wallet.In = *req.InitialAvailable
	}
	holderID, err := issuer.resolveCardHolderID(ctx, req)
	if err != nil {
		return nil, err
	}
	if assignment.CreateWallet {
		if err := issuer.walletRepo.Create(ctx, &WalletCreateRequest{Wallet: assignment.Wallet}); err != nil {
			zap.S().Errorw(
				"create shared card wallet",
				"account_id", req.AccountID,
				"channel", req.Channel,
				"error", err,
			)
			return nil, sharederrors.ErrDatabaseOperation
		}
	}
	if err := issuer.cardProductRepo.UpdateSeq(ctx, &CardProductUpdateSeqRequest{
		ID:           product.ID,
		Channel:      req.Channel,
		NextSequence: nextSequence,
	}); err != nil {
		zap.S().Errorw(
			"advance shared card product sequence",
			"card_product_id", product.ID,
			"channel", req.Channel,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}

	rawRequest := json.RawMessage(`{}`)
	if req.RawRequest != nil {
		rawRequest = slices.Clone(req.RawRequest)
	}
	card := &model.Card{
		AccountID:              req.AccountID,
		Channel:                req.Channel,
		CardProductID:          product.ID,
		CardBin:                generated.Bin,
		CardNumber:             generated.Number,
		Cvv:                    randomx.Digits(3),
		ExpireAt:               req.ExpireAt,
		Status:                 req.Status,
		VirtualAccountID:       assignment.VirtualAccountID,
		WalletID:               assignment.Wallet.ID,
		CardHolderID:           holderID,
		FormType:               req.FormType,
		RequestID:              types.Value(req.RequestID),
		LastOperationRequestID: types.Value(req.RequestID),
		LastOperationType:      enums.OperationType_OpenCard,
		LastOperationStatus:    enums.OperationStatus_Succeed,
		CardCurrency:           req.Currency,
		CardScheme:             req.CardScheme,
		CardType:               req.CardType,
		RawRequest:             rawRequest,
		Wallet:                 assignment.Wallet,
	}
	if err := issuer.cardRepo.Create(ctx, &CardCreateRequest{Card: card}); err != nil {
		zap.S().Errorw(
			"create shared card",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"error", err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return card, nil
}

func (issuer *CardIssuer) resolveCardHolderID(ctx context.Context, req *IssueCardReq) (model.ID, error) {
	holder := req.IssueCardHolder
	if holder == nil {
		return 0, nil
	}
	if holder.ID != nil {
		exists, err := issuer.cardHolderRepo.Exist(ctx, &CardHolderExistRequest{
			ID:        *holder.ID,
			AccountID: req.AccountID,
			Channel:   req.Channel,
		})
		if err != nil {
			zap.S().Errorw(
				"check shared card holder",
				"account_id", req.AccountID,
				"channel", req.Channel,
				"card_holder_id", *holder.ID,
				"error", err,
			)
			return 0, sharederrors.ErrDatabaseOperation
		}
		if !exists {
			return 0, sharederrors.ErrCardHolderNotFound
		}
		return *holder.ID, nil
	}
	item := &model.CardHolder{
		AccountID:    req.AccountID,
		Channel:      req.Channel,
		FirstName:    types.Value(holder.FirstName),
		LastName:     types.Value(holder.LastName),
		Email:        types.Value(holder.Email),
		Mobile:       types.Value(holder.Mobile),
		MobilePrefix: types.Value(holder.MobilePrefix),
		Status:       enums.CardHolderStatus_Normal,
		ReviewStatus: enums.CardHolderReviewStatus_Approved,
		Shared:       false,
	}
	if err := issuer.cardHolderRepo.Create(ctx, &CardHolderCreateRequest{CardHolder: item}); err != nil {
		zap.S().Errorw(
			"create shared card holder",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"error", err,
		)
		return 0, sharederrors.ErrDatabaseOperation
	}
	return item.ID, nil
}

func (issuer *CardIssuer) prepareCardWallet(ctx context.Context, req *IssueCardReq) (cardwallet.PrepareResult, error) {
	var virtualAccount *model.VirtualAccount
	if req.VirtualAccountID != nil {
		exists, err := issuer.virtualAccountRepo.Exist(ctx, &VirtualAccountExistRequest{
			ID:        *req.VirtualAccountID,
			AccountID: req.AccountID,
			Channel:   req.Channel,
		})
		if err != nil {
			zap.S().Errorw(
				"check shared card virtual account",
				"account_id", req.AccountID,
				"channel", req.Channel,
				"virtual_account_id", *req.VirtualAccountID,
				"error", err,
			)
			return cardwallet.PrepareResult{}, sharederrors.ErrDatabaseOperation
		}
		if !exists {
			return cardwallet.PrepareResult{}, sharederrors.ErrVirtualAccountNotFound
		}
		virtualAccount, err = issuer.virtualAccountRepo.Find(ctx, &VirtualAccountFindRequest{
			ID:        *req.VirtualAccountID,
			AccountID: req.AccountID,
			Channel:   req.Channel,
		})
		if err != nil {
			zap.S().Errorw(
				"find shared card virtual account",
				"account_id", req.AccountID,
				"channel", req.Channel,
				"virtual_account_id", *req.VirtualAccountID,
				"error", err,
			)
			return cardwallet.PrepareResult{}, sharederrors.ErrDatabaseOperation
		}
		if virtualAccount.ID != *req.VirtualAccountID {
			return cardwallet.PrepareResult{}, sharederrors.ErrInvalidWallet
		}
	}
	assignment, ok := cardwallet.Prepare(cardwallet.PrepareRequest{
		AccountID:      req.AccountID,
		Channel:        req.Channel,
		CardType:       req.CardType,
		Currency:       req.Currency,
		VirtualAccount: virtualAccount,
	})
	if !ok {
		return cardwallet.PrepareResult{}, sharederrors.ErrInvalidWallet
	}
	return assignment, nil
}
