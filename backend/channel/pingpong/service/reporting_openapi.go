package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	ping "generic-mock/channel/pingpong/enums"
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/pkg/idconv"
	"generic-mock/channel/pingpong/pkg/queryconv"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/timefmt"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
)

type AccountBalancesRequest struct {
	OpenAPIRequest
	AccountType      *string           `json:"account_type"`
	CurrencyList     []common.Currency `json:"currency_list"`
	SubaccountIDList []string          `json:"subaccount_id_list"`
	PageNo           *int              `json:"page_no" binding:"omitempty,min=1"`
	PageSize         *int              `json:"page_size" binding:"omitempty,min=1,max=100"`
}

type AccountBalanceData struct {
	SubaccountID    string          `json:"subaccount_id"`
	AccountType     string          `json:"account_type"`
	Currency        common.Currency `json:"currency"`
	Source          string          `json:"source"`
	Status          string          `json:"status"`
	AvailableAmount Number          `json:"available_amount"`
	FrozenAmount    Number          `json:"frozen_amount"`
	TotalAmount     Number          `json:"total_amount"`
}

type AccountBalancesData struct {
	PageNo   int                  `json:"page_no"`
	PageSize int                  `json:"page_size"`
	TotalNum int64                `json:"total_num"`
	ItemList []AccountBalanceData `json:"item_list"`
}

type CardTransactionsRequest struct {
	OpenAPIRequest
	PageRequest
	CardID           *string `form:"card_id"`
	Type             *string `form:"type"`
	StartTime        *string `form:"start_time"`
	EndTime          *string `form:"end_time"`
	StartPostingDate *string `form:"start_posting_date"`
	EndPostingDate   *string `form:"end_posting_date"`
	StartCreatedDate *string `form:"start_created_date"`
	EndCreatedDate   *string `form:"end_created_date"`
	ClearType        *string `form:"clear_type"`
	Status           *string `form:"status"`
}

type CardTransactionData struct {
	AuthorizationID     string              `json:"authorization_id"`
	CardNumber          string              `json:"card_number"`
	CardID              string              `json:"card_id"`
	BudgetID            string              `json:"budget_id"`
	TransactionDate     time.Time           `json:"transaction_date"`
	BillingAmount       string              `json:"billing_amount"`
	BillingCurrency     common.Currency     `json:"billing_currency"`
	TransactionAmount   string              `json:"transaction_amount"`
	TransactionCurrency common.Currency     `json:"transaction_currency"`
	MerchantName        string              `json:"merchant_name"`
	MerchantCountry     string              `json:"merchant_country"`
	MCC                 string              `json:"mcc"`
	Remark              string              `json:"remark"`
	Type                string              `json:"type"`
	Status              string              `json:"status"`
	FailReason          string              `json:"fail_reason"`
	ApproveCode         string              `json:"approve_code"`
	Fees                TransactionFeesData `json:"fees"`
}

type TransactionFeesData struct {
	RateFee            string          `json:"rate_fee"`
	RateFeeCurrency    common.Currency `json:"rate_fee_currency"`
	ThreeDSFee         string          `json:"threeds_fee"`
	ThreeDSFeeCurrency common.Currency `json:"threeds_fee_currency"`
}

type CardTransactionsData struct {
	TotalNum        int64                 `json:"total_num"`
	PageNo          int                   `json:"page_no"`
	PageSize        int                   `json:"page_size"`
	TransactionList []CardTransactionData `json:"transaction_list"`
}

type ThreeDSRequest struct {
	OpenAPIRequest
	CardID string `form:"card_id" binding:"required"`
}

