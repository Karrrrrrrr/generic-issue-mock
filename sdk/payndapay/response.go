package payndapay

import "encoding/json"

// APIResponse represents the standard API response structure
type APIResponse struct {
	Code    int64  `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Success bool   `json:"success,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Pages   int64  `json:"pages,omitempty"`
	Records any    `json:"records,omitempty"`
	Total   int64  `json:"total,omitempty"`
	Size    int64  `json:"size,omitempty"`
	Current int64  `json:"current,omitempty"`
	Orders  any    `json:"orders,omitempty"`
	Column  string `json:"column,omitempty"`
	Asc     bool   `json:"asc,omitempty"`
	// Success bool   `json:"success,omitempty"`
}

// Merchant represents a merchant entity
type Merchant struct {
	ID         string `json:"id,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
	UpdateTime string `json:"updateTime,omitempty"`
}

type MerchantWallet struct {
	ID         string `json:"id,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
	UpdateTime string `json:"updateTime,omitempty"`
	Name       string `json:"name,omitempty"`
	MerchantId string `json:"merchantId,omitempty"`
	Currency   string `json:"currency,omitempty"`
	Amount     string `json:"amount,omitempty"`
}

// BalanceAccount represents a balance account
type BalanceAccount struct {
	ID         string `json:"id,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
	UpdateTime string `json:"updateTime,omitempty"`
	Name       string `json:"name,omitempty"`
	MerchantId string `json:"merchantId,omitempty"`
}

// BalanceAccountWallet represents a wallet within a balance account
type BalanceAccountWallet struct {
	ID               string `json:"id,omitempty"`
	CreateTime       string `json:"createTime,omitempty"`
	UpdateTime       string `json:"updateTime,omitempty"`
	MerchantId       string `json:"merchantId,omitempty"`
	BalanceAccountId string `json:"balanceAccountId,omitempty"`
	Currency         string `json:"currency,omitempty"`
	Amount           string `json:"amount,omitempty"`
	FrozenAmount     string `json:"frozenAmount,omitempty"`
}
type BalanceAccountWalletTransfer struct {
	ID               string `json:"id,omitempty"`
	CreateTime       string `json:"createTime,omitempty"`
	UpdateTime       string `json:"updateTime,omitempty"`
	MerchantID       string `json:"merchantId,omitempty"`
	BalanceAccountID string `json:"balanceAccountId,omitempty"`
	Type             string `json:"type,omitempty"`
	Currency         string `json:"currency,omitempty"`
	Amount           string `json:"amount,omitempty"`
}

// Cardholder represents a cardholder
type Cardholder struct {
	ID                  string `json:"id,omitempty"`
	CreateTime          string `json:"createTime,omitempty"`
	UpdateTime          string `json:"updateTime,omitempty"`
	MerchantId          string `json:"merchantId,omitempty"`
	BalanceAccountId    string `json:"balanceAccountId,omitempty"`
	FirstName           string `json:"firstName,omitempty"`
	LastName            string `json:"lastName,omitempty"`
	MobilePrefix        string `json:"mobilePrefix,omitempty"`
	Mobile              string `json:"mobile,omitempty"`
	Email               string `json:"email,omitempty"`
	BillingAddressLine1 string `json:"billingAddressLine1,omitempty"`
	BillingAddressLine2 string `json:"billingAddressLine2,omitempty"`
	BillingCity         string `json:"billingCity,omitempty"`
	BillingCountryCode  string `json:"billingCountryCode,omitempty"`
	BillingPostalCode   string `json:"billingPostalCode,omitempty"`
	BillingState        string `json:"billingState,omitempty"`
	UnlimitedBalance    bool   `json:"unlimitedBalance,omitempty"`
}

// CardholderWallet
type CardholderWallet struct {
	ID               string `json:"id,omitempty"`
	CreateTime       string `json:"createTime,omitempty"`
	UpdateTime       string `json:"updateTime,omitempty"`
	MerchantId       string `json:"merchantId,omitempty"`
	BalanceAccountId string `json:"balanceAccountId,omitempty"`
	CardholderID     string `json:"cardholderId,omitempty"`
	Currency         string `json:"currency,omitempty"`
	Amount           string `json:"amount,omitempty"`
	FrozenAmount     string `json:"frozenAmount,omitempty"`
}

// map[string]interface {}
// [
// "card": map[string]interface {} ["createTime": *(*interface {})(0x140000744a0), "merchantId": *(*interface {})(0x140000744c0), "firstName": *(*interface {})(0x140000744e0), "mobilePrefix": *(*interface {})(0x14000074500), "id": *(*interface {})(0x140000745a8), "cardholderId": *(*interface {})(0x140000745c8), "creditLimitType": *(*interface {})(0x140000745e8), "updateTime": *(*interface {})(0x140000746b0), "status": *(*interface {})(0x140000746d0), "maskCardNo": *(*interface {})(0x140000746f0), "mobile": *(*interface {})(0x14000074710), "email": *(*interface {})(0x14000074730), "singleUse": *(*interface {})(0x14000074750), "balanceAccountId": *(*interface {})(0x140000747b8), "cardBin": *(*interface {})(0x140000747d8), "currency": *(*interface {})(0x140000747f8), "lastName": *(*interface {})(0x14000074818), ],
// "sensitiveInfo": map[string]interface {} ["id": *(*interface {})(0x1400021a258), "createTime": *(*interface {})(0x1400021a278), "updateTime": *(*interface {})(0x1400021a298), "cardId": *(*interface {})(0x1400021a2b8), "cvv": *(*interface {})(0x1400021a2d8), "expirationDate": *(*interface {})(0x1400021a2f8), "cardNo": *(*interface {})(0x1400021a318), ],
// "balance": map[string]interface {} ["id": *(*interface {})(0x1400021a378), "createTime": *(*interface {})(0x1400021a398), "updateTime": *(*interface {})(0x1400021a3b8), "amountUsed": *(*interface {})(0x1400021a3d8), "amountFrozen": *(*interface {})(0x1400021a3f8), "amount": *(*interface {})(0x1400021a418), "availableAmount": *(*interface {})(0x1400021a438), ],
// ]

type CardDetail struct {
	Card          Card          `json:"card"`
	CardSensitive CardSensitive `json:"sensitiveInfo"`
	CardBalance   CardBalance   `json:"balance"`
}

// Card represents a card
type Card struct {
	ID                  string `json:"id,omitempty"`
	CreateTime          string `json:"createTime,omitempty"`
	UpdateTime          string `json:"updateTime,omitempty"`
	MerchantId          string `json:"merchantId,omitempty"`
	BalanceAccountId    string `json:"balanceAccountId,omitempty"`
	CardholderId        string `json:"cardholderId,omitempty"`
	Status              string `json:"status,omitempty"` // DELETED, ACTIVE, FROZEN, EXPIRED, BLOCK, UNACTIVE, UNKNOWN
	MaskCardNo          string `json:"maskCardNo,omitempty"`
	CardScheme          string `json:"cardScheme,omitempty"`
	CardBin             string `json:"cardBin,omitempty"`
	Currency            string `json:"currency,omitempty"`
	FirstName           string `json:"firstName,omitempty"`
	LastName            string `json:"lastName,omitempty"`
	MobilePrefix        string `json:"mobilePrefix,omitempty"`
	Mobile              string `json:"mobile,omitempty"`
	Email               string `json:"email,omitempty"`
	BillingCountryCode  string `json:"billingCountryCode,omitempty"`
	BillingAddressLine1 string `json:"billingAddressLine1,omitempty"`
	BillingAddressLine2 string `json:"billingAddressLine2,omitempty"`
	BillingCity         string `json:"billingCity,omitempty"`
	BillingPostalCode   string `json:"billingPostalCode,omitempty"`
	BillingState        string `json:"billingState,omitempty"`
	SingleUse           bool   `json:"singleUse,omitempty"`
	CreditLimitType     string `json:"creditLimitType,omitempty"` // SHARED, INDEPENDENT
	AmountUsed          string `json:"amountUsed,omitempty"`
	Amount              string `json:"amount,omitempty"`
}

type CardSensitive struct {
	ID             string `json:"id,omitempty"`
	CreateTime     string `json:"createTime,omitempty"`
	UpdateTime     string `json:"updateTime,omitempty"`
	CardID         string `json:"cardId,omitempty"`
	CVV            string `json:"cvv,omitempty"`
	ExpirationDate string `json:"expirationDate,omitempty"`
	CardNo         string `json:"cardNo,omitempty"`
}

// CardBalance represents card balance information
type CardBalance struct {
	ID              string `json:"id,omitempty"`
	CreateTime      string `json:"createTime,omitempty"`
	UpdateTime      string `json:"updateTime,omitempty"`
	AmountUsed      string `json:"amountUsed,omitempty"`
	AmountFrozen    string `json:"amountFrozen,omitempty"`
	AvailableAmount string `json:"availableAmount,omitempty"`
	Amount          string `json:"amount,omitempty"`
}

// CardTransaction represents a card transaction
type CardTransaction struct {
	ID                                  string      `json:"id"`
	CreateTime                          string      `json:"createTime"`
	UpdateTime                          string      `json:"updateTime"`
	MerchantID                          json.Number `json:"merchantId"`
	BalanceAccountID                    json.Number `json:"balanceAccountId"`
	CardholderID                        json.Number `json:"cardholderId"`
	CardID                              json.Number `json:"cardId"`
	MaskCardNo                          string      `json:"maskCardNo"`
	Type                                string      `json:"type"`
	ApprovalCode                        string      `json:"approvalCode"`
	PreAuthAmount                       string      `json:"preAuthAmount"`
	PostedAmount                        string      `json:"postedAmount"`
	Currency                            string      `json:"currency"`
	OriginalCurrencyCode                string      `json:"originalCurrencyCode"`
	TransactionAmountInOriginalCurrency string      `json:"transactionAmountInOriginalCurrency"`
	ReversalFlag                        string      `json:"reversalFlag"`
	TransactionTime                     string      `json:"transactionTime"`
	AuthorizationTime                   string      `json:"authorizationTime"`
	AcquirerID                          string      `json:"acquirerId"`
	MerchantMcc                         string      `json:"merchantMcc"`
	MerchantName                        string      `json:"merchantName"`
	MerchantAddressAddressLine1         string      `json:"merchantAddressAddressLine1"`
	MerchantAddressCity                 string      `json:"merchantAddressCity"`
	MerchantAddressState                string      `json:"merchantAddressState"`
	MerchantAddressCountry              string      `json:"merchantAddressCountry"`
	MerchantAddressZip                  string      `json:"merchantAddressZip"`
	PosAcceptorID                       string      `json:"posAcceptorId"`
	PosAcceptLocation                   string      `json:"posAcceptLocation"`
	PosEntryDescription                 string      `json:"posEntryDescription"`
	SupplierTransactionID               string      `json:"supplierTransactionId"`
	SupplierTransactionLinkID           string      `json:"supplierTransactionLinkId"`
	DeclineMessage                      string      `json:"declineMessage"`
	WalletID                            string      `json:"walletId"`
}

// CardTransactionsData represents the paginated `data` block from OpenAPI
// for card transactions.
type CardTransactionsData struct {
	Pages                  int64              `json:"pages,omitempty"`
	Records                []*CardTransaction `json:"records,omitempty"`
	Total                  int64              `json:"total,omitempty"`
	Size                   int64              `json:"size,omitempty"`
	Current                int64              `json:"current,omitempty"`
	Orders                 []interface{}      `json:"orders,omitempty"`
	Column                 string             `json:"column,omitempty"`
	Asc                    bool               `json:"asc,omitempty"`
	OptimizeCountSql       bool               `json:"optimizeCountSql,omitempty"`
	SearchCount            bool               `json:"searchCount,omitempty"`
	OptimizeJoinOfCountSql int64              `json:"optimizeJoinOfCountSql,omitempty"`
	CountId                string             `json:"countId,omitempty"`
	Success                bool               `json:"success,omitempty"`
}

// CardTransactionResponse is the top-level response envelope used by OpenAPI
// endpoints that return card transactions.
type CardTransactionResponse struct {
	Code    int64                 `json:"code,omitempty"`
	Message string                `json:"message,omitempty"`
	Data    *CardTransactionsData `json:"data,omitempty"`
	Success bool                  `json:"success,omitempty"`
}

// CardBin represents card bin information
type CardBin struct {
	ID                string   `json:"id,omitempty"`
	CardBin           string   `json:"cardBin,omitempty"`
	Name              string   `json:"name,omitempty"`
	Type              string   `json:"type,omitempty"` // CARD_CLASS, CARD_BIN
	CardClass         string   `json:"cardClass,omitempty"`
	SupportCurrencies []string `json:"supportCurrencies,omitempty"`
}

// CardControl represents card transaction control limits
type CardControl struct {
	Period           string `json:"period,omitempty"` // DAY, MONTH, TOTAL, ONCE
	TransactionCount int64  `json:"transactionCount,omitempty"`
	Amount           string `json:"amount,omitempty"`
}

type CardBalanceTransfer struct {
	CardID string `json:"cardId,omitempty"`
	Amount string `json:"amount,omitempty"`
	Type   string `json:"type,omitempty"`
}

type CardBalanceUpdateHistory struct {
	ID               string  `json:"id,omitempty"`
	CreateTime       string  `json:"createTime,omitempty"`
	UpdateTime       string  `json:"updateTime,omitempty"`
	MerchantID       string  `json:"merchantId,omitempty"`
	BalanceAccountID string  `json:"balanceAccountId,omitempty"`
	CardholderID     string  `json:"cardholderId,omitempty"`
	CardID           string  `json:"cardId,omitempty"`
	NewAmount        float64 `json:"newAmount,string,omitempty"`
	OldAmount        float64 `json:"oldAmount,string,omitempty"`
}

type CardBalanceTransferReuslt struct {
	CreateTime     string
	Code           int64
	Message        string
	Success        bool
	TransferRecord *CardBalanceTransferRecord
}

type CardBalanceTransferRecord struct {
	ID               string `json:"id,omitempty"`
	RequestID        string `json:"requestId,omitempty"`
	CreateTime       string `json:"createTime,omitempty"`
	UpdateTime       string `json:"updateTime,omitempty"`
	MerchantID       string `json:"merchantId,omitempty"`
	BalanceAccountID string `json:"balanceAccountId,omitempty"`
	CardholderID     string `json:"cardholderId,omitempty"`
	CardID           string `json:"cardId,omitempty"`
	NewAmount        string `json:"newAmount,omitempty"`
	OldAmount        string `json:"oldAmount,omitempty"`
	Amount           string `json:"amount,omitempty"`
	Type             string `json:"type,omitempty"`
}

type CardCreateResult struct {
	CreateTime string
	Code       int64
	Message    string
	Success    bool
	Result     *CardDetail
}

type CardStatusUpdateResult struct {
	Code    int64
	Message string
	Success bool
	Record  *CardStatusUpdateRecord
}

type CardStatusUpdateRecord struct {
	ID         string `json:"id,omitempty"`
	RequestID  string `json:"requestId,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
	UpdateTime string `json:"updateTime,omitempty"`
}

