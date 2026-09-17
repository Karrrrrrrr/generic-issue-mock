package service

import (
	"context"
	"time"

	"generic-mock/channel/photonpay/biz"
	photon "generic-mock/channel/photonpay/enums"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
)

type UIListRequest struct {
	PageNumber int `form:"page_number"`
	PageSize   int `form:"page_size"`
}

type UIListResponse[T any] struct {
	TotalItems int `json:"total_items"`
	Data       []T `json:"data"`
}

type UIAccountData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UIVirtualAccountData struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	Balance   string    `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type UICreateVirtualAccountRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

func (s *PhotonPayUIService) CreateVirtualAccount(
	ctx context.Context,
	req *UICreateVirtualAccountRequest,
) (*UIVirtualAccountData, error) {
	accountID, err := photonPayAccountID(req.AccountID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.CreateVirtualAccount(ctx, &biz.UICreateVirtualAccountRequest{
		AccountID: accountID,
		Name:      req.Name,
	})
	if err != nil {
		return nil, err
	}
	return photonPayUIVirtualAccountData(item), nil
}

func (s *PhotonPayUIService) ListVirtualAccounts(
	ctx context.Context,
	_ *struct{},
) (*[]UIVirtualAccountData, error) {
	items, err := s.usecase.ListVirtualAccounts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]UIVirtualAccountData, 0, len(items))
	for _, item := range items {
		result = append(result, *photonPayUIVirtualAccountData(item))
	}
	return &result, nil
}

type UICreateAccountRequest struct {
	Name string `json:"name" binding:"required"`
}

type UIWebhookData struct {
	ID        string              `json:"id"`
	AccountID string              `json:"account_id"`
	Event     photon.WebhookEvent `json:"event"`
	TargetURL string              `json:"target_url"`
	Enabled   bool                `json:"enabled"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type UIAuthorizationConfigData struct {
	AccountID     string    `json:"account_id"`
	TargetURL     string    `json:"target_url"`
	Enabled       bool      `json:"enabled"`
	TimeoutMillis int       `json:"timeout_millis"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UIGetAuthorizationConfigRequest struct {
	AccountID string `form:"account_id" binding:"required"`
}

func (s *PhotonPayUIService) GetAuthorizationConfig(
	ctx context.Context,
	req *UIGetAuthorizationConfigRequest,
) (*UIAuthorizationConfigData, error) {
	accountID, err := photonPayAccountID(req.AccountID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetAuthorizationConfig(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return photonPayUIAuthorizationConfigData(item), nil
}

type UIUpdateAuthorizationConfigRequest struct {
	AccountID     string `json:"account_id" binding:"required"`
	TargetURL     string `json:"target_url" binding:"required,url"`
	Enabled       bool   `json:"enabled"`
	TimeoutMillis int    `json:"timeout_millis" binding:"required,min=1"`
}

func (s *PhotonPayUIService) UpdateAuthorizationConfig(
	ctx context.Context,
	req *UIUpdateAuthorizationConfigRequest,
) (*UIAuthorizationConfigData, error) {
	accountID, err := photonPayAccountID(req.AccountID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateAuthorizationConfig(ctx, &biz.UIUpdateAuthorizationConfigRequest{
		AccountID:     accountID,
		TargetURL:     req.TargetURL,
		Enabled:       req.Enabled,
		TimeoutMillis: req.TimeoutMillis,
	})
	if err != nil {
		return nil, err
	}
	return photonPayUIAuthorizationConfigData(item), nil
}

type UIWebhookRecordData struct {
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
type UICreateWebhookRequest struct {
	AccountID string              `json:"account_id" binding:"required"`
	Event     photon.WebhookEvent `json:"event" binding:"required"`
	TargetURL string              `json:"target_url" binding:"required,url"`
	Enabled   bool                `json:"enabled"`
}
type UIUpdateWebhookRequest struct {
	ID        string `uri:"id" binding:"required"`
	TargetURL string `json:"target_url" binding:"required,url"`
	Enabled   bool   `json:"enabled"`
}

func (s *PhotonPayUIService) CreateWebhook(ctx context.Context, req *UICreateWebhookRequest) (*UIWebhookData, error) {
	accountID, err := photonPayAccountID(req.AccountID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.CreateWebhook(ctx, &biz.UICreateWebhookRequest{
		AccountID: accountID,
		Event:     req.Event,
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return photonPayUIWebhookData(item), nil
}
func (s *PhotonPayUIService) ListWebhooks(ctx context.Context, req *struct {
	AccountID string `form:"account_id"`
}) (*[]UIWebhookData, error) {
	accountID := model.ID(0)
	if req.AccountID != "" {
		var err error
		accountID, err = photonPayAccountID(req.AccountID)
		if err != nil {
			return nil, err
		}
	}
	items, err := s.usecase.ListWebhooks(ctx, &biz.WebhookListRequest{
		AccountID: accountID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]UIWebhookData, 0, len(items))
	for _, item := range items {
		result = append(result, *photonPayUIWebhookData(item))
	}
	return &result, nil
}

func (s *PhotonPayUIService) ListWebhookEvents(context.Context, *struct{}) (*[]photon.WebhookEvent, error) {
	events := photon.WebhookEvents()
	return &events, nil
}
func (s *PhotonPayUIService) UpdateWebhook(ctx context.Context, req *UIUpdateWebhookRequest) (*UIWebhookData, error) {
	id, err := photonPayID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateWebhook(ctx, &biz.UIUpdateWebhookRequest{
		ID:        id,
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return photonPayUIWebhookData(item), nil
}
func (s *PhotonPayUIService) DeleteWebhook(ctx context.Context, req *struct {
	ID string `uri:"id" binding:"required"`
}) (*struct{}, error) {
	id, err := photonPayID(req.ID)
	if err != nil {
		return nil, err
	}
	if err := s.usecase.DeleteWebhook(ctx, id); err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

type UICardHolderRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Mobile    string `json:"phone_number" binding:"required"`
}

type UICardHolderData struct {
	ID        string                  `json:"id"`
	FirstName string                  `json:"first_name"`
	LastName  string                  `json:"last_name"`
	Email     string                  `json:"email"`
	Mobile    string                  `json:"phone_number"`
	Status    photon.CardHolderStatus `json:"status"`
	CreatedAt time.Time               `json:"created_at"`
}

type PhotonPayUIService struct {
	usecase *biz.PhotonPayUIUsecase
}

func (s *PhotonPayUIService) CreateAccount(ctx context.Context, req *UICreateAccountRequest) (*UIAccountData, error) {
	item, err := s.usecase.CreateAccount(ctx, &biz.UICreateAccountRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return photonPayUIAccountData(item), nil
}

func (s *PhotonPayUIService) ListAccounts(ctx context.Context, req *UIListRequest) (*UIListResponse[*UIAccountData], error) {
	page, size := types.NormalizePagination(req.PageNumber, req.PageSize)
	items, total, err := s.usecase.ListAccounts(ctx, &biz.ListRequest{Offset: (page - 1) * size, Limit: size})
	if err != nil {
		return nil, err
	}
	return &UIListResponse[*UIAccountData]{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, photonPayUIAccountData),
	}, nil
}

type UIUpdateAccountRequest struct {
	ID   string `uri:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (s *PhotonPayUIService) UpdateAccount(ctx context.Context, req *UIUpdateAccountRequest) (*UIAccountData, error) {
	id, err := photonPayAccountID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateAccount(ctx, &biz.UIUpdateAccountRequest{ID: id, Name: req.Name})
	if err != nil {
		return nil, err
	}
	return photonPayUIAccountData(item), nil
}

func (s *PhotonPayUIService) ListWebhookRecords(
	ctx context.Context,
	req *UIListRequest,
) (*UIListResponse[*UIWebhookRecordData], error) {
	page, size := types.NormalizePagination(req.PageNumber, req.PageSize)
	items, total, err := s.usecase.ListWebhookRecords(ctx, &biz.ListRequest{
		Offset: (page - 1) * size,
		Limit:  size,
	})
	if err != nil {
		return nil, err
	}

	return &UIListResponse[*UIWebhookRecordData]{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, photonPayUIWebhookRecordData),
	}, nil
}

