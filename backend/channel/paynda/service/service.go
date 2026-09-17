package service

import (
	"context"
	"encoding/json"
	"time"

	"generic-mock/channel/paynda/biz"
	paynda "generic-mock/channel/paynda/enums"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
)

type PayndaOpenAPIService struct {
	usecase *biz.PayndaOpenAPIUsecase
}

func NewPayndaOpenAPIService(injector *do.Injector) (*PayndaOpenAPIService, error) {
	return &PayndaOpenAPIService{
		usecase: do.MustInvoke[*biz.PayndaOpenAPIUsecase](injector),
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
	ID                  string    `json:"id"`
	CreateTime          time.Time `json:"createTime"`
	UpdateTime          time.Time `json:"updateTime"`
	BalanceAccountID    string    `json:"balanceAccountId"`
	FirstName           string    `json:"firstName"`
	LastName            string    `json:"lastName"`
	MobilePrefix        string    `json:"mobilePrefix"`
	Mobile              string    `json:"mobile"`
	Email               string    `json:"email"`
	BillingAddressLine1 string    `json:"billingAddressLine1"`
	BillingCity         string    `json:"billingCity"`
	BillingCountryCode  string    `json:"billingCountryCode"`
	BillingPostalCode   string    `json:"billingPostalCode"`
	BillingState        string    `json:"billingState"`
}

func (s *PayndaOpenAPIService) CreateCardHolder(ctx context.Context, req *PayndaCardHolderRequest) (*PayndaCardholderData, error) {
	item, err := s.usecase.CreateCardHolder(ctx, &biz.PayndaCreateCardHolderRequest{
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

	return payndaCardholderData(item, req.BalanceAccountID), nil
}

type PayndaCardHolderIDRequest struct {
	BalanceAccountID string `uri:"balanceAccountId"`
	CardholderID     string `uri:"cardholderId"`
}

func (s *PayndaOpenAPIService) GetCardHolder(
	ctx context.Context,
	req *PayndaCardHolderIDRequest,
) (*PayndaCardholderData, error) {
	id, err := payndaID(req.CardholderID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCardHolder(ctx, id)
	if err != nil {
		return nil, err
	}

	return payndaCardholderData(item, req.BalanceAccountID), nil
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
	id, err := payndaID(req.CardholderID)
	if err != nil {
		return nil, err
	}
	_, err = s.usecase.UpdateCardHolder(ctx, &biz.PayndaUpdateCardHolderRequest{
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
	Current          int    `form:"current"`
	PageSize         int    `form:"pageSize"`
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
	offset, limit, current := payndaPagination(req.Current, req.PageSize)
	items, err := s.usecase.ListCardHolders(ctx, &biz.PayndaListRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	records := make([]*PayndaCardholderData, 0, len(items))
	for _, item := range items {
		records = append(records, payndaCardholderData(item, req.BalanceAccountID))
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
	BalanceAccountID string `uri:"balanceAccountId"`
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

func (s *PayndaOpenAPIService) ListCardBins(ctx context.Context, _ *PayndaCardBinsRequest) (*PayndaCardBinsData, error) {
	items, err := s.usecase.ListCardProducts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*PayndaCardBinData, 0, len(items))
	for _, item := range items {
		result = append(result, &PayndaCardBinData{
			ID:        payndaIDString(item.ID),
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
	RequestID                  string          `header:"requestId" binding:"required"`
	CardholderID               string          `json:"cardholderId" binding:"required"`
	Currency                   common.Currency `json:"currency" binding:"required"`
	Amount                     string          `json:"amount"` // Invalid: generic card does not persist channel credit limits.
	ExpirationDate             string          `json:"expirationDate"`
	CardBinID                  string          `json:"cardBinId" binding:"required"`
	SingleUse                  bool            `json:"singleUse"`                    // Invalid: generic model has no one-time Paynda card flag.
	TransactionCountLimitTotal int64           `json:"transactionCountLimitToTotal"` // Invalid: generic model has no channel transaction limit.
}

func (s *PayndaOpenAPIService) CreateCard(ctx context.Context, req *PayndaCreateCardRequest) (*PayndaCardDetail, error) {
	cardholderID, err := payndaID(req.CardholderID)
	if err != nil {
		return nil, err
	}
	cardProductID, err := payndaID(req.CardBinID)
	if err != nil {
		return nil, err
	}
	expireAt := time.Now().UTC().AddDate(2, 0, 0)
	if req.ExpirationDate != "" {
		value, err := time.Parse("01/06", req.ExpirationDate)
		if err != nil {
			return nil, biz.ErrInvalidOperation
		}
		expireAt = value.UTC()
	}
	item, err := s.usecase.CreateCard(ctx, &biz.PayndaCreateCardRequest{
		CardHolderID:  cardholderID,
		CardProductID: cardProductID,
		Currency:      req.Currency,
		ExpireAt:      expireAt,
		RequestID:     req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	return payndaCardDetail(item, req.BalanceAccountID), nil
}

type PayndaCardRequest struct {
	BalanceAccountID string `uri:"balanceAccountId"`
	CardID           string `uri:"cardId"`
}

func (s *PayndaOpenAPIService) GetCard(ctx context.Context, req *PayndaCardRequest) (*PayndaCardData, error) {
	id, err := payndaID(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCard(ctx, id)
	if err != nil {
		return nil, err
	}

	return payndaCardData(item, req.BalanceAccountID), nil
}

type PayndaCardsData struct {
	Pages   int64             `json:"pages"`
	Records []*PayndaCardData `json:"records"`
	Total   int64             `json:"total"`
	Size    int64             `json:"size"`
	Current int64             `json:"current"`
}

func (s *PayndaOpenAPIService) ListCards(ctx context.Context, req *PayndaListRequest) (*PayndaCardsData, error) {
	offset, limit, current := payndaPagination(req.Current, req.PageSize)
	items, err := s.usecase.ListCards(ctx, &biz.PayndaListRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	records := make([]*PayndaCardData, 0, len(items))
	for _, item := range items {
		records = append(records, payndaCardData(item, req.BalanceAccountID))
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
	id, err := payndaID(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCard(ctx, id)
	if err != nil {
		return nil, err
	}

	return payndaCardSensitiveData(item), nil
}

func (s *PayndaOpenAPIService) GetCardBalance(
	ctx context.Context,
	req *PayndaCardRequest,
) (*PayndaCardBalanceData, error) {
	id, err := payndaID(req.CardID)
	if err != nil {
		return nil, err
	}
	wallet, err := s.usecase.GetCardBalance(ctx, id)
	if err != nil {
		return nil, err
	}

	return payndaCardBalanceData(wallet), nil
}

type PayndaBalanceAccountWalletData struct {
	ID               string          `json:"id"`
	CreateTime       time.Time       `json:"createTime"`
	UpdateTime       time.Time       `json:"updateTime"`
	BalanceAccountID string          `json:"balanceAccountId"`
	Currency         common.Currency `json:"currency"`
	Amount           string          `json:"amount"`
	FrozenAmount     string          `json:"frozenAmount"`
}

func (s *PayndaOpenAPIService) ListBalanceAccountWallets(
	ctx context.Context,
	req *PayndaListRequest,
) (*[]*PayndaBalanceAccountWalletData, error) {
	accountWallet, err := s.usecase.GetAccountWallet(ctx)
	if err != nil {
		return nil, err
	}

	result := []*PayndaBalanceAccountWalletData{
		payndaBalanceAccountWalletData(accountWallet, req.BalanceAccountID),
	}

	return &result, nil
}

type PayndaMerchantWalletData struct {
	ID         string          `json:"id"`
	CreateTime time.Time       `json:"createTime"`
	UpdateTime time.Time       `json:"updateTime"`
	Name       string          `json:"name"`
	Currency   common.Currency `json:"currency"`
	Amount     string          `json:"amount"`
}

type PayndaMerchantWalletsRequest struct {
	Current  int `form:"current"`
	PageSize int `form:"pageSize"`
}

func (s *PayndaOpenAPIService) ListMerchantWallets(
	ctx context.Context,
	_ *PayndaMerchantWalletsRequest,
) (*[]*PayndaMerchantWalletData, error) {
	accountWallet, err := s.usecase.GetAccountWallet(ctx)
	if err != nil {
		return nil, err
	}

	result := []*PayndaMerchantWalletData{
		{
			ID:         payndaIDString(accountWallet.Wallet.ID),
			CreateTime: accountWallet.Wallet.CreatedAt.UTC(),
			UpdateTime: accountWallet.Wallet.UpdatedAt.UTC(),
			Name:       accountWallet.Account.Name,
			Currency:   accountWallet.Wallet.Currency,
			Amount:     accountWallet.Wallet.Amount.String(),
		},
	}

	return &result, nil
}

type PayndaBalanceAccountWalletTransferRequest struct {
	RequestID        string              `header:"requestId" binding:"required"`
	BalanceAccountID string              `json:"balanceAccountId"` // Invalid: the path-independent mock account is resolved internally.
	Type             paynda.TransferType `json:"type" binding:"required"`
	Currency         common.Currency     `json:"currency" binding:"required"`
	Amount           string              `json:"amount" binding:"required"`
}

type PayndaBalanceAccountWalletTransferData struct {
	ID               string              `json:"id"`
	CreateTime       time.Time           `json:"createTime"`
	UpdateTime       time.Time           `json:"updateTime"`
	BalanceAccountID string              `json:"balanceAccountId"`
	Type             paynda.TransferType `json:"type"`
	Currency         common.Currency     `json:"currency"`
	Amount           string              `json:"amount"`
}

func (s *PayndaOpenAPIService) TransferBalanceAccountWallet(
	ctx context.Context,
	req *PayndaBalanceAccountWalletTransferRequest,
) (*PayndaBalanceAccountWalletTransferData, error) {
	if req.Type != paynda.TransferType_In || req.Currency != common.Currency_USD {
		return nil, biz.ErrInvalidOperation
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || !amount.IsPositive() {
		return nil, biz.ErrInvalidOperation
	}
	accountWallet, err := s.usecase.TransferAccountWallet(ctx, &biz.PayndaAccountWalletTransferRequest{
		Amount: amount,
	})
	if err != nil {
		return nil, err
	}

	return &PayndaBalanceAccountWalletTransferData{
		ID:               payndaIDString(accountWallet.Wallet.ID),
		CreateTime:       accountWallet.Wallet.CreatedAt.UTC(),
		UpdateTime:       accountWallet.Wallet.UpdatedAt.UTC(),
		BalanceAccountID: payndaIDString(accountWallet.Account.ID),
		Type:             req.Type,
		Currency:         req.Currency,
		Amount:           amount.String(),
	}, nil
}

type PayndaRequestResultRequest struct {
	RequestID string `form:"requestId" binding:"required"`
}
type PayndaRequestResultData struct {
	ID         string    `json:"id"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
	RequestID  string    `json:"requestId"`
	Result     string    `json:"result"`
}

func (s *PayndaOpenAPIService) RequestResult(ctx context.Context, req *PayndaRequestResultRequest) (*PayndaRequestResultData, error) {
	result, err := s.usecase.FindRequestResult(ctx, req.RequestID)
	if err != nil {
		return nil, err
	}

	var (
		data      any
		id        string
		createdAt time.Time
		updatedAt time.Time
	)
	if result.IsCardCreate {
		data = payndaCardDetail(result.Card, "")
		id = payndaIDString(result.Card.ID)
		createdAt = result.Card.CreatedAt
		updatedAt = result.Card.UpdatedAt
	} else if result.Transaction != nil {
		data = payndaCardBalanceTransferData(result.Transaction)
		id = payndaIDString(result.Transaction.ID)
		createdAt = result.Transaction.CreatedAt
		updatedAt = result.Transaction.UpdatedAt
	} else {
		data = struct{}{}
		id = payndaIDString(result.Card.ID)
		createdAt = result.Card.CreatedAt
		updatedAt = result.Card.UpdatedAt
	}
	raw, err := json.Marshal(PayndaEmbeddedResponse{
		Code:    paynda.SuccessCode,
		Message: paynda.SuccessMessage,
		Data:    data,
		Success: true,
	})
	if err != nil {
		return nil, biz.ErrInvalidOperation
	}
	return &PayndaRequestResultData{
		ID:         id,
		CreateTime: createdAt.UTC(),
		UpdateTime: updatedAt.UTC(),
		RequestID:  req.RequestID,
		Result:     string(raw),
	}, nil
}

type PayndaCardStatusRequest struct {
	BalanceAccountID string `uri:"balanceAccountId"`
	CardID           string `uri:"cardId"`
	RequestID        string `header:"requestId"`
}

func (s *PayndaOpenAPIService) FreezeCard(ctx context.Context, req *PayndaCardStatusRequest) (*struct{}, error) {
	cardID, err := payndaID(req.CardID)
	if err != nil {
		return nil, err
	}
	_, err = s.usecase.UpdateCardStatus(ctx, &biz.PayndaUpdateCardStatusRequest{
		CardID:    cardID,
		RequestID: req.RequestID,
		Status:    paynda.CardStatus_Frozen,
	})
	return &struct{}{}, err
}
func (s *PayndaOpenAPIService) UnfreezeCard(ctx context.Context, req *PayndaCardStatusRequest) (*struct{}, error) {
	cardID, err := payndaID(req.CardID)
	if err != nil {
		return nil, err
	}
	_, err = s.usecase.UpdateCardStatus(ctx, &biz.PayndaUpdateCardStatusRequest{
		CardID:    cardID,
		RequestID: req.RequestID,
		Status:    paynda.CardStatus_Active,
	})
	return &struct{}{}, err
}

func (s *PayndaOpenAPIService) ReleaseCard(ctx context.Context, req *PayndaCardStatusRequest) (*struct{}, error) {
	cardID, err := payndaID(req.CardID)
	if err != nil {
		return nil, err
	}
	_, err = s.usecase.ReleaseCard(ctx, cardID, req.RequestID)
	if err != nil {
		return nil, err
	}

	return &struct{}{}, nil
}

type PayndaCardBalanceTransferRequest struct {
	BalanceAccountID string              `uri:"balanceAccountId"`
	RequestID        string              `header:"requestId" binding:"required"`
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
	cardID, err := payndaID(req.CardID)
	if err != nil {
		return nil, err
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || !amount.IsPositive() {
		return nil, biz.ErrInvalidOperation
	}
	if req.Type != paynda.TransferType_In && req.Type != paynda.TransferType_Out {
		return nil, biz.ErrInvalidOperation
	}
	transaction, err := s.usecase.TransferCardBalance(ctx, &biz.PayndaTransferRequest{
		CardID:    cardID,
		RequestID: req.RequestID,
		Amount:    amount,
		Type:      req.Type,
	})
	if err != nil {
		return nil, err
	}

	return payndaCardBalanceTransferData(transaction), nil
}

type PayndaCardBalanceUpdateHistoryData struct {
	ID         string    `json:"id"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
	CardID     string    `json:"cardId"`
	NewAmount  string    `json:"newAmount"`
	OldAmount  string    `json:"oldAmount"`
}

func (s *PayndaOpenAPIService) ListCardBalanceUpdates(
	ctx context.Context,
	req *PayndaListRequest,
) (*[]*PayndaCardBalanceUpdateHistoryData, error) {
	offset, limit, _ := payndaPagination(req.Current, req.PageSize)
	items, err := s.usecase.ListCardBalanceUpdates(ctx, &biz.PayndaListRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	records := make([]*PayndaCardBalanceUpdateHistoryData, 0, len(items))
	for _, item := range items {
		oldAmount, err := decimal.NewFromString(string(item.RawPayload))
		if err != nil {
			return nil, biz.ErrInvalidOperation
		}
		newAmount := oldAmount
		if item.Type == common.CardTransactionType_FundIn {
			newAmount = newAmount.Add(item.TxAmount)
		} else {
			newAmount = newAmount.Sub(item.TxAmount)
		}
		records = append(records, &PayndaCardBalanceUpdateHistoryData{
			ID:         payndaIDString(item.ID),
			CreateTime: item.CreatedAt.UTC(),
			UpdateTime: item.UpdatedAt.UTC(),
			CardID:     payndaIDString(item.CardID),
			NewAmount:  newAmount.String(),
			OldAmount:  oldAmount.String(),
		})
	}

	return &records, nil
}

type PayndaTransactionsRequest struct {
	BalanceAccountID     string `uri:"balanceAccountId"`
	Current              int    `form:"current"`
	PageSize             int    `form:"pageSize"`
	CardholderID         string `form:"cardholderId"` // Invalid: generic transactions do not repeat the cardholder relation.
	CardID               string `form:"cardId"`
	TransactionTimeStart string `form:"transactionTimeStart"`
	TransactionTimeEnd   string `form:"transactionTimeEnd"`
}

type PayndaTransactionData struct {
	ID                string                 `json:"id"`
	CreateTime        time.Time              `json:"createTime"`
	UpdateTime        time.Time              `json:"updateTime"`
	CardID            string                 `json:"cardId"`
	MaskCardNo        string                 `json:"maskCardNo"`
	Type              paynda.TransactionType `json:"type"`
	ApprovalCode      string                 `json:"approvalCode"`
	PreAuthAmount     string                 `json:"preAuthAmount"`
	PostedAmount      string                 `json:"postedAmount"`
	Currency          common.Currency        `json:"currency"`
	TransactionTime   string                 `json:"transactionTime"`
	AuthorizationTime string                 `json:"authorizationTime"`
	MerchantMcc       string                 `json:"merchantMcc"`
	MerchantName      string                 `json:"merchantName"`
	DeclineMessage    string                 `json:"declineMessage"`
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
	offset, limit, current := payndaPagination(req.Current, req.PageSize)
	start, err := payndaTransactionTime(req.TransactionTimeStart)
	if err != nil {
		return nil, err
	}
	end, err := payndaTransactionTime(req.TransactionTimeEnd)
	if err != nil {
		return nil, err
	}
	cardID := model.ID(0)
	if req.CardID != "" {
		cardID, err = payndaID(req.CardID)
		if err != nil {
			return nil, err
		}
	}
	items, err := s.usecase.ListCardTransactions(ctx, &biz.PayndaListTransactionsRequest{
		PayndaListRequest: biz.PayndaListRequest{
			Offset: offset,
			Limit:  limit,
		},
		CardID:        cardID,
		OccurredAtGTE: start,
		OccurredAtLTE: end,
	})
	if err != nil {
		return nil, err
	}

	records := make([]*PayndaTransactionData, 0, len(items))
	for _, item := range items {
		records = append(records, payndaTransactionData(item))
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
	id, err := payndaID(req.TransactionID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCardTransaction(ctx, id)
	if err != nil {
		return nil, err
	}

	return payndaTransactionData(item), nil
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
	ID               string            `json:"id"`
	CreateTime       time.Time         `json:"createTime"`
	UpdateTime       time.Time         `json:"updateTime"`
	BalanceAccountID string            `json:"balanceAccountId"`
	CardholderID     string            `json:"cardholderId"`
	Status           paynda.CardStatus `json:"status"`
	MaskCardNo       string            `json:"maskCardNo"`
	CardScheme       common.CardScheme `json:"cardScheme"`
	CardBin          string            `json:"cardBin"`
	Currency         common.Currency   `json:"currency"`
	CreditLimitType  string            `json:"creditLimitType"`
	AmountUsed       string            `json:"amountUsed"`
	Amount           string            `json:"amount"`
}
type PayndaCardSensitiveData struct {
	ID             string    `json:"id"`
	CreateTime     time.Time `json:"createTime"`
	UpdateTime     time.Time `json:"updateTime"`
	CardID         string    `json:"cardId"`
	CVV            string    `json:"cvv"`
	ExpirationDate string    `json:"expirationDate"`
	CardNo         string    `json:"cardNo"`
}
type PayndaCardBalanceData struct {
	ID              string    `json:"id"`
	CreateTime      time.Time `json:"createTime"`
	UpdateTime      time.Time `json:"updateTime"`
	AmountUsed      string    `json:"amountUsed"`
	AmountFrozen    string    `json:"amountFrozen"`
	AvailableAmount string    `json:"availableAmount"`
	Amount          string    `json:"amount"`
}

func payndaCardholderData(item *model.CardHolder, balanceAccountID string) *PayndaCardholderData {
	return &PayndaCardholderData{
		ID:                  payndaIDString(item.ID),
		CreateTime:          item.CreatedAt.UTC(),
		UpdateTime:          item.UpdatedAt.UTC(),
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

func payndaCardDetail(item *model.Card, balanceAccountID string) *PayndaCardDetail {
	return &PayndaCardDetail{
		Card:          payndaCardData(item, balanceAccountID),
		CardSensitive: payndaCardSensitiveData(item),
		CardBalance: &PayndaCardBalanceData{
			ID:              payndaIDString(item.ID),
			CreateTime:      item.CreatedAt.UTC(),
			UpdateTime:      item.UpdatedAt.UTC(),
			AvailableAmount: decimal.Zero.String(),
			Amount:          decimal.Zero.String(),
			AmountUsed:      decimal.Zero.String(),
			AmountFrozen:    decimal.Zero.String(),
		},
	}
}

func payndaCardData(item *model.Card, balanceAccountID string) *PayndaCardData {
	return &PayndaCardData{
		ID:               payndaIDString(item.ID),
		CreateTime:       item.CreatedAt.UTC(),
		UpdateTime:       item.UpdatedAt.UTC(),
		BalanceAccountID: balanceAccountID,
		CardholderID:     payndaIDString(item.CardHolderID),
		Status:           paynda.CardStatusFromGeneric(item.Status),
		MaskCardNo:       "******" + item.CardNumber[len(item.CardNumber)-4:],
		CardScheme:       item.CardScheme,
		CardBin:          item.CardBin,
		Currency:         item.CardCurrency,
		CreditLimitType:  "INDEPENDENT",
		AmountUsed:       decimal.Zero.String(),
		Amount:           decimal.Zero.String(),
	}
}

func payndaCardSensitiveData(item *model.Card) *PayndaCardSensitiveData {
	return &PayndaCardSensitiveData{
		ID:             payndaIDString(item.ID),
		CreateTime:     item.CreatedAt.UTC(),
		UpdateTime:     item.UpdatedAt.UTC(),
		CardID:         payndaIDString(item.ID),
		CVV:            item.Cvv,
		ExpirationDate: item.ExpireAt.Format("01/06"),
		CardNo:         item.CardNumber,
	}
}

func payndaCardBalanceData(item *model.Wallet) *PayndaCardBalanceData {
	return &PayndaCardBalanceData{
		ID:              payndaIDString(item.ID),
		CreateTime:      item.CreatedAt.UTC(),
		UpdateTime:      item.UpdatedAt.UTC(),
		AmountUsed:      decimal.Zero.String(),
		AmountFrozen:    item.PendingOut.String(),
		AvailableAmount: item.Amount.String(),
		Amount:          item.Amount.String(),
	}
}

func payndaBalanceAccountWalletData(
	item *biz.PayndaAccountWallet,
	balanceAccountID string,
) *PayndaBalanceAccountWalletData {
	return &PayndaBalanceAccountWalletData{
		ID:               payndaIDString(item.Wallet.ID),
		CreateTime:       item.Wallet.CreatedAt.UTC(),
		UpdateTime:       item.Wallet.UpdatedAt.UTC(),
		BalanceAccountID: balanceAccountID,
		Currency:         item.Wallet.Currency,
		Amount:           item.Wallet.Amount.String(),
		FrozenAmount:     item.Wallet.PendingOut.String(),
	}
}

func payndaCardBalanceTransferData(item *model.CardTransaction) *PayndaCardBalanceTransferData {
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
		CardID:    payndaIDString(item.CardID),
		NewAmount: newAmount.String(),
		OldAmount: oldAmount.String(),
		Amount:    item.TxAmount.String(),
		Type:      transferType,
	}
}

func payndaTransactionData(item *model.CardTransaction) *PayndaTransactionData {
	transactionTime := item.OccurredAt.Format(time.DateTime)
	return &PayndaTransactionData{
		ID:                payndaIDString(item.ID),
		CreateTime:        item.CreatedAt.UTC(),
		UpdateTime:        item.UpdatedAt.UTC(),
		CardID:            payndaIDString(item.CardID),
		Type:              paynda.TransactionTypeFromGeneric(item.Type),
		ApprovalCode:      item.AuthorizationCode,
		PreAuthAmount:     item.TxAmount.String(),
		PostedAmount:      item.TxAmount.String(),
		Currency:          item.Currency,
		TransactionTime:   transactionTime,
		AuthorizationTime: transactionTime,
		MerchantMcc:       item.MerchantMCC,
		MerchantName:      item.MerchantName,
	}
}

func payndaPagination(current int, pageSize int) (int, int, int) {
	if current < 1 {
		current = 1
	}
	if pageSize < 1 {
		pageSize = 100
	}

	return (current - 1) * pageSize, pageSize, current
}

func payndaTransactionTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.DateTime, value)
	if err != nil {
		return nil, biz.ErrInvalidOperation
	}
	parsed = parsed.UTC()

	return &parsed, nil
}
