package service

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"time"

	"generic-mock/channel/paynda/biz"
	paynda "generic-mock/channel/paynda/enums"
	payndaerrors "generic-mock/channel/paynda/errors"
	"generic-mock/channel/paynda/pkg/idconv"
	"generic-mock/channel/paynda/pkg/timeconv"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/timefmt"
	"generic-mock/pkg/types"
	timeTypes "generic-mock/pkg/types/time"
	sharedbiz "generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/samber/do/v2"
	"github.com/shopspring/decimal"
)

type PayndaOpenAPIService struct {
	usecase     *biz.PayndaOpenAPIUsecase
	cardIssuer  *sharedbiz.CardIssuer
	notificator *biz.PayndaWebhookNotificator
}

func NewPayndaOpenAPIService(injector do.Injector) (*PayndaOpenAPIService, error) {
	return &PayndaOpenAPIService{
		usecase:     do.MustInvoke[*biz.PayndaOpenAPIUsecase](injector),
		cardIssuer:  do.MustInvoke[*sharedbiz.CardIssuer](injector),
		notificator: do.MustInvoke[*biz.PayndaWebhookNotificator](injector),
	}, nil
}

type PayndaCardHolderRequest struct {
	BalanceAccountID    string `uri:"balanceAccountId"`
	FirstName           string `json:"firstName" binding:"required"`
	LastName            string `json:"lastName" binding:"required"`
	MobilePrefix        string `json:"mobilePrefix"`
	Mobile              string `json:"mobile"`
	Email               string `json:"email"`
	BillingAddressLine1 string `json:"billingAddressLine1"`
	BillingAddressLine2 string `json:"billingAddressLine2"` // Invalid: no neutral second address field.
	BillingCity         string `json:"billingCity"`
	BillingCountryCode  string `json:"billingCountryCode"`
	BillingPostalCode   string `json:"billingPostalCode"`
	BillingState        string `json:"billingState"`
	UnlimitedBalance    bool   `json:"unlimitedBalance"` // Invalid: generic mock does not model channel wallet limits.
}
type PayndaCardholderData struct {
	ID                  string             `json:"id"`
	CreateTime          timeTypes.DateTime `json:"createTime"`
	UpdateTime          timeTypes.DateTime `json:"updateTime"`
	BalanceAccountID    string             `json:"balanceAccountId"`
	FirstName           string             `json:"firstName"`
	LastName            string             `json:"lastName"`
	MobilePrefix        string             `json:"mobilePrefix"`
	Mobile              string             `json:"mobile"`
	Email               string             `json:"email"`
	BillingAddressLine1 string             `json:"billingAddressLine1"`
	BillingCity         string             `json:"billingCity"`
	BillingCountryCode  string             `json:"billingCountryCode"`
	BillingPostalCode   string             `json:"billingPostalCode"`
	BillingState        string             `json:"billingState"`
}

func (s *PayndaOpenAPIService) CreateCardHolder(ctx context.Context, req *PayndaCardHolderRequest) (*PayndaCardholderData, error) {
	var accountID model.ID
	var err error
	accountID, err = idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.CreateCardHolder(ctx, &biz.PayndaCreateCardHolderRequest{
		AccountID:              accountID,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		MobilePrefix:           req.MobilePrefix,
		Mobile:                 req.Mobile,
		Email:                  req.Email,
		ResidentialAddress:     req.BillingAddressLine1,
		ResidentialCity:        req.BillingCity,
		ResidentialCountryCode: req.BillingCountryCode,
		ResidentialPostalCode:  req.BillingPostalCode,
		ResidentialState:       req.BillingState,
	})
	if err != nil {
		return nil, err
	}

	return convertCardHolderToPayndaCardholderData(item, req.BalanceAccountID), nil
}

type PayndaCardHolderIDRequest struct {
	BalanceAccountID string `uri:"balanceAccountId"`
	CardholderID     string `uri:"cardholderId"`
}