func (s *PhotonPayUIService) ReplayWebhookRecord(
	ctx context.Context,
	req *struct {
		ID string `uri:"id" binding:"required"`
	},
) (*UIWebhookRecordData, error) {
	id, err := photonPayID(req.ID)
	if err != nil {
		return nil, err
	}

	item, err := s.usecase.ReplayWebhookRecord(ctx, id)
	if err != nil {
		return nil, err
	}

	return photonPayUIWebhookRecordData(item), nil
}

func NewPhotonPayUIService(injector *do.Injector) (*PhotonPayUIService, error) {
	return &PhotonPayUIService{
		usecase: do.MustInvoke[*biz.PhotonPayUIUsecase](injector),
	}, nil
}

func (s *PhotonPayUIService) CreateCardHolder(ctx context.Context, req *UICardHolderRequest) (*UICardHolderData, error) {
	accountID, err := photonPayAccountID(req.AccountID)
	if err != nil {
		return nil, err
	}
	holder, err := s.usecase.CreateCardHolder(ctx, &biz.UICreateCardHolderRequest{
		AccountID: accountID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Mobile:    req.Mobile,
	})
	if err != nil {
		return nil, err
	}

	return photonPayUICardHolderData(holder), nil
}

func (s *PhotonPayUIService) ListCardHolders(ctx context.Context, req *UIListRequest) (*UIListResponse[*UICardHolderData], error) {
	holders, err := s.usecase.ListCardHolders(ctx, photonPayUIListRequest(req))
	if err != nil {
		return nil, err
	}

	return &UIListResponse[*UICardHolderData]{
		TotalItems: len(holders),
		Data:       types.BulkConvertSlice(holders, photonPayUICardHolderData),
	}, nil
}