type RequestResult struct {
	ID         string `json:"id,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
	UpdateTime string `json:"updateTime,omitempty"`
	RequestID  string `json:"requestId,omitempty"`
	Result     string `json:"result,omitempty"`
}

type CardData struct {
	Pages                  int64         `json:"pages,omitempty"`
	Records                []*Card       `json:"records,omitempty"`
	Total                  int64         `json:"total,omitempty"`
	Size                   int64         `json:"size,omitempty"`
	Current                int64         `json:"current,omitempty"`
	Orders                 []interface{} `json:"orders,omitempty"`
	Column                 string        `json:"column,omitempty"`
	Asc                    bool          `json:"asc,omitempty"`
	OptimizeCountSql       bool          `json:"optimizeCountSql,omitempty"`
	SearchCount            bool          `json:"searchCount,omitempty"`
	OptimizeJoinOfCountSql int64         `json:"optimizeJoinOfCountSql,omitempty"`
	CountId                string        `json:"countId,omitempty"`
	Success                bool          `json:"success,omitempty"`
}

type CardQueryResponse struct {
	Code    int64                 `json:"code,omitempty"`
	Message string                `json:"message,omitempty"`
	Data    *CardTransactionsData `json:"data,omitempty"`
	Success bool                  `json:"success,omitempty"`
}
