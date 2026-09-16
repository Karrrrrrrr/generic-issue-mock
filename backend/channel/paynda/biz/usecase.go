package biz

import (
	"context"
	"time"

	paynda "generic-mock/channel/paynda/enums"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/randomx"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type PayndaTransaction interface {
	InTx(context.Context, func(context.Context) error) error
}

type PayndaCardHolderRepository interface {
	Create(context.Context, *model.CardHolder) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardHolder, error)
	List(context.Context, *PayndaListRequest) ([]*model.CardHolder, error)
	Save(context.Context, *model.CardHolder) error
}

type PayndaCardProductRepository interface {
	ExistByID(context.Context, model.ID) (bool, error)
	FindByIDForUpdate(context.Context, model.ID) (*model.CardProduct, error)
	List(context.Context) ([]*model.CardProduct, error)
	Save(context.Context, *model.CardProduct) error
}

type PayndaCardRepository interface {
	Create(context.Context, *model.Card) error
	ExistByID(context.Context, model.ID) (bool, error)
	ExistByRequestID(context.Context, string) (bool, error)
	ExistByLastOperationRequestID(context.Context, string) (bool, error)
	FindByID(context.Context, model.ID) (*model.Card, error)
	FindByRequestID(context.Context, string) (*model.Card, error)
	FindByLastOperationRequestID(context.Context, string) (*model.Card, error)
	List(context.Context, *PayndaListRequest) ([]*model.Card, error)
	Save(context.Context, *model.Card) error
}

type PayndaWalletRepository interface {
	Create(context.Context, *model.Wallet) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.Wallet, error)
	FindByIDForUpdate(context.Context, model.ID) (*model.Wallet, error)
	Save(context.Context, *model.Wallet) error
}

type PayndaAccountRepository interface {
	Create(context.Context, *model.Account) error
	ExistByChannel(context.Context) (bool, error)
	FindByChannel(context.Context) (*model.Account, error)
}

type PayndaCardTransactionRepository interface {
	Create(context.Context, *model.CardTransaction) error
	Save(context.Context, *model.CardTransaction) error
	ExistByID(context.Context, model.ID) (bool, error)
	ExistByRequestID(context.Context, string) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardTransaction, error)
	FindByRequestID(context.Context, string) (*model.CardTransaction, error)
	List(context.Context, *PayndaListTransactionsRequest) ([]*model.CardTransaction, error)
}

type PayndaAuthorizationRepository interface {
	Create(context.Context, *model.Authorization) error
	List(context.Context, *PayndaListRequest) ([]*model.Authorization, error)
}

type PayndaListRequest struct {
	Offset int
	Limit  int
}

type PayndaListTransactionsRequest struct {
	PayndaListRequest
	CardID        model.ID
	OccurredAtGTE *time.Time
	OccurredAtLTE *time.Time
	Types         []enums.CardTransactionType
}

type PayndaOpenAPIUsecase struct {
	transaction               PayndaTransaction
	cardHolderRepository      PayndaCardHolderRepository
	cardProductRepository     PayndaCardProductRepository
	cardRepository            PayndaCardRepository
	walletRepository          PayndaWalletRepository
	accountRepository         PayndaAccountRepository
	cardTransactionRepository PayndaCardTransactionRepository
}

func NewPayndaOpenAPIUsecase(injector *do.Injector) (*PayndaOpenAPIUsecase, error) {
	return &PayndaOpenAPIUsecase{
		transaction:               do.MustInvoke[PayndaTransaction](injector),
		cardHolderRepository:      do.MustInvoke[PayndaCardHolderRepository](injector),
		cardProductRepository:     do.MustInvoke[PayndaCardProductRepository](injector),
		cardRepository:            do.MustInvoke[PayndaCardRepository](injector),
		walletRepository:          do.MustInvoke[PayndaWalletRepository](injector),
		accountRepository:         do.MustInvoke[PayndaAccountRepository](injector),
		cardTransactionRepository: do.MustInvoke[PayndaCardTransactionRepository](injector),
	}, nil
}