type ThreeDSDetailsData struct {
	CardholderID     string `json:"cardholder_id"`
	BudgetID         string `json:"budget_id"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	DateOfBirth      string `json:"date_of_birth"`
	CallPrefix       string `json:"call_prefix"`
	Mobile           string `json:"mobile"`
	PostCode         string `json:"post_code"`
	CountryCode      string `json:"country_code"`
	State            string `json:"state"`
	City             string `json:"city"`
	AddressLine      string `json:"address_line"`
	Email            string `json:"email"`
	SecurityIndex    string `json:"security_index"`
	SecurityQuestion string `json:"security_question"`
	SecurityAnswer   string `json:"security_answer"`
}

type AccountTransactionsRequest struct {
	OpenAPIRequest
	PageRequest
	BudgetID         *string `form:"budget_id"`
	CardID           *string `form:"card_id"`
	PostingStartTime *string `form:"posting_start_time"`
	PostingEndTime   *string `form:"posting_end_time"`
	TransactionType  *string `form:"transaction_type"`
	Direction        *string `form:"direction"`
}

type AccountTransactionData struct {
	TransactionID          string                     `json:"transaction_id"`
	TransactionType        string                     `json:"transaction_type"`
	TransactionDescription string                     `json:"transaction_description"`
	AccountType            string                     `json:"account_type"`
	BudgetID               string                     `json:"budget_id"`
	BudgetName             string                     `json:"budget_name"`
	CardID                 string                     `json:"card_id"`
	IsOTA                  bool                       `json:"is_ota"`
	CardNumber             string                     `json:"card_number"`
	CardRemark             string                     `json:"card_remark"`
	TransactionTime        time.Time                  `json:"transaction_time"`
	PostingTime            time.Time                  `json:"posting_time"`
	TransactionAmount      Number                     `json:"transaction_amount"`
	TransactionCurrency    common.Currency            `json:"transaction_currency"`
	BillingAmount          Number                     `json:"billing_amount"`
	BillingCurrency        common.Currency            `json:"billing_currency"`
	Direction              string                     `json:"direction"`
	BalancePaid            Number                     `json:"balance_paid"`
	OutstandingAmount      Number                     `json:"outstanding_amount"`
	BeforeBalance          Number                     `json:"before_balance"`
	AfterBalance           Number                     `json:"after_balance"`
	ReconciliationInfo     *AccountReconciliationInfo `json:"reconciliation_info"`
}

type AccountReconciliationInfo struct {
	ServiceType  string `json:"service_type"`
	CustomField1 string `json:"custom_field_1"`
}

type AccountTransactionsData struct {
	TotalNum        int64                    `json:"total_num"`
	PageNo          int                      `json:"page_no"`
	PageSize        int                      `json:"page_size"`
	TransactionList []AccountTransactionData `json:"transaction_list"`
}

func (s *PingPongOpenAPIService) QueryAccountsBalances(
	ctx context.Context,
	req *AccountBalancesRequest,
) (*AccountBalancesData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	if req.AccountType != nil {
		if !ping.AccountTypeValid(*req.AccountType) {
			return nil, pingerrors.ErrInvalid
		}
	}
	page, limit, err := resolveJSONPagination(req.PageNo, req.PageSize)
	if err != nil {
		return nil, err
	}
	virtualAccountIDs, err := idconv.FromStringList(req.SubaccountIDList)
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.AccountBalanceReport(ctx, &biz.AccountBalanceReportRequest{
		AccountID:         accountID,
		Currencies:        req.CurrencyList,
		VirtualAccountIDs: virtualAccountIDs,
		Offset:            (page - 1) * limit,
		Limit:             &limit,
	})
	if err != nil {
		return nil, err
	}
	result := &AccountBalancesData{
		PageNo:   page,
		PageSize: limit,
		TotalNum: total,
		ItemList: make([]AccountBalanceData, 0, len(items)),
	}
	for _, item := range items {
		if item.Wallet == nil {
			return nil, pingerrors.ErrInvalid
		}
		result.ItemList = append(result.ItemList, AccountBalanceData{
			SubaccountID:    idconv.ToString(item.ID),
			AccountType:     "budget",
			Currency:        item.Wallet.Currency,
			Source:          "mock",
			Status:          "ACTIVE",
			AvailableAmount: Number{item.Wallet.Available},
			FrozenAmount:    Number{item.Wallet.PendingOut},
			TotalAmount:     Number{item.Wallet.Available.Add(item.Wallet.PendingOut)},
		})
	}
	return result, nil
}

func (s *PingPongOpenAPIService) QueryCardTransactions(
	ctx context.Context,
	req *CardTransactionsRequest,
) (*CardTransactionsData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromOptionalString(req.CardID)
	if err != nil {
		return nil, err
	}
	var transactionTypes []common.CardTransactionType
	if req.Type != nil {
		var valid bool
		transactionTypes, valid = ping.TransactionTypesToGeneric(*req.Type)
		if !valid {
			return nil, pingerrors.ErrInvalid
		}
	}
	var statuses []common.CardTransactionStatus
	if req.Status != nil {
		var valid bool
		statuses, valid = ping.TransactionStatusesToGeneric(*req.Status)
		if !valid {
			return nil, pingerrors.ErrInvalid
		}
	}
	if req.ClearType != nil {
		var valid bool
		transactionTypes, valid = ping.ApplyTransactionDirectionToGenericTypes(
			transactionTypes,
			ping.TransactionDirection(*req.ClearType),
		)
		if !valid {
			return nil, pingerrors.ErrInvalid
		}
	}
	createdFrom, createdTo, valid := queryconv.CardTransactionTimeRange(queryconv.CardTransactionTimeRangeRequest{
		StartTime:        req.StartTime,
		StartCreatedDate: req.StartCreatedDate,
		StartPostingDate: req.StartPostingDate,
		EndTime:          req.EndTime,
		EndCreatedDate:   req.EndCreatedDate,
		EndPostingDate:   req.EndPostingDate,
	})
	if !valid {
		return nil, pingerrors.ErrInvalid
	}
	page, limit, err := req.PageRequest.resolvePagination()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.CardTransactionReport(ctx, &biz.CardTransactionReportRequest{
		AccountID:   accountID,
		CardID:      cardID,
		Types:       transactionTypes,
		Statuses:    statuses,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		Offset:      (page - 1) * limit,
		Limit:       &limit,
	})
	if err != nil {
		return nil, err
	}
	return toCardTransactionsData(items, total, page, limit), nil
}

func (s *PingPongOpenAPIService) Query3DSDetails(ctx context.Context, req *ThreeDSRequest) (*ThreeDSDetailsData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.uc.GetCard(ctx, &biz.GetCardRequest{AccountID: accountID, ID: cardID})
	if err != nil {
		return nil, err
	}
	if card.VirtualAccountID == nil {
		return nil, pingerrors.ErrInvalid
	}
	data := &ThreeDSDetailsData{BudgetID: idconv.ToString(*card.VirtualAccountID)}
	if card.CardHolderID > 0 {
		data.CardholderID = idconv.ToString(card.CardHolderID)
	}
	if card.CardHolder != nil {
		holder := card.CardHolder
		data.FirstName = holder.FirstName
		data.LastName = holder.LastName
		if holder.DateOfBirth != nil {
			data.DateOfBirth = timefmt.Date(*holder.DateOfBirth)
		}
		data.CallPrefix = holder.MobilePrefix
		data.Mobile = holder.Mobile
		data.PostCode = holder.ResidentialPostalCode
		data.CountryCode = holder.ResidentialCountryCode
		data.State = holder.ResidentialState
		data.City = holder.ResidentialCity
		data.AddressLine = holder.ResidentialAddress
		data.Email = holder.Email
	}
	return data, nil
}

func (s *PingPongOpenAPIService) QueryAccountTransactions(
	ctx context.Context,
	req *AccountTransactionsRequest,
) (*AccountTransactionsData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	virtualAccountID, err := idconv.FromOptionalString(req.BudgetID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromOptionalString(req.CardID)
	if err != nil {
		return nil, err
	}
	if virtualAccountID == nil && cardID == nil {
		return nil, pingerrors.ErrInvalid
	}
	var transactionTypes []common.CardTransactionType
	if req.TransactionType != nil {
		var valid bool
		transactionTypes, valid = ping.TransactionTypesToGeneric(*req.TransactionType)
		if !valid {
			return nil, pingerrors.ErrInvalid
		}
	}
	if req.Direction != nil {
		var valid bool
		transactionTypes, valid = ping.ApplyTransactionDirectionToGenericTypes(
			transactionTypes,
			ping.TransactionDirection(*req.Direction),
		)
		if !valid {
			return nil, pingerrors.ErrInvalid
		}
	}
	createdFrom, valid := queryconv.OptionalTime(req.PostingStartTime)
	if !valid {
		return nil, pingerrors.ErrInvalid
	}
	createdTo, valid := queryconv.OptionalTime(req.PostingEndTime)
	if !valid {
		return nil, pingerrors.ErrInvalid
	}
	if createdFrom != nil && createdTo != nil && createdFrom.After(*createdTo) {
		return nil, pingerrors.ErrInvalid
	}
	page, limit, err := req.PageRequest.resolvePagination()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.AccountTransactionReport(ctx, &biz.AccountTransactionReportRequest{
		AccountID:        accountID,
		VirtualAccountID: virtualAccountID,
		CardID:           cardID,
		Types:            transactionTypes,
		CreatedFrom:      createdFrom,
		CreatedTo:        createdTo,
		Offset:           (page - 1) * limit,
		Limit:            &limit,
	})
	if err != nil {
		return nil, err
	}
	return toAccountTransactionsData(items, total, page, limit), nil
}

func resolveJSONPagination(pageValue *int, sizeValue *int) (int, int, error) {
	page := types.Value(pageValue)
	size := types.Value(sizeValue)
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = types.DefaultPageSize
	}
	if page < 1 || size < 1 || size > 100 || page > 1000000 {
		return 0, 0, pingerrors.ErrInvalid
	}
	return page, size, nil
}

func toCardTransactionsData(items []biz.TransactionReportItem, total int64, page int, limit int) *CardTransactionsData {
	result := &CardTransactionsData{
		TotalNum:        total,
		PageNo:          page,
		PageSize:        limit,
		TransactionList: make([]CardTransactionData, 0, len(items)),
	}
	for _, item := range items {
		result.TransactionList = append(result.TransactionList, toCardTransactionData(item))
	}
	return result
}

func toCardTransactionData(item biz.TransactionReportItem) CardTransactionData {
	transaction := item.Transaction
	cardID := ""
	cardNumber := ""
	budgetID := ""
	if transaction.CardID > 0 {
		cardID = idconv.ToString(transaction.CardID)
	}
	if item.Card != nil {
		cardNumber = item.Card.CardNumber
		if item.Card.VirtualAccountID != nil {
			budgetID = idconv.ToString(*item.Card.VirtualAccountID)
		}
	}
	authorizationID := ""
	if transaction.AuthorizationID > 0 {
		authorizationID = idconv.ToString(transaction.AuthorizationID)
	}
	feeCurrency := transaction.Currency
	if feeCurrency == "" {
		feeCurrency = transaction.TxCurrency
	}
	return CardTransactionData{
		AuthorizationID:     authorizationID,
		CardNumber:          cardNumber,
		CardID:              cardID,
		BudgetID:            budgetID,
		TransactionDate:     transaction.CreatedAt,
		BillingAmount:       transaction.TxAmount.String(),
		BillingCurrency:     transaction.Currency,
		TransactionAmount:   transaction.TxAmount.String(),
		TransactionCurrency: transaction.TxCurrency,
		MerchantName:        transaction.MerchantName,
		MerchantCountry:     transaction.MerchantCountry,
		MCC:                 transaction.MerchantMCC,
		Type:                string(transaction.Type),
		Status:              string(transaction.Status),
		ApproveCode:         transaction.AuthorizationCode,
		Fees: TransactionFeesData{
			RateFee:            decimal.Zero.StringFixed(2),
			RateFeeCurrency:    feeCurrency,
			ThreeDSFee:         decimal.Zero.StringFixed(2),
			ThreeDSFeeCurrency: feeCurrency,
		},
	}
}

func toAccountTransactionsData(items []biz.TransactionReportItem, total int64, page int, limit int) *AccountTransactionsData {
	result := &AccountTransactionsData{
		TotalNum:        total,
		PageNo:          page,
		PageSize:        limit,
		TransactionList: make([]AccountTransactionData, 0, len(items)),
	}
	for _, item := range items {
		cardTransaction := toCardTransactionData(item)
		transaction := item.Transaction
		direction := "DEBIT"
		if transaction.Type == common.CardTransactionType_REFUND {
			direction = "CREDIT"
		}
		result.TransactionList = append(result.TransactionList, AccountTransactionData{
			TransactionID:          idconv.ToString(transaction.ID),
			TransactionType:        string(transaction.Type),
			TransactionDescription: string(transaction.Type),
			AccountType:            "budget",
			BudgetID:               cardTransaction.BudgetID,
			BudgetName:             reportBudgetName(item.Card),
			CardID:                 cardTransaction.CardID,
			CardNumber:             cardTransaction.CardNumber,
			TransactionTime:        transaction.CreatedAt,
			PostingTime:            transaction.CreatedAt,
			TransactionAmount:      Number{transaction.TxAmount},
			TransactionCurrency:    transaction.TxCurrency,
			BillingAmount:          Number{transaction.TxAmount},
			BillingCurrency:        transaction.Currency,
			Direction:              direction,
			BalancePaid:            Number{transaction.TxAmount},
			OutstandingAmount:      Number{decimal.Zero},
			BeforeBalance:          Number{decimal.Zero},
			AfterBalance:           Number{decimal.Zero},
		})
	}
	return result
}

func reportBudgetName(card *model.Card) string {
	if card == nil || card.VirtualAccount == nil {
		return ""
	}
	return card.VirtualAccount.Name
}
