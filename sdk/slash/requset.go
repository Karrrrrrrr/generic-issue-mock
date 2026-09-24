package slash

import (
	"errors"
)

// 错误定义
var (
	ErrEmptyCardID           = errors.New("cardID不能为空")
	ErrEmptyParams           = errors.New("参数不能为空")
	ErrEmptyResponse         = errors.New("响应数据为空")
	ErrSignatureVerify       = errors.New("签名验证失败")
	ErrNoSignature           = errors.New("响应头中没有签名信息")
	ErrEmptyRequestID        = errors.New("requestID不能为空")
	ErrEmptyCardBinID        = errors.New("cardBinID不能为空")
	ErrEmptyAccountID        = errors.New("accountID不能为空")
	ErrEmptyName             = errors.New("name不能为空")
	ErrEmptyAction           = errors.New("action不能为空")
	ErrEmptySource           = errors.New("source不能为空")
	ErrEmptyDestination      = errors.New("destination不能为空")
	ErrInvalidAmountCents    = errors.New("amountCents必须大于0")
	ErrEmptyVirtualAccountID = errors.New("virtualAccountID不能为空")
	ErrEmptyPaymentRail      = errors.New("paymentRail不能为空")
	ErrEmptyCurrency         = errors.New("currency不能为空")
	ErrEmptyType             = errors.New("type不能为空")
	ErrEmptyStatus           = errors.New("status不能为空")
	ErrEmptyURL              = errors.New("URL不能为空")
	ErrEmptyLegalEntityID    = errors.New("legalEntityID不能为空")
	ErrEmptyWebhookID        = errors.New("webhookID不能为空")
	ErrEmptyReason           = errors.New("reason不能为空")
	ErrEmptyWebhookUrl       = errors.New("webhookUrl不能为空")
	ErrEmptyCardGroupID      = errors.New("cardGroupID不能为空")
	ErrEmptyMerchantID       = errors.New("merchantID不能为空")
	ErrEmptyResult           = errors.New("result is empty")
)

type RestyRequestOptions struct {
	Method    string
	Url       string
	Path      string
	RequestID string
	Params    any
	Result    any
}

// userData
type UserData struct {
	RequestID string `json:"requestId,omitempty"`
	CardID    string `json:"cardId,omitempty"`
}

// Validate 验证UserData的所有字段是否为空
func (u *UserData) Validate() error {
	if u.RequestID == "" {
		return ErrEmptyRequestID
	}
	if u.CardID == "" {
		return ErrEmptyCardID
	}
	return nil
}

// CreateVirtualAccountRequest represents the request to create a virtual account
type CreateVirtualAccountRequest struct {
	AccountID         string            `json:"accountId"`
	Name              string            `json:"name"`
	CommissionDetails CommissionDetails `json:"commissionDetails,omitempty"`
}

// Validate 验证CreateVirtualAccountRequest的所有字段是否为空
func (r *CreateVirtualAccountRequest) Validate() error {
	if r.AccountID == "" {
		return ErrEmptyAccountID
	}
	if r.Name == "" {
		return ErrEmptyName
	}
	return nil
}

// UpdateVirtualAccountRequest represents the request to update a virtual account
type UpdateVirtualAccountRequest struct {
	Action            string             `json:"action"` // update, close
	Name              string             `json:"name,omitempty"`
	CommissionDetails *CommissionDetails `json:"commissionDetails,omitempty"`
}

// Validate 验证UpdateVirtualAccountRequest的所有字段是否为空
func (r *UpdateVirtualAccountRequest) Validate() error {
	if r.Action == "" {
		return ErrEmptyAction
	}
	if r.Name == "" {
		return ErrEmptyName
	}
	return nil
}

// CreateVirtualAccountTransferRequest represents the request to create a virtual account transfer
type CreateVirtualAccountTransferRequest struct {
	RequestID   string `json:"-"` // Headers: X-Idempotency-Key
	Source      string `json:"source"`
	Destination string `json:"destination"`
	AmountCents int    `json:"amountCents"`
}

// Validate 验证CreateVirtualAccountTransferRequest的所有字段是否为空
func (r *CreateVirtualAccountTransferRequest) Validate() error {
	if r.RequestID == "" {
		return ErrEmptyRequestID
	}
	if r.Source == "" {
		return ErrEmptySource
	}
	if r.Destination == "" {
		return ErrEmptyDestination
	}
	if r.AmountCents <= 0 {
		return ErrInvalidAmountCents
	}
	return nil
}

// GetCryptoOfframpAddressesRequest represents the request to get crypto offramp addresses
type GetCryptoOfframpAddressesRequest struct {
	VirtualAccountID string `json:"virtualAccountId"`
	PaymentRail      string `json:"paymentRail"`
	Currency         string `json:"currency"`
}

// Validate 验证GetCryptoOfframpAddressesRequest的所有字段是否为空
func (r *GetCryptoOfframpAddressesRequest) Validate() error {
	if r.VirtualAccountID == "" {
		return ErrEmptyVirtualAccountID
	}
	if r.PaymentRail == "" {
		return ErrEmptyPaymentRail
	}
	if r.Currency == "" {
		return ErrEmptyCurrency
	}
	return nil
}

