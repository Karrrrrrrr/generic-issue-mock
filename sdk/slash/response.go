package slash

import (
	"time"
)

// metadata
type Metadata struct {
	NextCursor string `json:"nextCursor"`
	Count      int    `json:"count"`
}

// Account
type Account struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"` // open, closed
	Name          string    `json:"name"`
	AccountNumber string    `json:"accountNumber"`
	RoutingNumber string    `json:"routingNumber"`
	CreatedAt     time.Time `json:"createdAt"`
	Type          string    `json:"type"` // debit, charge_card
	Balances      []string  `json:"balances"`
}

// VirtualAccount
type VirtualAccount struct {
	VirtualAccount VirtualAccountDetails `json:"virtualAccount"`
	Balance        Amount                `json:"balance"`
	Spend          Amount                `json:"spend"`
	CommissionRule CommissionRule        `json:"commissionRule"`
}

// VirtualAccountDetails
type VirtualAccountDetails struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	AccountNumber string `json:"accountNumber"`
	RoutingNumber string `json:"routingNumber"`
	AccountID     string `json:"accountId"`
	ClosedAt      string `json:"closedAt"`
	AccountType   string `json:"accountType"` // default, primary
}

// Amount
type Amount struct {
	AmountCents int `json:"amountCents"`
}

// CommissionRule
type CommissionRule struct {
	ID                string            `json:"id"`
	VirtualAccountID  string            `json:"virtualAccountId"`
	CommissionDetails CommissionDetails `json:"commissionDetails"`
}

// CommissionDetails
type CommissionDetails struct {
	Type      string    `json:"type"` // flateFee, takeRate
	Amount    Amount    `json:"amount"`
	Frequency string    `json:"frequency"` // monthly, yearly
	StartDate time.Time `json:"startDate"`
}

// Balance
type Balance struct {
	AccountID string    `json:"accountId"`
	Type      string    `json:"type"` // cash,credit,debit
	Available Amount    `json:"available"`
	Posted    Amount    `json:"posted"`
	Timestamp time.Time `json:"timestamp"`
}

// Card
type Card struct {
	ID                 string             `json:"id"`
	AccountID          string             `json:"accountId"`
	VirtualAccountID   string             `json:"virtualAccountId"`
	Last4              string             `json:"last4"`
	Name               string             `json:"name"`
	ExpiryMonth        string             `json:"expiryMonth"`
	ExpiryYear         string             `json:"expiryYear"`
	Status             string             `json:"status"`
	IsPhysical         bool               `json:"isPhysical"`
	IsSingleUse        bool               `json:"isSingleUse"`
	Pan                string             `json:"pan"`
	Cvv                string             `json:"cvv"`
	CardGroupID        string             `json:"cardGroupId"`
	CreatedAt          time.Time          `json:"createdAt"`
	SpendingConstraint SpendingConstraint `json:"spendingConstraint"`
	UserData           UserData           `json:"userData"`
	CardProductID      string             `json:"cardProductId"`
}

type ListCardsResponse struct {
	Items    []Card   `json:"items"`
	Metadata Metadata `json:"metadata"`
}

// CardModifier
type CardModifier struct {
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

type CardModifiersResponse struct {
	Modifiers []CardModifier `json:"modifiers"`
}

// Transaction
type Transaction struct {
	ID                      string           `json:"id"`
	Date                    string           `json:"date"`
	Description             string           `json:"description"`
	Memo                    string           `json:"memo"`
	MerchantDescription     string           `json:"merchantDescription"`
	MerchantData            MerchantData     `json:"merchantData"`
	AmountCents             int              `json:"amountCents"`
	Status                  string           `json:"status"`
	DetailedStatus          string           `json:"detailedStatus"`
	AccountID               string           `json:"accountId"`
	VirtualAccountID        string           `json:"virtualAccountId"`
	AccountSubtype          string           `json:"accountSubtype"`
	CardID                  string           `json:"cardId"`
	OriginalCurrency        OriginalCurrency `json:"originalCurrency"`
	OrderID                 string           `json:"orderId"`
	ReferenceNumber         string           `json:"referenceNumber"`
	AuthorizedAt            string           `json:"authorizedAt"`
	DeclineReason           string           `json:"declineReason"`
	ApprovalReason          string           `json:"approvalReason"`
	ProviderAuthorizationID string           `json:"providerAuthorizationId"`
	WireInfo                *WireInfo        `json:"wireInfo,omitempty"`
	AchInfo                 *AchInfo         `json:"achInfo,omitempty"`
	RtpInfo                 *RtpInfo         `json:"rtpInfo,omitempty"`
}
type TransactionsResponse struct {
	Items    []Transaction `json:"items"`
	Metadata Metadata      `json:"metadata"`
}

// MerchantData
type MerchantData struct {
	Description  string   `json:"description"`
	CategoryCode string   `json:"categoryCode"`
	Location     Location `json:"location"`
}

// Location
type Location struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
	Zip     string `json:"zip"`
}