type PayndaCreateCardHolderRequest struct {
	FirstName              string
	LastName               string
	MobilePrefix           string
	Mobile                 string
	Email                  string
	ResidentialAddress     string
	ResidentialCity        string
	ResidentialCountryCode string
	ResidentialPostalCode  string
	ResidentialState       string
}

func (u *PayndaOpenAPIUsecase) CreateCardHolder(
	ctx context.Context,
	req *PayndaCreateCardHolderRequest,
) (*model.CardHolder, error) {
	holder := &model.CardHolder{
		Channel:                enums.Channel_Paynda,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		MobilePrefix:           req.MobilePrefix,
		Mobile:                 req.Mobile,
		Email:                  req.Email,
		ResidentialAddress:     req.ResidentialAddress,
		ResidentialCity:        req.ResidentialCity,
		ResidentialCountryCode: req.ResidentialCountryCode,
		ResidentialPostalCode:  req.ResidentialPostalCode,
		ResidentialState:       req.ResidentialState,
		Status:                 enums.CardHolderStatus_Normal,
		ReviewStatus:           enums.CardHolderReviewStatus_Approved,
		Shared:                 true,
	}
	if err := u.cardHolderRepository.Create(ctx, holder); err != nil {
		zap.S().Errorw("create paynda card holder", "error", err)
		return nil, ErrDatabaseOperation
	}

	return holder, nil
}

func (u *PayndaOpenAPIUsecase) GetCardHolder(ctx context.Context, id model.ID) (*model.CardHolder, error) {
	if err := u.requireCardHolder(ctx, id); err != nil {
		return nil, err
	}

	holder, err := u.cardHolderRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find paynda card holder", "error", err)
		return nil, ErrDatabaseOperation
	}

	return holder, nil
}

type PayndaUpdateCardHolderRequest struct {
	ID                     model.ID
	FirstName              string
	LastName               string
	MobilePrefix           string
	Mobile                 string
	Email                  string
	ResidentialAddress     string
	ResidentialCity        string
	ResidentialCountryCode string
	ResidentialPostalCode  string
	ResidentialState       string
}

func (u *PayndaOpenAPIUsecase) UpdateCardHolder(
	ctx context.Context,
	req *PayndaUpdateCardHolderRequest,
) (*model.CardHolder, error) {
	var holder *model.CardHolder
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCardHolder(txCtx, req.ID); err != nil {
			return err
		}

		var err error
		holder, err = u.cardHolderRepository.FindByID(txCtx, req.ID)
		if err != nil {
			zap.S().Errorw("find paynda card holder for update", "error", err)
			return ErrDatabaseOperation
		}
		holder.FirstName = req.FirstName
		holder.LastName = req.LastName
		holder.MobilePrefix = req.MobilePrefix
		holder.Mobile = req.Mobile
		holder.Email = req.Email
		holder.ResidentialAddress = req.ResidentialAddress
		holder.ResidentialCity = req.ResidentialCity
		holder.ResidentialCountryCode = req.ResidentialCountryCode
		holder.ResidentialPostalCode = req.ResidentialPostalCode
		holder.ResidentialState = req.ResidentialState
		if err := u.cardHolderRepository.Save(txCtx, holder); err != nil {
			zap.S().Errorw("update paynda card holder", "error", err)
			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return holder, nil
}

func (u *PayndaOpenAPIUsecase) ListCardHolders(
	ctx context.Context,
	req *PayndaListRequest,
) ([]*model.CardHolder, error) {
	items, err := u.cardHolderRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list paynda card holders", "error", err)
		return nil, ErrDatabaseOperation
	}

	return items, nil
}

type PayndaCreateCardRequest struct {
	CardHolderID  model.ID
	CardProductID model.ID
	Currency      enums.Currency
	ExpireAt      time.Time
	RequestID     string
}