// CreateCardRequest represents the request to create a card
type CreateCardRequest struct {
	AccountID          string              `json:"accountId"`
	VirtualAccountID   string              `json:"virtualAccountId"`
	Type               string              `json:"type"` // default: "virtual"
	Name               string              `json:"name"`
	SpendingConstraint *SpendingConstraint `json:"spendingConstraint,omitempty"`
	IsSingleUse        bool                `json:"isSingleUse"` // false
	UserData           UserData            `json:"userData"`    // exceed 4kb
	CardGroupID        string              `json:"cardGroupId,omitempty"`
	CardProductID      string              `json:"cardProductId,omitempty"`
}

// Validate 验证CreateCardRequest的所有字段是否为空
func (r *CreateCardRequest) Validate() error {
	if r.Type == "" {
		return ErrEmptyType
	}
	if r.Name == "" {
		return ErrEmptyName
	}
	// UserData验证
	if err := r.UserData.Validate(); err != nil {
		return err
	}
	return nil
}

type CreateCardReq struct {
	RequestID string
	CardID    string
	Name      string
	CardBinID string
}

func (r *CreateCardReq) Validate() error {
	if r.RequestID == "" {
		return ErrEmptyRequestID
	}
	if r.Name == "" {
		return ErrEmptyName
	}
	if r.CardBinID == "" {
		return ErrEmptyCardBinID
	}
	return nil
}

// UpdateCardRequest represents the request to update a card
type UpdateCardRequest struct {
	Name               string              `json:"name,omitempty"`
	Status             string              `json:"status,omitempty"` //active, paused, inactive, closed
	CardGroupID        string              `json:"cardGroupId,omitempty"`
	SpendingConstraint *SpendingConstraint `json:"spendingConstraint,omitempty"`
	UserData           *UserData           `json:"userData,omitempty"`
}

func (r *UpdateCardRequest) Validate() error {
	if r.Name == "" {
		return ErrEmptyName
	}
	if r.Status == "" {
		return ErrEmptyStatus
	}
	// UserData验证
	if err := r.UserData.Validate(); err != nil {
		return err
	}
	// CardGroupID可能是可选的，不进行非空校验
	return nil
}

type QueryCardsParams struct {
	LegalEntityId    string `url:"filter:legalEntityId,omitempty"`
	AccountId        string `url:"filter:accountId,omitempty"`
	VirtualAccountId string `url:"filter:virtualAccountId,omitempty"`
	Status           string `url:"filter:status,omitempty"`
	CardGroupId      string `url:"filter:cardGroupId,omitempty"`
	CardGroupName    string `url:"filter:cardGroupName,omitempty"`
	Cursor           string `url:"cursor,omitempty"`
	Sort             string `url:"sort,omitempty"`
	SortDirection    string `url:"sortDirection,omitempty"`
}

func (p *QueryCardsParams) BuildQuery() map[string]string {
	if p == nil {
		return nil
	}
	q := make(map[string]string)
	if p.Cursor != "" {
		q["cursor"] = p.Cursor
	}
	if p.Sort != "" {
		q["sort"] = p.Sort
	}
	if p.SortDirection != "" {
		q["sortDirection"] = p.SortDirection
	}
	if p.LegalEntityId != "" {
		q["filter:legalEntityId"] = p.LegalEntityId
	}
	if p.AccountId != "" {
		q["filter:accountId"] = p.AccountId
	}

	if p.VirtualAccountId != "" {
		q["filter:virtualAccountId"] = p.VirtualAccountId
	}
	if p.Status != "" {
		q["filter:status"] = p.Status
	}
	if p.CardGroupId != "" {
		q["filter:cardGroupId"] = p.CardGroupId
	}
	if p.CardGroupName != "" {
		q["filter:cardGroupName"] = p.CardGroupName
	}
	return q
}

// UpdateCardSpendingConstraintRequest represents the request to update a card's spending constraint
type UpdateCardSpendingConstraintRequest SpendingConstraint

// Validate 验证UpdateCardSpendingConstraintRequest的所有字段是否为空
func (r *UpdateCardSpendingConstraintRequest) Validate() error {
	// TODO: 待完善
	return nil
}

// SetCardSpendingConstraintRequest represents the request to set a card's spending constraint
type SetCardSpendingConstraintRequest SpendingConstraint

// Validate 验证SetCardSpendingConstraintRequest的所有字段是否为空
func (r *SetCardSpendingConstraintRequest) Validate() error {
	// TODO: 待完善
	return nil
}