// OriginalCurrency
type OriginalCurrency struct {
	Code           string  `json:"code"`
	AmountCents    int     `json:"amountCents"`
	ConversionRate float64 `json:"conversionRate"`
}

// WireInfo
type WireInfo struct {
	CounterpartyBank     string `json:"counterpartyBank"`
	TypeCode             string `json:"typeCode"`
	SubtypeCode          string `json:"subtypeCode"`
	Omad                 string `json:"omad"`
	Imad                 string `json:"imad"`
	SenderReference      string `json:"senderReference"`
	BusinessFunctionCode string `json:"businessFunctionCode"`
}

// AchInfo
type AchInfo struct {
	CounterpartyBank         string `json:"counterpartyBank"`
	ReceiverID               string `json:"receiverId"`
	CompanyID                string `json:"companyId"`
	CompanyDiscretionaryData string `json:"companyDiscretionaryData"`
	CompanyEntryDescription  string `json:"companyEntryDescription"`
	TraceNumber              string `json:"traceNumber"`
	EntryClassCode           string `json:"entryClassCode"`
	PaymentRelatedInfo       string `json:"paymentRelatedInfo"`
}

// RtpInfo
type RtpInfo struct {
	CounterpartyBank string `json:"counterpartyBank"`
	EndToEndID       string `json:"endToEndId"`
	RoutingNumber    string `json:"routingNumber"`
	OriginatorName   string `json:"originatorName"`
	Description      string `json:"description"`
}

// SpendingConstraint
type SpendingConstraint struct {
	MerchantCategoryRule     *MerchantCategoryRule     `json:"merchantCategoryRule,omitempty"`
	MerchantRule             *MerchantRule             `json:"merchantRule,omitempty"`
	SpendingRule             *SpendingRule             `json:"spendingRule,omitempty"`
	CountryRule              *CountryRule              `json:"countryRule,omitempty"`
	MerchantCategoryCodeRule *MerchantCategoryCodeRule `json:"merchantCategoryCodeRule,omitempty"`
}

// MerchantCategoryRule
type MerchantCategoryRule struct {
	MerchantCategories []string `json:"merchantCategories"`
	Restriction        string   `json:"restriction"`
}

// MerchantRule
type MerchantRule struct {
	Merchants   []string `json:"merchants"`
	Restriction string   `json:"restriction"`
}

// Utilization
type Utilization struct {
	NextResetDate    string `json:"nextResetDate"`
	Spend            Amount `json:"spend"`
	AvailableBalance Amount `json:"availableBalance"`
}

// SpendingRule
type SpendingRule struct {
	UtilizationLimit     *UtilizationLimit     `json:"utilizationLimit,omitempty"`
	UtilizationLimitV2   []UtilizationLimit    `json:"utilizationLimitV2,omitempty"`
	TransactionSizeLimit *TransactionSizeLimit `json:"transactionSizeLimit,omitempty"`
}

// UtilizationLimit
type UtilizationLimit struct {
	Timezone    string `json:"timezone"`
	LimitAmount Amount `json:"limitAmount"`
	Preset      string `json:"preset"`
	StartDate   string `json:"startDate"`
}