type UICreateCardRequest struct {
	AccountID    string `json:"account_id" binding:"required"`
	CardHolderID string `json:"cardholder_id" binding:"required"`
	CardCurrency string `json:"card_currency" binding:"required"`
	RequestID    string `json:"request_id" binding:"required"`
}

type UICardData struct {
	ID            string            `json:"id"`
	CardHolderID  string            `json:"cardholder_id"`
	CardNumber    string            `json:"card_number"`
	CardBin       string            `json:"card_bin"`
	CardCurrency  string            `json:"card_currency"`
	CardStatus    photon.CardStatus `json:"card_status"`
	Cvv           string            `json:"cvv"`
	ExpiresAt     time.Time         `json:"expires_at"`
	CreatedAt     time.Time         `json:"created_at"`
	FundingSource string            `json:"funding_source"`
	Balance       string            `json:"balance"`
}

func (s *PhotonPayUIService) CreateCard(ctx context.Context, req *UICreateCardRequest) (*UICardData, error) {
	accountID, err := photonPayAccountID(req.AccountID)
	if err != nil {
		return nil, err
	}
	cardHolderID, err := photonPayID(req.CardHolderID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.OpenCard(ctx, &biz.UIOpenCardRequest{
		AccountID:    accountID,
		CardHolderID: cardHolderID,
		Currency:     common.Currency(req.CardCurrency),
		RequestID:    req.RequestID,
	})
	if err != nil {
		return nil, err
	}

	return photonPayUICardData(card), nil
}

func (s *PhotonPayUIService) ListCards(ctx context.Context, req *UIListRequest) (*UIListResponse[*UICardData], error) {
	cards, err := s.usecase.ListCards(ctx, photonPayUIListRequest(req))
	if err != nil {
		return nil, err
	}

	return &UIListResponse[*UICardData]{
		TotalItems: len(cards),
		Data:       types.BulkConvertSlice(cards, photonPayUICardData),
	}, nil
}

func (s *PhotonPayUIService) ListAuthorizations(ctx context.Context, req *UIListRequest) (*UIListResponse[*UIAuthorizationData], error) {
	items, err := s.usecase.ListAuthorizations(ctx, photonPayUIListRequest(req))
	if err != nil {
		return nil, err
	}
	return &UIListResponse[*UIAuthorizationData]{
		TotalItems: len(items),
		Data:       types.BulkConvertSlice(items, photonPayUIAuthorizationData),
	}, nil
}

type UIUpdateCardStatusRequest struct {
	ID         string            `uri:"id" binding:"required"`
	CardStatus photon.CardStatus `json:"card_status" binding:"required"`
}

type UIFundCardRequest struct {
	ID     string  `uri:"id" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func (s *PhotonPayUIService) FundCard(ctx context.Context, req *UIFundCardRequest) (*UICardData, error) {
	id, err := photonPayID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.FundCard(ctx, &biz.UIFundCardRequest{
		CardID: id,
		Amount: decimal.NewFromFloat(req.Amount),
	})
	if err != nil {
		return nil, err
	}

	return photonPayUICardData(item), nil
}

func (s *PhotonPayUIService) UpdateCardStatus(ctx context.Context, req *UIUpdateCardStatusRequest) (*UICardData, error) {
	cardID, err := photonPayID(req.ID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.ChangeCardStatus(ctx, &biz.UIChangeCardStatusRequest{
		CardID: cardID,
		Status: photon.CardStatusToGeneric(req.CardStatus),
	})
	if err != nil {
		return nil, err
	}

	return photonPayUICardData(card), nil
}

type UITransactionData struct {
	ID                   string                   `json:"id"`
	CardID               string                   `json:"card_id"`
	AuthorizationID      string                   `json:"authorization_id"`
	TransactionType      photon.TransactionType   `json:"transaction_type"`
	Status               photon.TransactionStatus `json:"status"`
	Amount               string                   `json:"amount"`
	Currency             string                   `json:"currency"`
	MerchantName         string                   `json:"merchant_name"`
	MerchantCategoryCode string                   `json:"merchant_category_code"`
	TransactedAt         time.Time                `json:"transacted_at"`
}

type UIAuthorizationData struct {
	ID                   string                     `json:"id"`
	CardID               string                     `json:"card_id"`
	Status               photon.AuthorizationStatus `json:"status"`
	AuthorizedAmount     string                     `json:"authorized_amount"`
	Currency             string                     `json:"currency"`
	MerchantName         string                     `json:"merchant_name"`
	MerchantCategoryCode string                     `json:"merchant_category_code"`
	AuthorizationCode    string                     `json:"authorization_code"`
	AuthorizedAt         time.Time                  `json:"authorized_at"`
}

type UISimulateAuthorizationRequest struct {
	CardID               string  `json:"card_id" binding:"required"`
	TransactionAmount    float64 `json:"transaction_amount" binding:"required"`
	TransactionCurrency  string  `json:"transaction_currency" binding:"required"`
	MerchantName         string  `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string  `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string  `json:"merchant_country"`
	MerchantCity         string  `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

type UISimulateAuthorizationData struct {
	Approved      bool                     `json:"approved"`
	Status        photon.TransactionStatus `json:"status"`
	Authorization *UIAuthorizationData     `json:"authorization"`
	Transaction   *UITransactionData       `json:"transaction"`
}

type UISimulateRefundRequest struct {
	CardID               string  `json:"card_id" binding:"required"`
	Amount               float64 `json:"amount" binding:"required,gt=0"`
	Currency             string  `json:"currency" binding:"required"`
	MerchantName         string  `json:"merchant_name" binding:"required"`
	MerchantCategoryCode string  `json:"merchant_category_code" binding:"required"`
	MerchantCountry      string  `json:"merchant_country" binding:"required"`
	MerchantCity         string  `json:"merchant_city"` // Invalid: generic model has no merchant city field.
}

func (s *PhotonPayUIService) SimulateRefund(ctx context.Context, req *UISimulateRefundRequest) (*UITransactionData, error) {
	cardID, err := photonPayID(req.CardID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.SimulateRefund(ctx, &biz.UISimulateRefundRequest{
		CardID:          cardID,
		Amount:          decimal.NewFromFloat(req.Amount),
		Currency:        common.Currency(req.Currency),
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}
	return photonPayUITransactionData(item), nil
}

func (s *PhotonPayUIService) SimulateAuthorization(ctx context.Context, req *UISimulateAuthorizationRequest) (*UISimulateAuthorizationData, error) {
	cardID, err := photonPayID(req.CardID)
	if err != nil {
		return nil, err
	}
	result, err := s.usecase.SimulateAuthorization(ctx, &biz.UISimulateAuthorizationRequest{
		CardID:          cardID,
		Amount:          decimal.NewFromFloat(req.TransactionAmount),
		Currency:        common.Currency(req.TransactionCurrency),
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantCategoryCode,
	})
	if err != nil {
		return nil, err
	}

	return &UISimulateAuthorizationData{
		Approved:      true,
		Status:        photon.TransactionStatusFromGeneric(result.Authorization.Status),
		Authorization: photonPayUIAuthorizationData(result.Authorization),
		Transaction:   photonPayUITransactionData(result.CardTransaction),
	}, nil
}

type UIApplyTransactionStepRequest struct {
	ID     string  `uri:"id" binding:"required"`
	Amount float64 `json:"amount"`
}

func (s *PhotonPayUIService) ClearTransaction(ctx context.Context, req *UIApplyTransactionStepRequest) (*UITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_CLEAR)
}

func (s *PhotonPayUIService) ReverseTransaction(ctx context.Context, req *UIApplyTransactionStepRequest) (*UITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_VOID)
}

func (s *PhotonPayUIService) RefundTransaction(ctx context.Context, req *UIApplyTransactionStepRequest) (*UITransactionData, error) {
	return s.applyTransactionStep(ctx, req, common.CardTransactionType_REFUND)
}

func (s *PhotonPayUIService) applyTransactionStep(ctx context.Context, req *UIApplyTransactionStepRequest, transactionType common.CardTransactionType) (*UITransactionData, error) {
	transactionID, err := photonPayID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.ApplyTransactionStep(ctx, &biz.UIApplyTransactionStepRequest{
		CardTransactionID: transactionID,
		Type:              transactionType,
		Amount:            decimal.NewFromFloat(req.Amount),
	})
	if err != nil {
		return nil, err
	}

	return photonPayUITransactionData(item), nil
}

func (s *PhotonPayUIService) ListTransactions(ctx context.Context, req *UIListRequest) (*UIListResponse[*UITransactionData], error) {
	transactions, err := s.usecase.ListTransactions(ctx, photonPayUIListRequest(req))
	if err != nil {
		return nil, err
	}

	return &UIListResponse[*UITransactionData]{
		TotalItems: len(transactions),
		Data:       types.BulkConvertSlice(transactions, photonPayUITransactionData),
	}, nil
}

func photonPayUIListRequest(req *UIListRequest) *biz.ListRequest {
	page, size := types.NormalizePagination(req.PageNumber, req.PageSize)

	return &biz.ListRequest{
		Offset: (page - 1) * size,
		Limit:  size,
	}
}

func photonPayUICardHolderData(item *model.CardHolder) *UICardHolderData {
	return &UICardHolderData{
		ID:        photonPayIDString(item.ID),
		FirstName: item.FirstName,
		LastName:  item.LastName,
		Email:     item.Email,
		Mobile:    item.Mobile,
		Status:    photon.CardHolderStatusFromGeneric(item.Status),
		CreatedAt: item.CreatedAt,
	}
}

func photonPayUICardData(item *model.Card) *UICardData {
	balance := decimal.Zero
	fundingSource := "卡资金"
	if item.VirtualAccount != nil && item.VirtualAccount.Wallet != nil {
		balance = item.VirtualAccount.Wallet.Amount
		fundingSource = "虚拟账户共享资金"
	} else if item.Wallet != nil {
		balance = item.Wallet.Amount
	}

	return &UICardData{
		ID:            photonPayIDString(item.ID),
		CardHolderID:  photonPayIDString(item.CardHolderID),
		CardNumber:    item.CardNumber,
		CardBin:       item.CardBin,
		CardCurrency:  string(item.CardCurrency),
		CardStatus:    photon.CardStatusFromGeneric(item.Status),
		Cvv:           item.Cvv,
		ExpiresAt:     item.ExpireAt,
		CreatedAt:     item.CreatedAt,
		FundingSource: fundingSource,
		Balance:       balance.String(),
	}
}

func photonPayUIAuthorizationData(item *model.Authorization) *UIAuthorizationData {
	return &UIAuthorizationData{
		ID:                   photonPayIDString(item.ID),
		CardID:               photonPayIDString(item.CardID),
		Status:               photon.AuthorizationStatusFromGeneric(item.Status),
		AuthorizedAmount:     item.Amount.String(),
		Currency:             string(item.Currency),
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		AuthorizationCode:    item.AuthorizationCode,
		AuthorizedAt:         item.CreatedAt,
	}
}

func photonPayUIWebhookData(item *model.WebhookConfig) *UIWebhookData {
	return &UIWebhookData{
		ID:        photonPayIDString(item.ID),
		AccountID: photonPayIDString(item.AccountID),
		Event:     photon.WebhookEvent(item.Event),
		TargetURL: item.TargetURL,
		Enabled:   item.Enabled,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func photonPayUIVirtualAccountData(item *model.VirtualAccount) *UIVirtualAccountData {
	return &UIVirtualAccountData{
		ID:        photonPayIDString(item.ID),
		AccountID: photonPayIDString(item.AccountID),
		Name:      item.Name,
		Currency:  string(item.Wallet.Currency),
		Balance:   item.Wallet.Amount.String(),
		CreatedAt: item.CreatedAt,
	}
}

func photonPayUIAuthorizationConfigData(item *model.AuthorizationConfig) *UIAuthorizationConfigData {
	return &UIAuthorizationConfigData{
		AccountID:     photonPayIDString(item.AccountID),
		TargetURL:     item.TargetURL,
		Enabled:       item.Enabled,
		TimeoutMillis: item.TimeoutMillis,
		UpdatedAt:     item.UpdatedAt,
	}
}

func photonPayUIWebhookRecordData(item *model.WebhookRecord) *UIWebhookRecordData {
	return &UIWebhookRecordData{
		ID:              photonPayIDString(item.ID),
		AccountID:       photonPayIDString(item.AccountID),
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

func photonPayUIAccountData(item *model.Account) *UIAccountData {
	return &UIAccountData{
		ID:        photonPayIDString(item.ID),
		Name:      item.Name,
		CreatedAt: item.CreatedAt,
	}
}

func photonPayUITransactionData(item *model.CardTransaction) *UITransactionData {
	return &UITransactionData{
		ID:                   photonPayIDString(item.ID),
		CardID:               photonPayIDString(item.CardID),
		AuthorizationID:      photonPayIDString(item.AuthorizationID),
		TransactionType:      photon.TransactionTypeFromGeneric(item.Type),
		Status:               photon.TransactionStatusFromGeneric(item.Status),
		Amount:               item.TxAmount.String(),
		Currency:             string(item.TxCurrency),
		MerchantName:         item.MerchantName,
		MerchantCategoryCode: item.MerchantMCC,
		TransactedAt:         item.CreatedAt,
	}
}