func (u *PayndaOpenAPIUsecase) CreateCard(ctx context.Context, req *PayndaCreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		if err := u.requireCardHolder(txCtx, req.CardHolderID); err != nil {
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
		cardNumber, ok := cardnumber.Generate(product.Prefix, product.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance paynda card product sequence", "error", err)
			return ErrDatabaseOperation
		}

		wallet := &model.Wallet{
			Amount:   decimal.Zero,
			Type:     enums.WalletType_Card,
			Currency: req.Currency,
		}
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create paynda card wallet", "error", err)
			return ErrDatabaseOperation
		}

		card = &model.Card{
			Channel:                enums.Channel_Paynda,
			CardProductID:          product.ID,
			CardBin:                product.Prefix,
			CardNumber:             cardNumber,
			Cvv:                    randomx.Digits(3),
			ExpireAt:               req.ExpireAt,
			Status:                 enums.CardStatus_Active,
			WalletID:               &wallet.ID,
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

func (u *PayndaOpenAPIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	items, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list paynda card products", "error", err)
		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaOpenAPIUsecase) GetCard(ctx context.Context, id model.ID) (*model.Card, error) {
	if err := u.requireCard(ctx, id); err != nil {
		return nil, err
	}

	card, err := u.cardRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find paynda card", "error", err)
		return nil, ErrDatabaseOperation
	}

	return card, nil
}

func (u *PayndaOpenAPIUsecase) ListCards(ctx context.Context, req *PayndaListRequest) ([]*model.Card, error) {
	items, err := u.cardRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list paynda cards", "error", err)
		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaOpenAPIUsecase) GetCardBalance(ctx context.Context, cardID model.ID) (*model.Wallet, error) {
	card, err := u.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if card.WalletID == nil {
		return nil, ErrResourceNotFound
	}

	exists, err := u.walletRepository.ExistByID(ctx, *card.WalletID)
	if err != nil {
		zap.S().Errorw("check paynda card wallet", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	wallet, err := u.walletRepository.FindByID(ctx, *card.WalletID)
	if err != nil {
		zap.S().Errorw("find paynda card wallet", "error", err)
		return nil, ErrDatabaseOperation
	}

	return wallet, nil
}

type PayndaAccountWallet struct {
	Account *model.Account
	Wallet  *model.Wallet
}

func (u *PayndaOpenAPIUsecase) GetAccountWallet(ctx context.Context) (*PayndaAccountWallet, error) {
	var result *PayndaAccountWallet
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.accountRepository.ExistByChannel(txCtx)
		if err != nil {
			zap.S().Errorw("check paynda account", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			wallet := &model.Wallet{
				Amount:   decimal.Zero,
				Type:     enums.WalletType_Account,
				Currency: enums.Currency_USD,
			}
			if err := u.walletRepository.Create(txCtx, wallet); err != nil {
				zap.S().Errorw("create paynda account wallet", "error", err)
				return ErrDatabaseOperation
			}
			account := &model.Account{
				Channel:  enums.Channel_Paynda,
				Name:     "Paynda",
				WalletID: wallet.ID,
			}
			if err := u.accountRepository.Create(txCtx, account); err != nil {
				zap.S().Errorw("create paynda account", "error", err)
				return ErrDatabaseOperation
			}
			result = &PayndaAccountWallet{
				Account: account,
				Wallet:  wallet,
			}
			return nil
		}

		account, err := u.accountRepository.FindByChannel(txCtx)
		if err != nil {
			zap.S().Errorw("find paynda account", "error", err)
			return ErrDatabaseOperation
		}
		wallet, err := u.walletRepository.FindByID(txCtx, account.WalletID)
		if err != nil {
			zap.S().Errorw("find paynda account wallet", "error", err)
			return ErrDatabaseOperation
		}
		result = &PayndaAccountWallet{
			Account: account,
			Wallet:  wallet,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

type PayndaAccountWalletTransferRequest struct {
	Amount decimal.Decimal
}

func (u *PayndaOpenAPIUsecase) TransferAccountWallet(
	ctx context.Context,
	req *PayndaAccountWalletTransferRequest,
) (*PayndaAccountWallet, error) {
	accountWallet, err := u.GetAccountWallet(ctx)
	if err != nil {
		return nil, err
	}

	err = u.transaction.InTx(ctx, func(txCtx context.Context) error {
		wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, accountWallet.Wallet.ID)
		if err != nil {
			zap.S().Errorw("lock paynda account wallet", "error", err)
			return ErrDatabaseOperation
		}
		wallet.Amount = wallet.Amount.Add(req.Amount)
		wallet.In = wallet.In.Add(req.Amount)
		if err := u.walletRepository.Save(txCtx, wallet); err != nil {
			zap.S().Errorw("save paynda account wallet", "error", err)
			return ErrDatabaseOperation
		}
		accountWallet.Wallet = wallet

		return nil
	})
	if err != nil {
		return nil, err
	}

	return accountWallet, nil
}

type PayndaTransferRequest struct {
	CardID    model.ID
	RequestID string
	Amount    decimal.Decimal
	Type      paynda.TransferType
}

func (u *PayndaOpenAPIUsecase) TransferCardBalance(
	ctx context.Context,
	req *PayndaTransferRequest,
) (*model.CardTransaction, error) {
	var transaction *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		card, err := u.GetCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.WalletID == nil {
			return ErrResourceNotFound
		}

		wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, *card.WalletID)
		if err != nil {
			zap.S().Errorw("lock paynda card wallet", "error", err)
			return ErrDatabaseOperation
		}
		oldAmount := wallet.Amount
		transactionType := enums.CardTransactionType_FundIn
		if req.Type == paynda.TransferType_Out {
			if wallet.Amount.LessThan(req.Amount) {
				return ErrInvalidOperation
			}
			wallet.Amount = wallet.Amount.Sub(req.Amount)
			wallet.Out = wallet.Out.Add(req.Amount)
			transactionType = enums.CardTransactionType_FundOut
		} else {
			wallet.Amount = wallet.Amount.Add(req.Amount)
			wallet.In = wallet.In.Add(req.Amount)
		}
		if err := u.walletRepository.Save(txCtx, wallet); err != nil {
			zap.S().Errorw("save paynda card wallet", "error", err)
			return ErrDatabaseOperation
		}

		transaction = &model.CardTransaction{
			Channel:    enums.Channel_Paynda,
			CardID:     card.ID,
			Status:     enums.TransactionStatus_SUCCEED,
			Type:       transactionType,
			Currency:   wallet.Currency,
			TxAmount:   req.Amount,
			TxCurrency: wallet.Currency,
			RequestID:  req.RequestID,
			OccurredAt: time.Now().UTC(),
			RawPayload: []byte(oldAmount.String()),
		}
		if err := u.cardTransactionRepository.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create paynda card balance transfer", "error", err)
			return ErrDatabaseOperation
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

type PayndaUpdateCardStatusRequest struct {
	CardID    model.ID
	RequestID string
	Status    paynda.CardStatus
}

func (u *PayndaOpenAPIUsecase) UpdateCardStatus(
	ctx context.Context,
	req *PayndaUpdateCardStatusRequest,
) (*model.Card, error) {
	if req.RequestID == "" {
		req.RequestID = randomx.Digits(20)
	}

	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		var err error
		card, err = u.GetCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		card.Status = paynda.CardStatusToGeneric(req.Status)
		card.LastOperationRequestID = req.RequestID
		card.LastOperationType = enums.OperationType_UpdateCard
		card.LastOperationStatus = enums.OperationStatus_Succeed
		if err := u.cardRepository.Save(txCtx, card); err != nil {
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

func (u *PayndaOpenAPIUsecase) ReleaseCard(
	ctx context.Context,
	cardID model.ID,
	requestID string,
) (*model.Card, error) {
	return u.UpdateCardStatus(ctx, &PayndaUpdateCardStatusRequest{
		CardID:    cardID,
		RequestID: requestID,
		Status:    paynda.CardStatus_Deleted,
	})
}

func (u *PayndaOpenAPIUsecase) GetCardTransaction(
	ctx context.Context,
	id model.ID,
) (*model.CardTransaction, error) {
	exists, err := u.cardTransactionRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check paynda card transaction", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	transaction, err := u.cardTransactionRepository.FindByID(ctx, id)
	if err != nil {
		zap.S().Errorw("find paynda card transaction", "error", err)
		return nil, ErrDatabaseOperation
	}

	return transaction, nil
}

func (u *PayndaOpenAPIUsecase) ListCardTransactions(
	ctx context.Context,
	req *PayndaListTransactionsRequest,
) ([]*model.CardTransaction, error) {
	items, err := u.cardTransactionRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list paynda card transactions", "error", err)
		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaOpenAPIUsecase) ListCardBalanceUpdates(
	ctx context.Context,
	req *PayndaListRequest,
) ([]*model.CardTransaction, error) {
	return u.ListCardTransactions(ctx, &PayndaListTransactionsRequest{
		PayndaListRequest: *req,
		Types: []enums.CardTransactionType{
			enums.CardTransactionType_FundIn,
			enums.CardTransactionType_FundOut,
		},
	})
}

type PayndaRequestResult struct {
	Card         *model.Card
	Transaction  *model.CardTransaction
	IsCardCreate bool
}

func (u *PayndaOpenAPIUsecase) FindRequestResult(
	ctx context.Context,
	requestID string,
) (*PayndaRequestResult, error) {
	exists, err := u.cardRepository.ExistByRequestID(ctx, requestID)
	if err != nil {
		zap.S().Errorw("check paynda create card request", "error", err)
		return nil, ErrDatabaseOperation
	}
	if exists {
		card, err := u.cardRepository.FindByRequestID(ctx, requestID)
		if err != nil {
			zap.S().Errorw("find paynda create card request", "error", err)
			return nil, ErrDatabaseOperation
		}
		return &PayndaRequestResult{Card: card, IsCardCreate: true}, nil
	}

	exists, err = u.cardRepository.ExistByLastOperationRequestID(ctx, requestID)
	if err != nil {
		zap.S().Errorw("check paynda card operation request", "error", err)
		return nil, ErrDatabaseOperation
	}
	if exists {
		card, err := u.cardRepository.FindByLastOperationRequestID(ctx, requestID)
		if err != nil {
			zap.S().Errorw("find paynda card operation request", "error", err)
			return nil, ErrDatabaseOperation
		}
		return &PayndaRequestResult{Card: card}, nil
	}

	exists, err = u.cardTransactionRepository.ExistByRequestID(ctx, requestID)
	if err != nil {
		zap.S().Errorw("check paynda card balance transfer request", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	transaction, err := u.cardTransactionRepository.FindByRequestID(ctx, requestID)
	if err != nil {
		zap.S().Errorw("find paynda card balance transfer request", "error", err)
		return nil, ErrDatabaseOperation
	}

	return &PayndaRequestResult{Transaction: transaction}, nil
}

func (u *PayndaOpenAPIUsecase) requireCardHolder(ctx context.Context, id model.ID) error {
	exists, err := u.cardHolderRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check paynda card holder", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}

func (u *PayndaOpenAPIUsecase) requireCardProduct(ctx context.Context, id model.ID) error {
	exists, err := u.cardProductRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check paynda card product", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}

func (u *PayndaOpenAPIUsecase) requireCard(ctx context.Context, id model.ID) error {
	exists, err := u.cardRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check paynda card", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}

	return nil
}

type PayndaUIUsecase struct {
	transaction               PayndaTransaction
	cardRepository            PayndaCardRepository
	cardHolderRepository      PayndaCardHolderRepository
	cardProductRepository     PayndaCardProductRepository
	walletRepository          PayndaWalletRepository
	cardTransactionRepository PayndaCardTransactionRepository
	authorizationRepository   PayndaAuthorizationRepository
}

func (u *PayndaUIUsecase) ListAuthorizations(ctx context.Context, req *PayndaListRequest) ([]*model.Authorization, error) {
	items, err := u.authorizationRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list paynda UI authorizations", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

func NewPayndaUIUsecase(injector *do.Injector) (*PayndaUIUsecase, error) {
	return &PayndaUIUsecase{
		transaction:               do.MustInvoke[PayndaTransaction](injector),
		cardRepository:            do.MustInvoke[PayndaCardRepository](injector),
		cardHolderRepository:      do.MustInvoke[PayndaCardHolderRepository](injector),
		cardProductRepository:     do.MustInvoke[PayndaCardProductRepository](injector),
		walletRepository:          do.MustInvoke[PayndaWalletRepository](injector),
		cardTransactionRepository: do.MustInvoke[PayndaCardTransactionRepository](injector),
		authorizationRepository:   do.MustInvoke[PayndaAuthorizationRepository](injector),
	}, nil
}

type PayndaUICreateCardHolderRequest struct {
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (u *PayndaUIUsecase) CreateCardHolder(ctx context.Context, req *PayndaUICreateCardHolderRequest) (*model.CardHolder, error) {
	holder := &model.CardHolder{
		Channel:      enums.Channel_Paynda,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Mobile:       req.Mobile,
		Status:       enums.CardHolderStatus_Normal,
		ReviewStatus: enums.CardHolderReviewStatus_Approved,
		Shared:       true,
	}
	if err := u.cardHolderRepository.Create(ctx, holder); err != nil {
		zap.S().Errorw("create paynda UI card holder", "error", err)
		return nil, ErrDatabaseOperation
	}
	return holder, nil
}

type PayndaUICreateCardRequest struct {
	CardHolderID model.ID
	Currency     enums.Currency
}

func (u *PayndaUIUsecase) CreateCard(ctx context.Context, req *PayndaUICreateCardRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardHolderRepository.ExistByID(txCtx, req.CardHolderID)
		if err != nil {
			zap.S().Errorw("check paynda UI card holder", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		products, err := u.cardProductRepository.List(txCtx)
		if err != nil {
			zap.S().Errorw("list paynda UI card products", "error", err)
			return ErrDatabaseOperation
		}
		if len(products) == 0 {
			return ErrResourceNotFound
		}

		var defaultProduct *model.CardProduct
		for _, item := range products {
			if item.IsDefault {
				defaultProduct = item
				break
			}
		}
		if defaultProduct == nil {
			return ErrResourceNotFound
		}

		product, err := u.cardProductRepository.FindByIDForUpdate(txCtx, defaultProduct.ID)
		if err != nil {
			zap.S().Errorw("lock paynda UI card product", "error", err)
			return ErrDatabaseOperation
		}
		product.NextCardNumber++
		cardNumber, ok := cardnumber.Generate(product.Prefix, product.NextCardNumber)
		if !ok {
			return ErrInvalidOperation
		}
		if err := u.cardProductRepository.Save(txCtx, product); err != nil {
			zap.S().Errorw("advance paynda UI card product sequence", "error", err)
			return ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			Amount:   decimal.Zero,
			Type:     enums.WalletType_Card,
			Currency: req.Currency,
		}
		if err := u.walletRepository.Create(txCtx, wallet); err != nil {
			zap.S().Errorw("create paynda UI card wallet", "error", err)
			return ErrDatabaseOperation
		}
		card = &model.Card{
			Channel:       enums.Channel_Paynda,
			CardProductID: product.ID,
			CardBin:       product.Prefix,
			CardNumber:    cardNumber,
			Cvv:           randomx.Digits(3),
			ExpireAt:      time.Now().UTC().AddDate(2, 0, 0),
			Status:        enums.CardStatus_Active,
			WalletID:      &wallet.ID,
			CardHolderID:  req.CardHolderID,
			FormType:      enums.CardFormType_Virtual,
			CardCurrency:  req.Currency,
			CardScheme:    enums.CardScheme_MasterCard,
			CardType:      enums.CardType_Single,
		}
		if err := u.cardRepository.Create(txCtx, card); err != nil {
			zap.S().Errorw("create paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return card, nil
}

type PayndaUIUpdateCardStatusRequest struct {
	CardID model.ID
	Status enums.CardStatus
}

func (u *PayndaUIUsecase) UpdateCardStatus(ctx context.Context, req *PayndaUIUpdateCardStatusRequest) (*model.Card, error) {
	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("check paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		card, err = u.cardRepository.FindByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("find paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		card.Status = req.Status
		if err := u.cardRepository.Save(txCtx, card); err != nil {
			zap.S().Errorw("update paynda UI card status", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (u *PayndaUIUsecase) ListCards(ctx context.Context, req *PayndaListRequest) ([]*model.Card, error) {
	items, err := u.cardRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list paynda UI cards", "error", err)
		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaUIUsecase) ListCardHolders(ctx context.Context, req *PayndaListRequest) ([]*model.CardHolder, error) {
	items, err := u.cardHolderRepository.List(ctx, req)
	if err != nil {
		zap.S().Errorw("list paynda UI card holders", "error", err)
		return nil, ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaUIUsecase) ListTransactions(ctx context.Context, req *PayndaListRequest) ([]*model.CardTransaction, error) {
	items, err := u.cardTransactionRepository.List(ctx, &PayndaListTransactionsRequest{
		PayndaListRequest: *req,
	})
	if err != nil {
		zap.S().Errorw("list paynda UI transactions", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
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

type PayndaSimulateRefundRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
}

func (u *PayndaUIUsecase) SimulateAuthorization(
	ctx context.Context,
	req *PayndaSimulateAuthorizationRequest,
) (*PayndaSimulateAuthorizationResult, error) {
	var result *PayndaSimulateAuthorizationResult
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("check paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}

		card, err := u.cardRepository.FindByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("find paynda UI card", "error", err)
			return ErrDatabaseOperation
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}

		now := time.Now().UTC()
		authorization := &model.Authorization{
			Channel:           enums.Channel_Paynda,
			CardID:            card.ID,
			Currency:          req.Currency,
			Amount:            req.Amount,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: randomx.Digits(6),
			Status:            enums.TransactionStatus_AUTHORIZED,
			OccurredAt:        now,
		}
		if err := u.authorizationRepository.Create(txCtx, authorization); err != nil {
			zap.S().Errorw("create paynda UI authorization", "error", err)
			return ErrDatabaseOperation
		}
		transaction := &model.CardTransaction{
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
			OccurredAt:        now,
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

	return result, nil
}

// SimulateRefund creates a posted refund directly for an active card.
func (u *PayndaUIUsecase) SimulateRefund(ctx context.Context, req *PayndaSimulateRefundRequest) (*model.CardTransaction, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}
	var transaction *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardRepository.ExistByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("check paynda UI card for simulated refund", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		card, err := u.cardRepository.FindByID(txCtx, req.CardID)
		if err != nil {
			zap.S().Errorw("find paynda UI card for simulated refund", "error", err)
			return ErrDatabaseOperation
		}
		if card.Status != enums.CardStatus_Active {
			return ErrInvalidOperation
		}
		transaction = &model.CardTransaction{
			Channel:           enums.Channel_Paynda,
			CardID:            card.ID,
			Status:            enums.TransactionStatus_SUCCEED,
			Type:              enums.CardTransactionType_REFUND,
			Currency:          req.Currency,
			TxAmount:          req.Amount,
			TxCurrency:        req.Currency,
			MerchantName:      req.MerchantName,
			MerchantCountry:   req.MerchantCountry,
			MerchantMCC:       req.MerchantMCC,
			AuthorizationCode: randomx.Digits(6),
			OccurredAt:        time.Now().UTC(),
		}
		if err := u.cardTransactionRepository.Create(txCtx, transaction); err != nil {
			zap.S().Errorw("create paynda UI simulated refund", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

type PayndaUIApplyTransactionStepRequest struct {
	CardTransactionID model.ID
	Type              enums.CardTransactionType
	Amount            decimal.Decimal
}

func (u *PayndaUIUsecase) ApplyTransactionStep(ctx context.Context, req *PayndaUIApplyTransactionStepRequest) (*model.CardTransaction, error) {
	var next *model.CardTransaction
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		exists, err := u.cardTransactionRepository.ExistByID(txCtx, req.CardTransactionID)
		if err != nil {
			zap.S().Errorw("check paynda UI transaction", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		origin, err := u.cardTransactionRepository.FindByID(txCtx, req.CardTransactionID)
		if err != nil {
			zap.S().Errorw("find paynda UI transaction", "error", err)
			return ErrDatabaseOperation
		}
		amount := req.Amount
		if amount.IsZero() {
			amount = origin.TxAmount
		}
		if !amount.IsPositive() || amount.GreaterThan(origin.TxAmount) {
			return ErrInvalidOperation
		}
		if origin.CardID == 0 {
			return ErrInvalidOperation
		}

		card, err := u.cardRepository.FindByID(txCtx, origin.CardID)
		if err != nil {
			zap.S().Errorw("find paynda UI transaction card", "error", err)
			return ErrDatabaseOperation
		}
		if card.WalletID == nil {
			return ErrResourceNotFound
		}
		wallet, err := u.walletRepository.FindByIDForUpdate(txCtx, *card.WalletID)
		if err != nil {
			zap.S().Errorw("lock paynda UI card wallet", "error", err)
			return ErrDatabaseOperation
		}

		status := enums.TransactionStatus_SUCCEED
		switch req.Type {
		case enums.CardTransactionType_CLEAR:
			if origin.Type != enums.CardTransactionType_AUTH || origin.Status != enums.TransactionStatus_AUTHORIZED || wallet.Amount.LessThan(amount) {
				return ErrInvalidOperation
			}
			wallet.Amount = wallet.Amount.Sub(amount)
			wallet.Out = wallet.Out.Add(amount)
			origin.Status = enums.TransactionStatus_SUCCEED
		case enums.CardTransactionType_VOID:
			if origin.Type != enums.CardTransactionType_AUTH || origin.Status != enums.TransactionStatus_AUTHORIZED {
				return ErrInvalidOperation
			}
			status = enums.TransactionStatus_VOID
			origin.Status = enums.TransactionStatus_VOID
		case enums.CardTransactionType_REFUND:
			if origin.Type != enums.CardTransactionType_CLEAR || origin.Status != enums.TransactionStatus_SUCCEED {
				return ErrInvalidOperation
			}
			wallet.Amount = wallet.Amount.Add(amount)
			wallet.In = wallet.In.Add(amount)
			origin.Status = enums.TransactionStatus_VOID
		default:
			return ErrInvalidOperation
		}
		if err := u.walletRepository.Save(txCtx, wallet); err != nil {
			zap.S().Errorw("save paynda UI card wallet", "error", err)
			return ErrDatabaseOperation
		}
		if err := u.cardTransactionRepository.Save(txCtx, origin); err != nil {
			zap.S().Errorw("update paynda UI origin transaction", "error", err)
			return ErrDatabaseOperation
		}
		next = &model.CardTransaction{
			Channel:                 enums.Channel_Paynda,
			OriginCardTransactionID: origin.ID,
			AuthorizationID:         origin.AuthorizationID,
			CardID:                  origin.CardID,
			Status:                  status,
			Type:                    req.Type,
			Currency:                origin.Currency,
			TxAmount:                amount,
			TxCurrency:              origin.TxCurrency,
			MerchantName:            origin.MerchantName,
			MerchantCountry:         origin.MerchantCountry,
			MerchantMCC:             origin.MerchantMCC,
			AuthorizationCode:       origin.AuthorizationCode,
			OccurredAt:              time.Now().UTC(),
		}
		if err := u.cardTransactionRepository.Create(txCtx, next); err != nil {
			zap.S().Errorw("create paynda UI transaction step", "error", err)
			return ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return next, nil
}