// TransactionSizeLimit
type TransactionSizeLimit struct {
	Minimum Amount `json:"minimum"`
	Maximum Amount `json:"maximum"`
}

// CountryRule
type CountryRule struct {
	Countries   []string `json:"countries"`
	Restriction string   `json:"restriction"`
}

// MerchantCategoryCodeRule
type MerchantCategoryCodeRule struct {
	MerchantCategoryCodes []string `json:"merchantCategoryCodes"`
	Restriction           string   `json:"restriction"`
}

// SlashHandle
type SlashHandle struct {
	ID        string `json:"id"`
	Handle    string `json:"slashHandle"`
	Name      string `json:"name"`
	AccountID string `json:"accountId"`
}

type CardProduct struct {
	ID     string `json:"id"`
	Prefix string `json:"prefix"`
	Status string `json:"status"` //active, inactive
}

// CardGroup
type CardGroup struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	VirtualAccountID   string             `json:"virtualAccountId"`
	SpendingConstraint SpendingConstraint `json:"spendingConstraint"`
	Cards              []string           `json:"cards"`
}

type ListCardGroupsResponse struct {
	Items    []CardGroup `json:"items"`
	Metadata Metadata    `json:"metadata"`
}

// AccountsResponse
type AccountsResponse struct {
	Items    []Account `json:"items"`
	Metadata Metadata  `json:"metadata"`
}

// AccountBalancesResponse
type AccountBalancesResponse struct {
	Balances []Balance `json:"balances"`
}

// CreateVirtualAccountResponse
type VirtualAccountResponse struct {
	VirtualAccount VirtualAccountDetails `json:"virtualAccount"`
	CommissionRule CommissionRule        `json:"commissionRule"`
}

// ListVirtualAccountResponse
type ListVirtualAccountResponse struct {
	Items    []VirtualAccount `json:"items"`
	Metadata Metadata         `json:"metadata"`
}

type CardProductsResponse struct {
	Items    []CardProduct `json:"items"`
	Metadata Metadata      `json:"metadata"`
}

type Merchant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type MerchantsResponse struct {
	Items    []Merchant `json:"items"`
	Metadata Metadata   `json:"metadata"`
}

type MerchantCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type MerchantCategoriesResponse struct {
	Items    []MerchantCategory `json:"items"`
	Metadata Metadata           `json:"metadata"`
}

type Webhook struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	CreateAt string `json:"createAt"`
	UpdateAt string `json:"updateAt"`
}

type WebhooksResponse struct {
	Items    []Webhook `json:"items"`
	Metadata Metadata  `json:"metadata"`
}

// authorization webhook
type AuthWebhook struct {
	WebhookUrl        string            `json:"webhookUrl"`
	SigningSecret     string            `json:"signingSecret"`
	Status            string            `json:"status"`
	TimeoutDurationMs int               `json:"timeoutDurationMs"`
	Config            AuthWebhookConfig `json:"config"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
}

type AuthWebhookConfig struct {
	FallbackBehavior string `json:"fallbackBehavior"`
}

type TransactionAggregation struct {
	Count     int `json:"count"`
	TotalIn   int `json:"totalIn"`
	TotalOut  int `json:"totalOut"`
	NetChange int `json:"netChange"`
}

type Fee struct {
	ID                  string      `json:"id"`
	DateCharged         string      `json:"dateCharged"`
	FeeAmountCents      int         `json:"feeAmountCents"`
	FeeType             string      `json:"feeType"`
	AccountID           string      `json:"accountId"`
	OriginalTransaction Transaction `json:"originalTransaction"`
	Card                Card        `json:"card"`
}

// FeesResponse represents the response containing a list of fees
type FeesResponse struct {
	Items []Fee `json:"items"`
}

type TransferResponse struct {
	TransferID string `json:"transferId"`
}

// LegalEntity
type LegalEntity struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Structure string `json:"structure"`
}

type LegalEntitiesResponse struct {
	Items    []LegalEntity `json:"items"`
	Metadata Metadata      `json:"metadata"`
}