// SetCardModifierRequest represents the request to set a card modifier
type SetCardModifierRequest struct {
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

// Validate 验证SetCardModifierRequest的所有字段是否为空
func (r *SetCardModifierRequest) Validate() error {
	if r.Name == "" {
		return ErrEmptyName
	}
	return nil
}

// CreateCardGroupRequest represents the request to create a card group
type CreateCardGroupRequest struct {
	Name               string             `json:"name"`
	VirtualAccountID   string             `json:"virtualAccountId"`
	SpendingConstraint SpendingConstraint `json:"spendingConstraint"`
}

// Validate 验证CreateCardGroupRequest的所有字段是否为空
func (r *CreateCardGroupRequest) Validate() error {
	if r.Name == "" {
		return ErrEmptyName
	}
	// TODO: 待完善
	// SpendingConstraint是一个结构体，可能需要进一步验证
	return nil
}

// UpdateCardGroupRequest represents the request to update a card group
type UpdateCardGroupRequest struct {
	Name               string             `json:"name"`
	SpendingConstraint SpendingConstraint `json:"spendingConstraint"`
}

// Validate 验证UpdateCardGroupRequest的所有字段是否为空
func (r *UpdateCardGroupRequest) Validate() error {
	if r.Name == "" {
		return ErrEmptyName
	}
	// TODO: 待完善
	// SpendingConstraint是一个结构体，可能需要进一步验证
	return nil
}

// UpdateCardGroupSpendingConstraintRequest represents the request to update a card group's spending constraint
type UpdateCardGroupSpendingConstraintRequest SpendingConstraint

// Validate 验证UpdateCardGroupSpendingConstraintRequest的所有字段是否为空
func (r *UpdateCardGroupSpendingConstraintRequest) Validate() error {
	// TODO: 待完善
	return nil
}

// SetCardGroupSpendingConstraintRequest represents the request to set a card group's spending constraint
type SetCardGroupSpendingConstraintRequest SpendingConstraint

// Validate 验证SetCardGroupSpendingConstraintRequest的所有字段是否为空
func (r *SetCardGroupSpendingConstraintRequest) Validate() error {
	// TODO: 待完善
	return nil
}

// Create Webhook
type CreateWebhookRequest struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Validate 验证CreateWebhookRequest的所有字段是否为空
func (r *CreateWebhookRequest) Validate() error {
	if r.URL == "" {
		return ErrEmptyURL
	}
	if r.Name == "" {
		return ErrEmptyName
	}
	return nil
}

type UpdateWebhookRequest struct {
	WebhookID string `json:"-"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
}

// Validate 验证UpdateWebhookRequest的所有字段是否为空
func (r *UpdateWebhookRequest) Validate() error {
	if r.WebhookID == "" {
		return ErrEmptyWebhookID
	}
	if r.Status == "" {
		return ErrEmptyStatus
	}
	if r.Reason == "" {
		return ErrEmptyReason
	}
	return nil
}

type UpdateAuthWebhookRequest struct {
	WebhookUrl string            `json:"webhookUrl"`
	Status     string            `json:"status"`
	Config     AuthWebhookConfig `json:"config"`
}

// Validate 验证UpdateAuthWebhookRequest的所有字段是否为空
func (r *UpdateAuthWebhookRequest) Validate() error {
	if r.WebhookUrl == "" {
		return ErrEmptyWebhookUrl
	}
	if r.Status == "" {
		return ErrEmptyStatus
	}
	// TODO: 待完善
	// Config是一个结构体，可能需要进一步验证
	return nil
}

type QueryTransactionsParams struct {
	Cursor                  string `json:"cursor,omitempty"`
	AccountId               string `json:"accountId,omitempty"`
	FilterLegalEntityId     string `json:"filter:legalEntityId,omitempty"`
	FilterAccountId         string `json:"filter:accountId,omitempty"`
	FilterVirtualAccountId  string `json:"filter:virtualAccountId,omitempty"`
	FilterStatus            string `json:"filter:status,omitempty"` // pending, posted, failed
	FilterDetailedStatus    string `json:"filter:detailed_status,omitempty"`
	FilterCardId            string `json:"filter:cardId,omitempty"`
	ProviderAuthorizationId string `json:"filter:providerAuthorizationId,omitempty"`
	FromDate                string `json:"filter:from_date,omitempty"`          // unix timestamp in milliseconds
	ToDate                  string `json:"filter:to_date,omitempty"`            // unix timestamp in milliseconds
	FromAuthorizedAt        string `json:"filter:from_authorized_at,omitempty"` // unix timestamp in milliseconds
	ToAuthorizedAt          string `json:"filter:to_authorized_at,omitempty"`   // unix timestamp in milliseconds
}

type QueryTransactionAggregationsParams struct {
	AccountId              string `json:"accountId"`
	FilterLegalEntityId    string `json:"filter:legalEntityId,omitempty"`
	FilterAccountId        string `json:"filter:accountId,omitempty"`
	FilterVirtualAccountId string `json:"filter:virtualAccountId,omitempty"`
	FilterStatus           string `json:"filter:status,omitempty"`
	FilterCardId           string `json:"filter:cardId,omitempty"`
}

type TransferRequest struct {
	RequestID   string `json:"-"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	AmountCents int    `json:"amountCents"`
}

// Validate 验证TransferRequest的所有字段是否为空
func (r *TransferRequest) Validate() error {
	if r.RequestID == "" {
		return ErrEmptyRequestID
	}
	if r.Source == "" {
		return ErrEmptySource
	}
	if r.Destination == "" {
		return ErrEmptyDestination
	}
	if r.AmountCents <= 0 {
		return ErrInvalidAmountCents
	}
	return nil
}
