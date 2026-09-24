package service

import (
	"context"
	"time"

	"generic-mock/channel/paynda/biz"
	paynda "generic-mock/channel/paynda/enums"
	"generic-mock/channel/paynda/pkg/idconv"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
)

type PayndaUIService struct {
	usecase *biz.PayndaUIUsecase
}

func NewPayndaUIService(injector *do.Injector) (*PayndaUIService, error) {
	return &PayndaUIService{
		usecase: do.MustInvoke[*biz.PayndaUIUsecase](injector),
	}, nil
}

type PayndaUIListRequest struct {
	AccountID  *string `form:"account_id"`
	PageNumber *int    `form:"page_number" binding:"omitempty,min=1"`
	PageSize   *int    `form:"page_size" binding:"omitempty,min=1"`
}

type PayndaUIListResponse[T any] struct {
	TotalItems int `json:"total_items"`
	Data       []T `json:"data"`
}

type PayndaUIWebhookData struct {
	AccountName string              `json:"account_name"`
	ID          string              `json:"id"`
	AccountID   string              `json:"account_id"`
	Event       paynda.WebhookEvent `json:"event"`
	TargetURL   string              `json:"target_url"`
	Enabled     bool                `json:"enabled"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type PayndaUIWebhookRecordData struct {
	AccountName     string     `json:"account_name"`
	ID              string     `json:"id"`
	AccountID       string     `json:"account_id"`
	Event           string     `json:"event"`
	TargetURL       string     `json:"target_url"`
	SourceID        string     `json:"source_id"`
	Payload         string     `json:"payload"`
	RequestHeaders  string     `json:"request_headers"`
	ResponseBody    string     `json:"response_body"`
	ResponseHeaders string     `json:"response_headers"`
	StatusCode      int        `json:"status_code"`
	Status          string     `json:"status"`
	AttemptCount    int        `json:"attempt_count"`
	DeliveredAt     *time.Time `json:"delivered_at"`
	ErrorMessage    string     `json:"error_message"`
	CreatedAt       time.Time  `json:"created_at"`
}

type PayndaUIAccountData struct {
	Balance   decimal.Decimal `json:"balance"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	WalletID  string          `json:"wallet_id"`
	CreatedAt time.Time       `json:"created_at"`
}

type PayndaUICreateAccountRequest struct {
	Name string `json:"name" binding:"required"`
}

func (s *PayndaUIService) CreateAccount(
	ctx context.Context,
	req *PayndaUICreateAccountRequest,
) (*PayndaUIAccountData, error) {
	item, err := s.usecase.CreateAccount(ctx, &biz.PayndaUICreateAccountRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}

	return payndaUIAccountData(item.Account), nil
}

func (s *PayndaUIService) ListAccounts(
	ctx context.Context,
	req *PayndaUIListRequest,
) (*PayndaUIListResponse[*PayndaUIAccountData], error) {
	listRequest, err := payndaUIListRequest(req)
	if err != nil {
		return nil, err
	}
	items, total, err := s.usecase.ListAccounts(ctx, listRequest)
	if err != nil {
		return nil, err
	}

	return &PayndaUIListResponse[*PayndaUIAccountData]{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, payndaUIAccountData),
	}, nil
}