func (s *PayndaOpenAPIService) GetCardHolder(
	ctx context.Context,
	req *PayndaCardHolderIDRequest,
) (*PayndaCardholderData, error) {
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromString(req.CardholderID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCardHolder(ctx, &biz.PayndaResourceRequest{
		AccountID: accountID,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}

	return convertCardHolderToPayndaCardholderData(item, req.BalanceAccountID), nil
}

type PayndaUpdateCardHolderRequest struct {
	PayndaCardHolderIDRequest
	FirstName           string `json:"firstName" binding:"required"`
	LastName            string `json:"lastName" binding:"required"`
	MobilePrefix        string `json:"mobilePrefix"`
	Mobile              string `json:"mobile"`
	Email               string `json:"email"`
	BillingAddressLine1 string `json:"billingAddressLine1"`
	BillingAddressLine2 string `json:"billingAddressLine2"` // Invalid: no neutral second address field.
	BillingCity         string `json:"billingCity"`
	BillingCountryCode  string `json:"billingCountryCode"`
	BillingPostalCode   string `json:"billingPostalCode"`
	BillingState        string `json:"billingState"`
	UnlimitedBalance    bool   `json:"unlimitedBalance"` // Invalid: generic mock does not model channel wallet limits.
}

func (s *PayndaOpenAPIService) UpdateCardHolder(
	ctx context.Context,
	req *PayndaUpdateCardHolderRequest,
) (*struct{}, error) {
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromString(req.CardholderID)
	if err != nil {
		return nil, err
	}
	_, err = s.usecase.UpdateCardHolder(ctx, &biz.PayndaUpdateCardHolderRequest{
		AccountID:              accountID,
		ID:                     id,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		MobilePrefix:           req.MobilePrefix,
		Mobile:                 req.Mobile,
		Email:                  req.Email,
		ResidentialAddress:     req.BillingAddressLine1,
		ResidentialCity:        req.BillingCity,
		ResidentialCountryCode: req.BillingCountryCode,
		ResidentialPostalCode:  req.BillingPostalCode,
		ResidentialState:       req.BillingState,
	})
	if err != nil {
		return nil, err
	}

	return &struct{}{}, nil
}

type PayndaListRequest struct {
	BalanceAccountID string `uri:"balanceAccountId"`
	Current          *int   `form:"current" binding:"omitempty,min=1"`
	PageSize         *int   `form:"pageSize" binding:"omitempty,min=1"`
}

type PayndaCardholdersData struct {
	Pages   int64                   `json:"pages"`
	Records []*PayndaCardholderData `json:"records"`
	Total   int64                   `json:"total"`
	Size    int64                   `json:"size"`
	Current int64                   `json:"current"`
}

func (s *PayndaOpenAPIService) ListCardHolders(
	ctx context.Context,
	req *PayndaListRequest,
) (*PayndaCardholdersData, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	offset, limit, current := resolvePayndaPagination(types.Value(req.Current), types.Value(req.PageSize))
	items, err := s.usecase.ListCardHolders(ctx, &biz.PayndaListRequest{
		AccountID: &accountID,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	records := make([]*PayndaCardholderData, 0, len(items))
	for _, item := range items {
		records = append(records, convertCardHolderToPayndaCardholderData(item, req.BalanceAccountID))
	}

	return &PayndaCardholdersData{
		Pages:   int64(current),
		Records: records,
		Total:   int64(len(records)),
		Size:    int64(limit),
		Current: int64(current),
	}, nil
}

type PayndaCardBinsRequest struct {
	BalanceAccountID string `uri:"balanceAccountId"` // Invalid: retained for the SDK path; products are channel-level.
	CreditLimitType  string `form:"creditLimitType"` // Invalid: all mock cards use independent generic balances.
}
type PayndaCardBinData struct {
	ID                string            `json:"id"`
	CardBin           string            `json:"cardBin"`
	Name              string            `json:"name"`
	Type              string            `json:"type"`
	CardClass         string            `json:"cardClass"`
	SupportCurrencies []common.Currency `json:"supportCurrencies"`
}

type PayndaCardBinsData struct {
	Records []*PayndaCardBinData `json:"records"`
}

func (s *PayndaOpenAPIService) ListCardBins(ctx context.Context, req *PayndaCardBinsRequest) (*PayndaCardBinsData, error) {
	items, err := s.usecase.ListCardProducts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*PayndaCardBinData, 0, len(items))
	for _, item := range items {
		result = append(result, &PayndaCardBinData{
			ID:        idconv.ToString(item.ID),
			CardBin:   item.Prefix,
			Name:      item.Prefix,
			Type:      "CARD_BIN",
			CardClass: "VIRTUAL",
			SupportCurrencies: []common.Currency{
				common.Currency_USD,
			},
		})
	}
	return &PayndaCardBinsData{
		Records: result,
	}, nil
}

type PayndaCreateCardRequest struct {
	BalanceAccountID           string          `uri:"balanceAccountId"`
	RequestID                  string          `form:"requestId" binding:"required"`
	CardholderID               string          `json:"cardholderId" binding:"required"`
	Currency                   common.Currency `json:"currency" binding:"required"`
	Amount                     string          `json:"amount" binding:"required"`
	ExpirationDate             string          `json:"expirationDate"`
	CardBinID                  string          `json:"cardBinId" binding:"required"`
	SingleUse                  bool            `json:"singleUse"`                    // Invalid: generic model has no one-time Paynda card flag.
	TransactionCountLimitTotal int64           `json:"transactionCountLimitToTotal"` // Invalid: generic model has no channel transaction limit.
}

func (s *PayndaOpenAPIService) CreateCard(ctx context.Context, req *PayndaCreateCardRequest) (*PayndaCardDetail, error) {
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	cardholderID, err := idconv.FromString(req.CardholderID)
	if err != nil {
		return nil, err
	}
	cardProductID, err := idconv.FromString(req.CardBinID)
	if err != nil {
		return nil, err
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || !amount.IsPositive() {
		return nil, payndaerrors.ErrInvalidOperation
	}
	expireAt, valid := timeconv.ParseCardExpiration(
		req.ExpirationDate,
		time.Now().UTC().AddDate(2, 0, 0),
	)
	if !valid {
		return nil, payndaerrors.ErrInvalidOperation
	}
	result, err := s.cardIssuer.Issue(ctx, &sharedbiz.IssueCardReq{
		IssueCardHolder: &sharedbiz.IssueCardHolder{
			ID: &cardholderID,
		},
		Channel:          common.Channel_Paynda,
		AccountID:        accountID,
		CardType:         common.CardType_Single,
		CardProductID:    cardProductID,
		Currency:         req.Currency,
		CardScheme:       common.CardScheme_MasterCard,
		FormType:         common.CardFormType_Virtual,
		Status:           common.CardStatus_Active,
		ExpireAt:         expireAt,
		RequestID:        &req.RequestID,
		InitialAvailable: &amount,
		Notificator:      s.notificator,
	})
	if err != nil {
		return nil, convertPayndaIssueCardError(err)
	}
	return convertCardToPayndaCardDetail(result.Card, req.BalanceAccountID), nil
}

func convertPayndaIssueCardError(err error) error {
	switch {
	case stderrors.Is(err, sharederrors.ErrAccountNotFound),
		stderrors.Is(err, sharederrors.ErrCardHolderNotFound),
		stderrors.Is(err, sharederrors.ErrCardProductNotFound):
		return payndaerrors.ErrResourceNotFound
	case stderrors.Is(err, sharederrors.ErrDatabaseOperation):
		return payndaerrors.ErrDatabaseOperation
	default:
		return payndaerrors.ErrInvalidOperation
	}
}

type PayndaCardRequest struct {
	BalanceAccountID string `uri:"balanceAccountId"`
	CardID           string `uri:"cardId"`
}

func (s *PayndaOpenAPIService) GetCard(ctx context.Context, req *PayndaCardRequest) (*PayndaCardData, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCard(ctx, &biz.PayndaResourceRequest{
		AccountID: accountID,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}

	return convertCardToPayndaCardData(item, req.BalanceAccountID), nil
}

type PayndaCardsData struct {
	Pages   int64             `json:"pages"`
	Records []*PayndaCardData `json:"records"`
	Total   int64             `json:"total"`
	Size    int64             `json:"size"`
	Current int64             `json:"current"`
}

func (s *PayndaOpenAPIService) ListCards(ctx context.Context, req *PayndaListRequest) (*PayndaCardsData, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	offset, limit, current := resolvePayndaPagination(types.Value(req.Current), types.Value(req.PageSize))
	items, err := s.usecase.ListCards(ctx, &biz.PayndaListRequest{
		AccountID: &accountID,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	records := make([]*PayndaCardData, 0, len(items))
	for _, item := range items {
		records = append(records, convertCardToPayndaCardData(item, req.BalanceAccountID))
	}

	return &PayndaCardsData{
		Pages:   int64(current),
		Records: records,
		Total:   int64(len(records)),
		Size:    int64(limit),
		Current: int64(current),
	}, nil
}

func (s *PayndaOpenAPIService) GetCardSensitive(ctx context.Context, req *PayndaCardRequest) (*PayndaCardSensitiveData, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCard(ctx, &biz.PayndaResourceRequest{
		AccountID: accountID,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}

	return convertCardToPayndaCardSensitiveData(item), nil
}

func (s *PayndaOpenAPIService) GetCardBalance(
	ctx context.Context,
	req *PayndaCardRequest,
) (*PayndaCardBalanceData, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	wallet, err := s.usecase.GetCardBalance(ctx, &biz.PayndaResourceRequest{
		AccountID: accountID,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}

	return convertWalletToPayndaCardBalanceData(wallet), nil
}

type PayndaBalanceAccountWalletData struct {
	ID               string             `json:"id"`
	CreateTime       timeTypes.DateTime `json:"createTime"`
	UpdateTime       timeTypes.DateTime `json:"updateTime"`
	BalanceAccountID string             `json:"balanceAccountId"`
	Currency         common.Currency    `json:"currency"`
	Amount           string             `json:"amount"`
	FrozenAmount     string             `json:"frozenAmount"`
}

func (s *PayndaOpenAPIService) ListBalanceAccountWallets(
	ctx context.Context,
	req *PayndaListRequest,
) (*[]*PayndaBalanceAccountWalletData, error) {
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	accountWallet, err := s.usecase.GetAccountWallet(ctx, accountID)
	if err != nil {
		return nil, err
	}

	result := []*PayndaBalanceAccountWalletData{
		convertAccountWalletToPayndaBalanceAccountWalletData(accountWallet, req.BalanceAccountID),
	}

	return &result, nil
}

type PayndaMerchantWalletData struct {
	ID         string             `json:"id"`
	CreateTime timeTypes.DateTime `json:"createTime"`
	UpdateTime timeTypes.DateTime `json:"updateTime"`
	Name       string             `json:"name"`
	Currency   common.Currency    `json:"currency"`
	Amount     string             `json:"amount"`
}

type PayndaMerchantWalletsRequest struct {
	AppID    string `header:"appId" binding:"required"`
	Current  *int   `form:"current" binding:"omitempty,min=1"`
	PageSize *int   `form:"pageSize" binding:"omitempty,min=1"`
}

func (s *PayndaOpenAPIService) ListMerchantWallets(
	ctx context.Context,
	req *PayndaMerchantWalletsRequest,
) (*[]*PayndaMerchantWalletData, error) {
	accountID, err := idconv.FromAccountString(req.AppID)
	if err != nil {
		return nil, err
	}
	account, err := s.usecase.GetAccountWallet(ctx, accountID)
	if err != nil {
		return nil, err
	}
	items := []*PayndaMerchantWalletData{{
		ID:         idconv.ToString(account.Wallet.ID),
		Name:       account.Account.Name,
		CreateTime: timeTypes.DateTime(account.Wallet.CreatedAt.UTC()),
		UpdateTime: timeTypes.DateTime(account.Wallet.UpdatedAt.UTC()),
		Currency:   account.Wallet.Currency,
		Amount:     account.Wallet.Available.String(),
	}}
	return &items, nil
}

type PayndaBalanceAccountWalletTransferRequest struct {
	RequestID        string              `form:"requestId" binding:"required"`
	BalanceAccountID string              `json:"balanceAccountId"` // Invalid: the path-independent mock account is resolved internally.
	Type             paynda.TransferType `json:"type" binding:"required"`
	Currency         common.Currency     `json:"currency" binding:"required"`
	Amount           string              `json:"amount" binding:"required"`
}

type PayndaBalanceAccountWalletTransferData struct {
	ID               string              `json:"id"`
	CreateTime       timeTypes.DateTime  `json:"createTime"`
	UpdateTime       timeTypes.DateTime  `json:"updateTime"`
	BalanceAccountID string              `json:"balanceAccountId"`
	Type             paynda.TransferType `json:"type"`
	Currency         common.Currency     `json:"currency"`
	Amount           string              `json:"amount"`
}

func (s *PayndaOpenAPIService) TransferBalanceAccountWallet(
	ctx context.Context,
	req *PayndaBalanceAccountWalletTransferRequest,
) (*PayndaBalanceAccountWalletTransferData, error) {
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	if req.Type != paynda.TransferType_In || req.Currency != common.Currency_USD {
		return nil, payndaerrors.ErrInvalidOperation
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || !amount.IsPositive() {
		return nil, payndaerrors.ErrInvalidOperation
	}
	accountWallet, err := s.usecase.TransferAccountWallet(ctx, &biz.PayndaAccountWalletTransferRequest{
		AccountID: accountID,
		Amount:    amount,
	})
	if err != nil {
		return nil, err
	}

	return &PayndaBalanceAccountWalletTransferData{
		ID:               idconv.ToString(accountWallet.Wallet.ID),
		CreateTime:       timeTypes.DateTime(accountWallet.Wallet.CreatedAt.UTC()),
		UpdateTime:       timeTypes.DateTime(accountWallet.Wallet.UpdatedAt.UTC()),
		BalanceAccountID: idconv.ToString(accountWallet.Account.ID),
		Type:             req.Type,
		Currency:         req.Currency,
		Amount:           amount.String(),
	}, nil
}

type PayndaRequestResultRequest struct {
	AppID     string `header:"appId" binding:"required"`
	RequestID string `form:"requestId" binding:"required"`
}
type PayndaRequestResultData struct {
	ID         string             `json:"id"`
	CreateTime timeTypes.DateTime `json:"createTime"`
	UpdateTime timeTypes.DateTime `json:"updateTime"`
	RequestID  string             `json:"requestId"`
	Result     string             `json:"result"`
}

func (s *PayndaOpenAPIService) RequestResult(
	ctx context.Context,
	req *PayndaRequestResultRequest,
) (*PayndaRequestResultData, error) {
	accountID, err := idconv.FromAccountString(req.AppID)
	if err != nil {
		return nil, err
	}
	result, err := s.usecase.FindRequestResult(ctx, &biz.PayndaRequestLookup{
		AccountID: accountID,
		RequestID: req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	var payload any
	var base model.BaseModel
	if result.Transaction != nil {
		payload = convertCardTransactionToPayndaCardBalanceTransferData(result.Transaction)
		base = result.Transaction.BaseModel
	} else {
		base = result.Card.BaseModel
		if result.IsCardCreate {
			payload = convertCardToPayndaCardDetail(result.Card, idconv.ToString(accountID))
		} else {
			payload = struct {
				ID        string `json:"id"`
				RequestID string `json:"requestId"`
			}{
				ID:        idconv.ToString(result.Card.ID),
				RequestID: req.RequestID,
			}
		}
	}
	encoded, err := json.Marshal(struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    any    `json:"data"`
	}{
		Code:    200,
		Message: "success",
		Success: true,
		Data:    payload,
	})
	if err != nil {
		return nil, payndaerrors.ErrInvalidOperation
	}
	return &PayndaRequestResultData{
		ID:         idconv.ToString(base.ID),
		CreateTime: timeTypes.DateTime(base.CreatedAt.UTC()),
		UpdateTime: timeTypes.DateTime(base.UpdatedAt.UTC()),
		RequestID:  req.RequestID,
		Result:     string(encoded),
	}, nil
}

type PayndaCardStatusRequest struct {
	BalanceAccountID string `uri:"balanceAccountId"`
	CardID           string `uri:"cardId"`
	RequestID        string `form:"requestId"`
}

func (s *PayndaOpenAPIService) FreezeCard(ctx context.Context, req *PayndaCardStatusRequest) (*struct{}, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.UpdateCardStatus(ctx, &biz.PayndaUpdateCardStatusRequest{
		AccountID: accountID,
		CardID:    cardID,
		RequestID: req.RequestID,
		Status:    paynda.CardStatus_Frozen,
	})
	if err != nil {
		return nil, err
	}
	_ = s.notificator.NotifyCardStatus(ctx, &sharedbiz.NotifyCardStatusReq{
		AccountID: card.AccountID,
		Channel:   common.Channel_Paynda,
		CardID:    card.ID,
		Status:    card.Status,
	})
	return &struct{}{}, nil
}

func (s *PayndaOpenAPIService) UnfreezeCard(ctx context.Context, req *PayndaCardStatusRequest) (*struct{}, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.UpdateCardStatus(ctx, &biz.PayndaUpdateCardStatusRequest{
		AccountID: accountID,
		CardID:    cardID,
		RequestID: req.RequestID,
		Status:    paynda.CardStatus_Active,
	})
	if err != nil {
		return nil, err
	}
	_ = s.notificator.NotifyCardStatus(ctx, &sharedbiz.NotifyCardStatusReq{
		AccountID: card.AccountID,
		Channel:   common.Channel_Paynda,
		CardID:    card.ID,
		Status:    card.Status,
	})
	return &struct{}{}, nil
}

func (s *PayndaOpenAPIService) ReleaseCard(ctx context.Context, req *PayndaCardStatusRequest) (*struct{}, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.ReleaseCard(ctx, &biz.PayndaUpdateCardStatusRequest{
		AccountID: accountID,
		CardID:    cardID,
		RequestID: req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	_ = s.notificator.NotifyCardStatus(ctx, &sharedbiz.NotifyCardStatusReq{
		AccountID: card.AccountID,
		Channel:   common.Channel_Paynda,
		CardID:    card.ID,
		Status:    card.Status,
	})

	return &struct{}{}, nil
}

type PayndaCardBalanceTransferRequest struct {
	BalanceAccountID string              `uri:"balanceAccountId"`
	RequestID        string              `form:"requestId" binding:"required"`
	CardID           string              `json:"cardId" binding:"required"`
	Amount           string              `json:"amount" binding:"required"`
	Type             paynda.TransferType `json:"type" binding:"required"`
}

type PayndaCardBalanceTransferData struct {
	CardID    string              `json:"cardId"`
	NewAmount string              `json:"newAmount"`
	OldAmount string              `json:"oldAmount"`
	Amount    string              `json:"amount"`
	Type      paynda.TransferType `json:"type"`
}

func (s *PayndaOpenAPIService) TransferCardBalance(
	ctx context.Context,
	req *PayndaCardBalanceTransferRequest,
) (*PayndaCardBalanceTransferData, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || !amount.IsPositive() {
		return nil, payndaerrors.ErrInvalidOperation
	}
	if req.Type != paynda.TransferType_In && req.Type != paynda.TransferType_Out {
		return nil, payndaerrors.ErrInvalidOperation
	}
	transaction, err := s.usecase.TransferCardBalance(ctx, &biz.PayndaTransferRequest{
		AccountID: accountID,
		CardID:    cardID,
		RequestID: req.RequestID,
		Amount:    amount,
		Type:      req.Type,
	})
	if err != nil {
		return nil, err
	}

	return convertCardTransactionToPayndaCardBalanceTransferData(transaction), nil
}

type PayndaCardBalanceUpdateHistoryData struct {
	ID         string             `json:"id"`
	CreateTime timeTypes.DateTime `json:"createTime"`
	UpdateTime timeTypes.DateTime `json:"updateTime"`
	CardID     string             `json:"cardId"`
	NewAmount  string             `json:"newAmount"`
	OldAmount  string             `json:"oldAmount"`
}

func (s *PayndaOpenAPIService) ListCardBalanceUpdates(
	ctx context.Context,
	req *PayndaListRequest,
) (*[]*PayndaCardBalanceUpdateHistoryData, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	offset, limit, _ := resolvePayndaPagination(types.Value(req.Current), types.Value(req.PageSize))
	items, err := s.usecase.ListCardBalanceUpdates(ctx, &biz.PayndaListRequest{
		AccountID: &accountID,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	records := make([]*PayndaCardBalanceUpdateHistoryData, 0, len(items))
	for _, item := range items {
		oldAmount, err := decimal.NewFromString(string(item.RawPayload))
		if err != nil {
			return nil, payndaerrors.ErrInvalidOperation
		}
		newAmount := oldAmount
		if item.Type == common.CardTransactionType_FundIn {
			newAmount = newAmount.Add(item.TxAmount)
		} else {
			newAmount = newAmount.Sub(item.TxAmount)
		}
		records = append(records, &PayndaCardBalanceUpdateHistoryData{
			ID:         idconv.ToString(item.ID),
			CreateTime: timeTypes.DateTime(item.CreatedAt.UTC()),
			UpdateTime: timeTypes.DateTime(item.UpdatedAt.UTC()),
			CardID:     idconv.ToString(item.CardID),
			NewAmount:  newAmount.String(),
			OldAmount:  oldAmount.String(),
		})
	}

	return &records, nil
}

type PayndaTransactionsRequest struct {
	BalanceAccountID     string              `uri:"balanceAccountId"`
	Current              *int                `form:"current" binding:"omitempty,min=1"`
	PageSize             *int                `form:"pageSize" binding:"omitempty,min=1"`
	CardholderID         string              `form:"cardholderId"` // Invalid: generic transactions do not repeat the cardholder relation.
	CardID               *string             `form:"cardId"`
	TransactionTimeStart *timeTypes.DateTime `form:"transactionTimeStart"`
	TransactionTimeEnd   *timeTypes.DateTime `form:"transactionTimeEnd"`
}

type PayndaTransactionData struct {
	ID                                  string                 `json:"id"`
	CreateTime                          timeTypes.DateTime     `json:"createTime"`
	UpdateTime                          timeTypes.DateTime     `json:"updateTime"`
	CardID                              string                 `json:"cardId"`
	MaskCardNo                          string                 `json:"maskCardNo"`
	Type                                paynda.TransactionType `json:"type"`
	ApprovalCode                        string                 `json:"approvalCode"`
	PreAuthAmount                       string                 `json:"preAuthAmount"`
	PostedAmount                        string                 `json:"postedAmount"`
	Currency                            common.Currency        `json:"currency"`
	OriginalCurrencyCode                common.Currency        `json:"originalCurrencyCode"`
	TransactionAmountInOriginalCurrency string                 `json:"transactionAmountInOriginalCurrency"`
	TransactionTime                     timeTypes.DateTime     `json:"transactionTime"`
	AuthorizationTime                   *timeTypes.DateTime    `json:"authorizationTime,omitempty"`
	MerchantMcc                         string                 `json:"merchantMcc"`
	MerchantName                        string                 `json:"merchantName"`
	SupplierTransactionID               string                 `json:"supplierTransactionId"`
	SupplierTransactionLinkID           string                 `json:"supplierTransactionLinkId"`
	DeclineMessage                      string                 `json:"declineMessage"`
}

type PayndaTransactionsData struct {
	Pages   int64                    `json:"pages"`
	Records []*PayndaTransactionData `json:"records"`
	Total   int64                    `json:"total"`
	Size    int64                    `json:"size"`
	Current int64                    `json:"current"`
}

func (s *PayndaOpenAPIService) ListCardTransactions(
	ctx context.Context,
	req *PayndaTransactionsRequest,
) (*PayndaTransactionsData, error) {
	var err error
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	offset, limit, current := resolvePayndaPagination(types.Value(req.Current), types.Value(req.PageSize))
	start := (*time.Time)(req.TransactionTimeStart)
	end := (*time.Time)(req.TransactionTimeEnd)
	if start != nil && end != nil && start.After(*end) {
		return nil, payndaerrors.ErrInvalidOperation
	}
	cardID, err := idconv.FromOptionalString(req.CardID)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListCardTransactions(ctx, &biz.PayndaListTransactionsRequest{
		PayndaListRequest: biz.PayndaListRequest{
			AccountID: &accountID,
			Offset:    offset,
			Limit:     limit,
		},
		CardID:         cardID,
		StartCreatedAt: start,
		EndCreatedAt:   end,
	})
	if err != nil {
		return nil, err
	}

	records := make([]*PayndaTransactionData, 0, len(items))
	for _, item := range items {
		records = append(records, convertCardTransactionDetailToPayndaTransactionData(item))
	}

	return &PayndaTransactionsData{
		Pages:   int64(current),
		Records: records,
		Total:   int64(len(records)),
		Size:    int64(limit),
		Current: int64(current),
	}, nil
}

type PayndaCardTransactionRequest struct {
	BalanceAccountID string `uri:"balanceAccountId"`
	TransactionID    string `uri:"transactionId"`
}

func (s *PayndaOpenAPIService) GetCardTransaction(
	ctx context.Context,
	req *PayndaCardTransactionRequest,
) (*PayndaTransactionData, error) {
	accountID, err := idconv.FromAccountString(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromString(req.TransactionID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCardTransaction(ctx, &biz.PayndaResourceRequest{
		AccountID: accountID,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}

	return convertCardTransactionDetailToPayndaTransactionData(item), nil
}

type PayndaEmbeddedResponse struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Success bool   `json:"success"`
}
type PayndaCardDetail struct {
	Card          *PayndaCardData          `json:"card"`
	CardSensitive *PayndaCardSensitiveData `json:"sensitiveInfo"`
	CardBalance   *PayndaCardBalanceData   `json:"balance"`
}
type PayndaCardData struct {
	ID               string             `json:"id"`
	CreateTime       timeTypes.DateTime `json:"createTime"`
	UpdateTime       timeTypes.DateTime `json:"updateTime"`
	BalanceAccountID string             `json:"balanceAccountId"`
	CardholderID     string             `json:"cardholderId"`
	Status           paynda.CardStatus  `json:"status"`
	MaskCardNo       string             `json:"maskCardNo"`
	CardScheme       common.CardScheme  `json:"cardScheme"`
	CardBin          string             `json:"cardBin"`
	Currency         common.Currency    `json:"currency"`
	CreditLimitType  string             `json:"creditLimitType"`
	AmountUsed       string             `json:"amountUsed"`
	Amount           string             `json:"amount"`
}
type PayndaCardSensitiveData struct {
	ID             string             `json:"id"`
	CreateTime     timeTypes.DateTime `json:"createTime"`
	UpdateTime     timeTypes.DateTime `json:"updateTime"`
	CardID         string             `json:"cardId"`
	CVV            string             `json:"cvv"`
	ExpirationDate string             `json:"expirationDate"`
	CardNo         string             `json:"cardNo"`
}
type PayndaCardBalanceData struct {
	ID              string             `json:"id"`
	CreateTime      timeTypes.DateTime `json:"createTime"`
	UpdateTime      timeTypes.DateTime `json:"updateTime"`
	AmountUsed      string             `json:"amountUsed"`
	AmountFrozen    string             `json:"amountFrozen"`
	AvailableAmount string             `json:"availableAmount"`
	Amount          string             `json:"amount"`
}

func convertCardHolderToPayndaCardholderData(item *model.CardHolder, balanceAccountID string) *PayndaCardholderData {
	return &PayndaCardholderData{
		ID:                  idconv.ToString(item.ID),
		CreateTime:          timeTypes.DateTime(item.CreatedAt.UTC()),
		UpdateTime:          timeTypes.DateTime(item.UpdatedAt.UTC()),
		BalanceAccountID:    balanceAccountID,
		FirstName:           item.FirstName,
		LastName:            item.LastName,
		MobilePrefix:        item.MobilePrefix,
		Mobile:              item.Mobile,
		Email:               item.Email,
		BillingAddressLine1: item.ResidentialAddress,
		BillingCity:         item.ResidentialCity,
		BillingCountryCode:  item.ResidentialCountryCode,
		BillingPostalCode:   item.ResidentialPostalCode,
		BillingState:        item.ResidentialState,
	}
}

func convertCardToPayndaCardDetail(item *model.Card, balanceAccountID string) *PayndaCardDetail {
	wallet := item.Wallet
	cardBalance := &PayndaCardBalanceData{
		ID:              idconv.ToString(item.ID),
		CreateTime:      timeTypes.DateTime(item.CreatedAt.UTC()),
		UpdateTime:      timeTypes.DateTime(item.UpdatedAt.UTC()),
		AvailableAmount: decimal.Zero.String(),
		Amount:          decimal.Zero.String(),
		AmountUsed:      decimal.Zero.String(),
		AmountFrozen:    decimal.Zero.String(),
	}
	if wallet != nil {
		cardBalance = convertWalletToPayndaCardBalanceData(wallet)
	}
	return &PayndaCardDetail{
		Card:          convertCardToPayndaCardData(item, balanceAccountID),
		CardSensitive: convertCardToPayndaCardSensitiveData(item),
		CardBalance:   cardBalance,
	}
}

func convertCardToPayndaCardData(item *model.Card, balanceAccountID string) *PayndaCardData {
	amount := decimal.Zero.String()
	if item.Wallet != nil {
		amount = item.Wallet.Available.String()
	}
	return &PayndaCardData{
		ID:               idconv.ToString(item.ID),
		CreateTime:       timeTypes.DateTime(item.CreatedAt.UTC()),
		UpdateTime:       timeTypes.DateTime(item.UpdatedAt.UTC()),
		BalanceAccountID: balanceAccountID,
		CardholderID:     idconv.ToString(item.CardHolderID),
		Status:           paynda.ConvertGenericCardStatusToCardStatus(item.Status),
		MaskCardNo:       "******" + item.CardNumber[len(item.CardNumber)-4:],
		CardScheme:       item.CardScheme,
		CardBin:          item.CardBin,
		Currency:         item.CardCurrency,
		CreditLimitType:  "INDEPENDENT",
		AmountUsed:       decimal.Zero.String(),
		Amount:           amount,
	}
}

func convertCardToPayndaCardSensitiveData(item *model.Card) *PayndaCardSensitiveData {
	return &PayndaCardSensitiveData{
		ID:             idconv.ToString(item.ID),
		CreateTime:     timeTypes.DateTime(item.CreatedAt.UTC()),
		UpdateTime:     timeTypes.DateTime(item.UpdatedAt.UTC()),
		CardID:         idconv.ToString(item.ID),
		CVV:            item.Cvv,
		ExpirationDate: timefmt.CardExpiration(item.ExpireAt),
		CardNo:         item.CardNumber,
	}
}

func convertWalletToPayndaCardBalanceData(item *model.Wallet) *PayndaCardBalanceData {
	return &PayndaCardBalanceData{
		ID:              idconv.ToString(item.ID),
		CreateTime:      timeTypes.DateTime(item.CreatedAt.UTC()),
		UpdateTime:      timeTypes.DateTime(item.UpdatedAt.UTC()),
		AmountUsed:      decimal.Zero.String(),
		AmountFrozen:    item.PendingOut.String(),
		AvailableAmount: item.Available.String(),
		Amount:          item.Available.String(),
	}
}

func convertAccountWalletToPayndaBalanceAccountWalletData(
	item *biz.PayndaAccountWallet,
	balanceAccountID string,
) *PayndaBalanceAccountWalletData {
	return &PayndaBalanceAccountWalletData{
		ID:               idconv.ToString(item.Wallet.ID),
		CreateTime:       timeTypes.DateTime(item.Wallet.CreatedAt.UTC()),
		UpdateTime:       timeTypes.DateTime(item.Wallet.UpdatedAt.UTC()),
		BalanceAccountID: balanceAccountID,
		Currency:         item.Wallet.Currency,
		Amount:           item.Wallet.Available.String(),
		FrozenAmount:     item.Wallet.PendingOut.String(),
	}
}

func convertCardTransactionToPayndaCardBalanceTransferData(item *model.CardTransaction) *PayndaCardBalanceTransferData {
	transferType := paynda.TransferType_In
	if item.Type == common.CardTransactionType_FundOut {
		transferType = paynda.TransferType_Out
	}
	oldAmount, err := decimal.NewFromString(string(item.RawPayload))
	if err != nil {
		oldAmount = decimal.Zero
	}
	newAmount := oldAmount.Add(item.TxAmount)
	if transferType == paynda.TransferType_Out {
		newAmount = oldAmount.Sub(item.TxAmount)
	}

	return &PayndaCardBalanceTransferData{
		CardID:    idconv.ToString(item.CardID),
		NewAmount: newAmount.String(),
		OldAmount: oldAmount.String(),
		Amount:    item.TxAmount.String(),
		Type:      transferType,
	}
}

func convertCardTransactionDetailToPayndaTransactionData(item *biz.PayndaCardTransactionDetail) *PayndaTransactionData {
	transactionTime := timeTypes.DateTime(item.Transaction.CreatedAt)
	var authorizationTime *timeTypes.DateTime
	if item.Authorization != nil {
		value := timeTypes.DateTime(item.Authorization.CreatedAt)
		authorizationTime = &value
	}
	transactionID := idconv.ToString(item.Transaction.ID)
	supplierTransactionLinkID := ""
	if item.Transaction.Type != common.CardTransactionType_AUTH && item.AuthorizationTransaction != nil {
		supplierTransactionLinkID = idconv.ToString(item.AuthorizationTransaction.ID)
	}
	return &PayndaTransactionData{
		ID:                                  transactionID,
		CreateTime:                          timeTypes.DateTime(item.Transaction.CreatedAt.UTC()),
		UpdateTime:                          timeTypes.DateTime(item.Transaction.UpdatedAt.UTC()),
		CardID:                              idconv.ToString(item.Transaction.CardID),
		Type:                                paynda.ConvertGenericTransactionTypeToTransactionType(item.Transaction.Type),
		ApprovalCode:                        item.Transaction.AuthorizationCode,
		PreAuthAmount:                       item.Transaction.TxAmount.String(),
		PostedAmount:                        item.Transaction.TxAmount.String(),
		Currency:                            item.Transaction.Currency,
		OriginalCurrencyCode:                item.Transaction.TxCurrency,
		TransactionAmountInOriginalCurrency: item.Transaction.TxAmount.String(),
		TransactionTime:                     transactionTime,
		AuthorizationTime:                   authorizationTime,
		MerchantMcc:                         item.Transaction.MerchantMCC,
		MerchantName:                        item.Transaction.MerchantName,
		SupplierTransactionID:               transactionID,
		SupplierTransactionLinkID:           supplierTransactionLinkID,
	}
}

func resolvePayndaPagination(current int, pageSize int) (int, int, int) {
	if current < 1 {
		current = 1
	}
	if pageSize < 1 {
		pageSize = 100
	}

	return (current - 1) * pageSize, pageSize, current
}