type PayndaUIUpdateAccountRequest struct {
	ID   string `uri:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (s *PayndaUIService) UpdateAccount(
	ctx context.Context,
	req *PayndaUIUpdateAccountRequest,
) (*PayndaUIAccountData, error) {
	id, err := idconv.FromAccountString(req.ID)
	if err != nil {
		return nil, err
	}

	item, err := s.usecase.UpdateAccount(ctx, &biz.PayndaUIUpdateAccountRequest{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}

	return payndaUIAccountData(item), nil
}

func (s *PayndaUIService) ListWebhookRecords(
	ctx context.Context,
	req *PayndaUIListRequest,
) (*PayndaUIListResponse[*PayndaUIWebhookRecordData], error) {
	listRequest, err := payndaUIListRequest(req)
	if err != nil {
		return nil, err
	}
	items, total, err := s.usecase.ListWebhookRecords(ctx, listRequest)
	if err != nil {
		return nil, err
	}

	return &PayndaUIListResponse[*PayndaUIWebhookRecordData]{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, payndaUIWebhookRecordData),
	}, nil
}

func (s *PayndaUIService) ReplayWebhookRecord(
	ctx context.Context,
	req *struct {
		ID string `uri:"id" binding:"required"`
	},
) (*PayndaUIWebhookRecordData, error) {
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}

	item, err := s.usecase.ReplayWebhookRecord(ctx, id)
	if err != nil {
		return nil, err
	}

	return payndaUIWebhookRecordData(item), nil
}

type PayndaUICreateWebhookRequest struct {
	AccountID string              `json:"account_id" binding:"required"`
	Event     paynda.WebhookEvent `json:"event" binding:"required"`
	TargetURL string              `json:"target_url" binding:"required,url"`
	Enabled   bool                `json:"enabled"`
}
type PayndaUIUpdateWebhookRequest struct {
	ID        string `uri:"id" binding:"required"`
	TargetURL string `json:"target_url" binding:"required,url"`
	Enabled   bool   `json:"enabled"`
}

func (s *PayndaUIService) CreateWebhook(ctx context.Context, req *PayndaUICreateWebhookRequest) (*PayndaUIWebhookData, error) {
	accountID, err := idconv.FromString(req.AccountID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.CreateWebhook(ctx, &biz.PayndaUICreateWebhookRequest{
		AccountID: accountID,
		Event:     req.Event,
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return payndaUIWebhookData(item), nil
}
func (s *PayndaUIService) ListWebhooks(ctx context.Context, req *struct {
	AccountID *string `form:"account_id"`
}) (*[]PayndaUIWebhookData, error) {
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListWebhooks(ctx, &biz.PayndaListWebhooksRequest{AccountID: accountID})
	if err != nil {
		return nil, err
	}
	result := make([]PayndaUIWebhookData, 0, len(items))
	for _, item := range items {
		result = append(result, *payndaUIWebhookData(item))
	}
	return &result, nil
}

func (s *PayndaUIService) ListWebhookEvents(context.Context, *struct{}) (*[]paynda.WebhookEvent, error) {
	events := paynda.WebhookEvents()
	return &events, nil
}
func (s *PayndaUIService) UpdateWebhook(ctx context.Context, req *PayndaUIUpdateWebhookRequest) (*PayndaUIWebhookData, error) {
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateWebhook(ctx, &biz.PayndaUIUpdateWebhookRequest{
		ID:        id,
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return payndaUIWebhookData(item), nil
}
func (s *PayndaUIService) DeleteWebhook(ctx context.Context, req *struct {
	ID string `uri:"id" binding:"required"`
}) (*struct{}, error) {
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	if err := s.usecase.DeleteWebhook(ctx, id); err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

type PayndaUICardHolderRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Mobile    string `json:"phone_number" binding:"required"`
}

type PayndaUICardHolderData struct {
	AccountName string                  `json:"account_name"`
	AccountID   string                  `json:"account_id"`
	ID          string                  `json:"id"`
	FirstName   string                  `json:"first_name"`
	LastName    string                  `json:"last_name"`
	Email       string                  `json:"email"`
	Mobile      string                  `json:"phone_number"`
	Status      paynda.CardHolderStatus `json:"status"`
	CreatedAt   time.Time               `json:"created_at"`
}
type PayndaUICreateCardRequest struct {
	AccountID    string          `json:"account_id" binding:"required"`
	CardHolderID string          `json:"cardholder_id" binding:"required"`
	CardCurrency common.Currency `json:"card_currency" binding:"required"`
}
type PayndaUICardData struct {
	AccountName   string            `json:"account_name"`
	AccountID     string            `json:"account_id"`
	WalletID      string            `json:"wallet_id"`
	ID            string            `json:"id"`
	CardHolderID  string            `json:"cardholder_id"`
	CardNumber    string            `json:"card_number"`
	CardBin       string            `json:"card_bin"`
	CardCurrency  common.Currency   `json:"card_currency"`
	CardStatus    paynda.CardStatus `json:"card_status"`
	Cvv           string            `json:"cvv"`
	ExpiresAt     time.Time         `json:"expires_at"`
	CreatedAt     time.Time         `json:"created_at"`
	FundingSource string            `json:"funding_source"`
	Balance       string            `json:"balance"`
}
type PayndaUIUpdateCardStatusRequest struct {
	ID         string            `uri:"id" binding:"required"`
	CardStatus paynda.CardStatus `json:"card_status" binding:"required"`
}

func (s *PayndaUIService) CreateCardHolder(ctx context.Context, req *PayndaUICardHolderRequest) (*PayndaUICardHolderData, error) {
	accountID, err := idconv.FromAccountString(req.AccountID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.CreateCardHolder(ctx, &biz.PayndaUICreateCardHolderRequest{
		AccountID: accountID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Mobile:    req.Mobile,
	})
	if err != nil {
		return nil, err
	}
	return payndaUICardHolderData(item), nil
}

func (s *PayndaUIService) ListCards(ctx context.Context, req *PayndaUIListRequest) (*PayndaUIListResponse[*PayndaUICardData], error) {
	listRequest, err := payndaUIListRequest(req)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListCards(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	return &PayndaUIListResponse[*PayndaUICardData]{
		TotalItems: len(items),
		Data:       types.BulkConvertSlice(items, payndaUICardData),
	}, nil
}

func (s *PayndaUIService) ListAuthorizations(ctx context.Context, req *PayndaUIListRequest) (*PayndaUIListResponse[*PayndaUIAuthorizationData], error) {
	listRequest, err := payndaUIListRequest(req)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListAuthorizations(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	return &PayndaUIListResponse[*PayndaUIAuthorizationData]{
		TotalItems: len(items),
		Data:       types.BulkConvertSlice(items, payndaUIAuthorizationData),
	}, nil
}

func (s *PayndaUIService) ListCardHolders(ctx context.Context, req *PayndaUIListRequest) (*PayndaUIListResponse[*PayndaUICardHolderData], error) {
	listRequest, err := payndaUIListRequest(req)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListCardHolders(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	return &PayndaUIListResponse[*PayndaUICardHolderData]{
		TotalItems: len(items),
		Data:       types.BulkConvertSlice(items, payndaUICardHolderData),
	}, nil
}

func (s *PayndaUIService) CreateCard(ctx context.Context, req *PayndaUICreateCardRequest) (*PayndaUICardData, error) {
	accountID, err := idconv.FromAccountString(req.AccountID)
	if err != nil {
		return nil, err
	}
	holderID, err := idconv.FromString(req.CardHolderID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.CreateCard(ctx, &biz.PayndaUICreateCardRequest{
		AccountID:    accountID,
		CardHolderID: holderID,
		Currency:     req.CardCurrency,
	})
	if err != nil {
		return nil, err
	}
	return payndaUICardData(item), nil
}

func (s *PayndaUIService) UpdateCardStatus(ctx context.Context, req *PayndaUIUpdateCardStatusRequest) (*PayndaUICardData, error) {
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateCardStatus(ctx, &biz.PayndaUIUpdateCardStatusRequest{
		CardID: id,
		Status: paynda.CardStatusToGeneric(req.CardStatus),
	})
	if err != nil {
		return nil, err
	}
	return payndaUICardData(item), nil
}

type PayndaUISimulateAuthorizationRequest struct {
	CardID          string          `json:"card_id" binding:"required"`
	Amount          decimal.Decimal `json:"transaction_amount" binding:"required"`
	Currency        common.Currency `json:"transaction_currency" binding:"required"`
	MerchantName    string          `json:"merchant_name" binding:"required"`
	MerchantCountry string          `json:"merchant_country" binding:"required"`
	MerchantCity    string          `json:"merchant_city"` // Invalid: generic model has no merchant city field.
	MerchantMCC     string          `json:"merchant_category_code"`
}

type PayndaUISimulateAuthorizationData struct {
	Authorization   *PayndaUIAuthorizationData `json:"authorization"`
	CardTransaction *PayndaUITransactionData   `json:"transaction"`
}

type PayndaUISimulateRefundRequest struct {
	AuthorizationID      *string         `json:"authorization_id"`
	CardID               string          `json:"card_id" binding:"required"`
	Amount               decimal.Decimal `json:"amount" binding:"required"`
	Currency             common.Currency `json:"currency" binding:"required"`
	MerchantName         string          `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string          `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string          `json:"merchant_country" binding:"required"`
	MerchantCity         string          `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

func (s *PayndaUIService) SimulateRefund(ctx context.Context, req *PayndaUISimulateRefundRequest) (*PayndaUITransactionData, error) {
	authorizationID, err := idconv.FromRefundAuthorizationString(req.AuthorizationID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.SimulateRefund(ctx, &biz.PayndaSimulateRefundRequest{
		AuthorizationID: authorizationID,
		CardID:          cardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}
	return payndaUITransactionData(item), nil
}

type PayndaUIAuthorizationData struct {
	AccountID            string                   `json:"account_id"`
	AccountName          string                   `json:"account_name"`
	ID                   string                   `json:"id"`
	CardID               string                   `json:"card_id"`
	Status               paynda.TransactionStatus `json:"status"`
	AuthorizedAmount     string                   `json:"authorized_amount"`
	Currency             common.Currency          `json:"currency"`
	MerchantName         string                   `json:"merchant_name"`
	MerchantCategoryCode string                   `json:"merchant_category_code"`
	AuthorizedAt         time.Time                `json:"authorized_at"`
}
type PayndaUITransactionData struct {
	AccountName          string                   `json:"account_name"`
	AccountID            string                   `json:"account_id"`
	ID                   string                   `json:"id"`
	CardID               string                   `json:"card_id"`
	AuthorizationID      string                   `json:"authorization_id"`
	TransactionType      paynda.TransactionType   `json:"transaction_type"`
	Status               paynda.TransactionStatus `json:"status"`
	Amount               string                   `json:"amount"`
	Currency             common.Currency          `json:"currency"`
	MerchantName         string                   `json:"merchant_name"`
	MerchantCategoryCode string                   `json:"merchant_category_code"`
	TransactedAt         time.Time                `json:"transacted_at"`
}

func (s *PayndaUIService) SimulateAuthorization(
	ctx context.Context,
	req *PayndaUISimulateAuthorizationRequest,
) (*PayndaUISimulateAuthorizationData, error) {
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	result, err := s.usecase.SimulateAuthorization(ctx, &biz.PayndaSimulateAuthorizationRequest{
		CardID:          cardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
	})
	if err != nil {
		return nil, err
	}

	return &PayndaUISimulateAuthorizationData{
		Authorization:   payndaUIAuthorizationData(result.Authorization),
		CardTransaction: payndaUITransactionData(result.CardTransaction),
	}, nil
}

type PayndaUIApplyTransactionStepRequest struct {
	ID     string           `uri:"id" binding:"required"`
	Amount *decimal.Decimal `json:"amount"`
}

func (s *PayndaUIService) ClearTransaction(ctx context.Context, req *PayndaUIApplyTransactionStepRequest) (*PayndaUITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_CLEAR)
}
func (s *PayndaUIService) ReverseTransaction(ctx context.Context, req *PayndaUIApplyTransactionStepRequest) (*PayndaUITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_VOID)
}
func (s *PayndaUIService) RefundTransaction(ctx context.Context, req *PayndaUIApplyTransactionStepRequest) (*PayndaUITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_REFUND)
}
func (s *PayndaUIService) ListTransactions(ctx context.Context, req *PayndaUIListRequest) (*PayndaUIListResponse[*PayndaUITransactionData], error) {
	listRequest, err := payndaUIListRequest(req)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListTransactions(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	return &PayndaUIListResponse[*PayndaUITransactionData]{
		TotalItems: len(items),
		Data:       types.BulkConvertSlice(items, payndaUITransactionData),
	}, nil
}
func (s *PayndaUIService) applyTransactionStep(
	ctx context.Context,
	req *PayndaUIApplyTransactionStepRequest,
	kind common.CardTransactionType,
) (*PayndaUITransactionData, error) {
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.ApplyTransactionStep(ctx, &biz.PayndaUIApplyTransactionStepRequest{
		CardTransactionID: id,
		Type:              kind,
		Amount:            req.Amount,
	})
	if err != nil {
		return nil, err
	}
	return payndaUITransactionData(item), nil
}
func payndaUIListRequest(req *PayndaUIListRequest) (*biz.PayndaListRequest, error) {
	page, size := types.NormalizePagination(types.Value(req.PageNumber), types.Value(req.PageSize))
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
	}
	return &biz.PayndaListRequest{
		AccountID: accountID,
		Offset:    (page - 1) * size,
		Limit:     size,
	}, nil
}
func payndaUICardHolderData(item *model.CardHolder) *PayndaUICardHolderData {
	return &PayndaUICardHolderData{
		AccountID:   idconv.ToString(item.AccountID),
		AccountName: uiAccountName(item.Account),
		ID:          idconv.ToString(item.ID),
		FirstName:   item.FirstName,
		LastName:    item.LastName,
		Email:       item.Email,
		Mobile:      item.Mobile,
		Status:      paynda.CardHolderStatusFromGeneric(item.Status),
		CreatedAt:   item.CreatedAt,
	}
}
func payndaUICardData(item *model.Card) *PayndaUICardData {
	balance := decimal.Zero
	fundingSource := "卡资金"
	if item.Wallet != nil {
		balance = item.Wallet.Amount
	}

	return &PayndaUICardData{
		AccountID:     idconv.ToString(item.AccountID),
		AccountName:   uiAccountName(item.Account),
		WalletID:      idconv.ToString(item.WalletID),
		ID:            idconv.ToString(item.ID),
		CardHolderID:  idconv.ToString(item.CardHolderID),
		CardNumber:    item.CardNumber,
		CardBin:       item.CardBin,
		CardCurrency:  item.CardCurrency,
		CardStatus:    paynda.CardStatusFromGeneric(item.Status),
		Cvv:           item.Cvv,
		ExpiresAt:     item.ExpireAt,
		CreatedAt:     item.CreatedAt,
		FundingSource: fundingSource,
		Balance:       balance.String(),
	}
}
func payndaUIAuthorizationData(item *model.Authorization) *PayndaUIAuthorizationData {
	return &PayndaUIAuthorizationData{
		AccountID:            idconv.ToString(item.AccountID),
		AccountName:          uiAccountName(item.Account),
		ID:                   idconv.ToString(item.ID),
		CardID:               idconv.ToString(item.CardID),
		Status:               paynda.TransactionStatusFromGeneric(item.Status),
		AuthorizedAmount:     item.Amount.String(),
		Currency:             item.Currency,
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		AuthorizedAt:         item.CreatedAt,
	}
}
func payndaUIWebhookData(item *model.WebhookConfig) *PayndaUIWebhookData {
	return &PayndaUIWebhookData{
		ID:          idconv.ToString(item.ID),
		AccountID:   idconv.ToString(item.AccountID),
		AccountName: uiAccountName(item.Account),
		Event:       paynda.WebhookEvent(item.Event),
		TargetURL:   item.TargetURL,
		Enabled:     item.Enabled,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func payndaUIWebhookRecordData(item *model.WebhookRecord) *PayndaUIWebhookRecordData {
	return &PayndaUIWebhookRecordData{
		ID:              idconv.ToString(item.ID),
		AccountID:       idconv.ToString(item.AccountID),
		AccountName:     uiAccountName(item.Account),
		Event:           item.Event,
		TargetURL:       item.TargetURL,
		SourceID:        item.SourceID,
		Payload:         string(item.Payload),
		RequestHeaders:  string(item.RequestHeaders),
		ResponseBody:    item.ResponseBody,
		ResponseHeaders: string(item.ResponseHeaders),
		StatusCode:      item.StatusCode,
		Status:          string(item.Status),
		AttemptCount:    item.AttemptCount,
		DeliveredAt:     item.DeliveredAt,
		ErrorMessage:    item.ErrorMessage,
		CreatedAt:       item.CreatedAt,
	}
}

func payndaUIAccountData(item *model.Account) *PayndaUIAccountData {
	balance := decimal.Zero
	if item.Wallet != nil {
		balance = item.Wallet.Amount
	}
	return &PayndaUIAccountData{
		Balance:   balance,
		ID:        idconv.ToString(item.ID),
		Name:      item.Name,
		WalletID:  idconv.ToString(item.WalletID),
		CreatedAt: item.CreatedAt,
	}
}
func payndaUITransactionData(item *model.CardTransaction) *PayndaUITransactionData {
	return &PayndaUITransactionData{
		AccountID:            idconv.ToString(item.AccountID),
		AccountName:          uiAccountName(item.Account),
		ID:                   idconv.ToString(item.ID),
		CardID:               idconv.ToString(item.CardID),
		AuthorizationID:      idconv.ToString(item.AuthorizationID),
		TransactionType:      paynda.TransactionTypeFromGeneric(item.Type),
		Status:               paynda.TransactionStatusFromGeneric(item.Status),
		Amount:               item.TxAmount.String(),
		Currency:             item.TxCurrency,
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		TransactedAt:         item.CreatedAt,
	}
}

type ManagementListRequest struct {
	AccountID *string `form:"account_id"`
}

type ManagementAccountRequest struct {
	AccountID string `form:"account_id" json:"account_id" binding:"required"`
}
type FundsData struct {
	AccountName string            `json:"account_name"`
	AccountID   string            `json:"account_id"`
	ID          string            `json:"id"`
	Currency    common.Currency   `json:"currency"`
	Kind        paynda.WalletKind `json:"kind"`
	Amount      string            `json:"amount"`
}

func (s *PayndaUIService) ListFunds(ctx context.Context, req *ManagementListRequest) (*[]FundsData, error) {
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListFunds(ctx, accountID)
	if err != nil {
		return nil, err
	}
	result := make([]FundsData, 0, len(items))
	for _, item := range items {
		var kind paynda.WalletKind
		switch item.Type {
		case common.WalletType_Account:
			kind = paynda.WalletKindAccount
		case common.WalletType_Card:
			kind = paynda.WalletKindCard
		default:
			continue
		}
		result = append(result, FundsData{
			AccountID:   idconv.ToString(item.AccountID),
			AccountName: uiAccountName(item.Account),
			ID:          idconv.ToString(item.ID),
			Currency:    item.Currency,
			Kind:        kind,
			Amount:      item.Amount.String(),
		})
	}
	return &result, nil
}

type MoveFundsRequest struct {
	ManagementAccountRequest
	SourceID string          `json:"source_id"`
	TargetID string          `json:"target_id"`
	Amount   decimal.Decimal `json:"amount"`
}

func (s *PayndaUIService) MoveFunds(ctx context.Context, req *MoveFundsRequest) (*struct{}, error) {
	accountID, err := idconv.FromAccountString(req.AccountID)
	if err != nil {
		return nil, err
	}
	var sourceID, targetID int64
	if req.SourceID != "" {
		sourceID, err = idconv.FromString(req.SourceID)
		if err != nil {
			return nil, err
		}
	}
	if req.TargetID != "" {
		targetID, err = idconv.FromString(req.TargetID)
		if err != nil {
			return nil, err
		}
	}
	if err := s.usecase.MoveFunds(ctx, &biz.MoveFundsRequest{
		AccountID: accountID,
		SourceID:  sourceID,
		TargetID:  targetID,
		Amount:    req.Amount,
	}); err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

type AuthorizationBalanceData struct {
	AccountName  string          `json:"account_name"`
	AccountID    string          `json:"account_id"`
	ID           string          `json:"id"`
	CardID       string          `json:"card_id"`
	Currency     common.Currency `json:"currency"`
	Amount       string          `json:"amount"`
	Settled      string          `json:"settled"`
	Remaining    string          `json:"remaining"`
	MerchantName string          `json:"merchant_name"`
	CreatedAt    time.Time       `json:"created_at"`
}

func (s *PayndaUIService) ListAuthorizationBalances(ctx context.Context, req *ManagementListRequest) (*[]AuthorizationBalanceData, error) {
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListAuthorizationBalances(ctx, accountID)
	if err != nil {
		return nil, err
	}
	result := make([]AuthorizationBalanceData, 0, len(items))
	for _, item := range items {
		auth := item.Authorization
		result = append(result, AuthorizationBalanceData{
			AccountID:    idconv.ToString(auth.AccountID),
			AccountName:  uiAccountName(auth.Account),
			ID:           idconv.ToString(auth.ID),
			CardID:       idconv.ToString(auth.CardID),
			Currency:     auth.Currency,
			Amount:       auth.Amount.String(),
			Settled:      item.Settled.String(),
			Remaining:    item.Remaining.String(),
			MerchantName: auth.MerchantName,
			CreatedAt:    auth.CreatedAt,
		})
	}
	return &result, nil
}

type ClearAuthorizationRequest struct {
	ManagementAccountRequest
	ID     string          `uri:"id" binding:"required"`
	Amount decimal.Decimal `json:"amount"`
}
type ClearAuthorizationData struct {
	ID string `json:"id"`
}

func (s *PayndaUIService) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*ClearAuthorizationData, error) {
	accountID, err := idconv.FromAccountString(req.AccountID)
	if err != nil {
		return nil, err
	}
	authID, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.ClearAuthorization(ctx, &biz.ClearAuthorizationRequest{
		AccountID: accountID,
		ID:        authID,
		Amount:    req.Amount,
	})
	if err != nil {
		return nil, err
	}

	return &ClearAuthorizationData{ID: idconv.ToString(item.ID)}, nil
}

func uiAccountName(account *model.Account) string {
	if account == nil {
		return ""
	}
	return account.Name
}
