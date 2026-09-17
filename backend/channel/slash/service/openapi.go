package service

import (
	"context"
	"encoding/json"
	"time"

	"generic-mock/channel/slash/biz"
	slash "generic-mock/channel/slash/enums"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
)

type SlashOpenAPIService struct {
	usecase        *biz.SlashOpenAPIUsecase
	webhookUsecase *biz.SlashWebhookUsecase
}

func NewSlashOpenAPIService(injector *do.Injector) (*SlashOpenAPIService, error) {
	return &SlashOpenAPIService{
		usecase:        do.MustInvoke[*biz.SlashOpenAPIUsecase](injector),
		webhookUsecase: do.MustInvoke[*biz.SlashWebhookUsecase](injector),
	}, nil
}

type OpenAPIAccountRequest struct {
	APIKey string `header:"X-API-Key" binding:"required"`
}

func (s *SlashOpenAPIService) accountID(req *OpenAPIAccountRequest) (model.ID, error) {
	return slashAccountID(req.APIKey)
}

type OpenAPIListRequest struct {
	OpenAPIAccountRequest
	PageNumber int `form:"page_number"`
	PageSize   int `form:"page_size"`
}

type OpenAPIMetadata struct {
	Count int `json:"count"`
}

type OpenAPIVirtualAccountData struct {
	VirtualAccount OpenAPIVirtualAccountDetails `json:"virtualAccount"`
	Balance        OpenAPIAmount                `json:"balance"`
	Spend          OpenAPIAmount                `json:"spend"`
}

type OpenAPIVirtualAccountDetails struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AccountType string `json:"accountType"`
}

type OpenAPIAmount struct {
	AmountCents int64 `json:"amountCents"`
}

type OpenAPIListVirtualAccountsData struct {
	Items    []*OpenAPIVirtualAccountData `json:"items"`
	Metadata OpenAPIMetadata              `json:"metadata"`
}

func (s *SlashOpenAPIService) ListVirtualAccounts(ctx context.Context, req *OpenAPIAccountRequest) (*OpenAPIListVirtualAccountsData, error) {
	accountID, err := s.accountID(req)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListVirtualAccounts(ctx, accountID)
	if err != nil {
		return nil, err
	}
	data := make([]*OpenAPIVirtualAccountData, 0, len(items))
	for _, item := range items {
		data = append(data, &OpenAPIVirtualAccountData{
			VirtualAccount: OpenAPIVirtualAccountDetails{
				ID:          slashIDString(item.ID),
				Name:        item.Name,
				AccountType: "primary",
			},
			Balance: OpenAPIAmount{AmountCents: item.Wallet.Amount.Mul(decimal.NewFromInt(100)).IntPart()},
			Spend:   OpenAPIAmount{AmountCents: item.Wallet.Out.Mul(decimal.NewFromInt(100)).IntPart()},
		})
	}
	return &OpenAPIListVirtualAccountsData{
		Items: data,
		Metadata: OpenAPIMetadata{
			Count: len(data),
		},
	}, nil
}

type OpenAPIVirtualAccountTransferRequest struct {
	OpenAPIAccountRequest
	Source      string `json:"source" binding:"required"`
	Destination string `json:"destination" binding:"required"`
	AmountCents int64  `json:"amountCents" binding:"required"`
}

type OpenAPIVirtualAccountTransferData struct {
	ID string `json:"id"`
}

func (s *SlashOpenAPIService) TransferVirtualAccount(ctx context.Context, req *OpenAPIVirtualAccountTransferRequest) (*OpenAPIVirtualAccountTransferData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	source, err := slashID(req.Source)
	if err != nil {
		return nil, err
	}
	destination, err := slashID(req.Destination)
	if err != nil {
		return nil, err
	}
	err = s.usecase.TransferVirtualAccount(ctx, &biz.OpenAPIVirtualAccountTransferRequest{
		AccountID:   accountID,
		Source:      source,
		Destination: destination,
		AmountCents: req.AmountCents,
	})
	if err != nil {
		return nil, err
	}
	return &OpenAPIVirtualAccountTransferData{ID: req.Source}, nil
}

type OpenAPICard struct {
	ID               string           `json:"id"`
	AccountID        string           `json:"accountId"`
	VirtualAccountID string           `json:"virtualAccountId"`
	Last4            string           `json:"last4"`
	Name             string           `json:"name"`
	ExpiryMonth      string           `json:"expiryMonth"`
	ExpiryYear       string           `json:"expiryYear"`
	Status           slash.CardStatus `json:"status"`
	IsPhysical       bool             `json:"isPhysical"`
	IsSingleUse      bool             `json:"isSingleUse"`
	Pan              string           `json:"pan"`
	Cvv              string           `json:"cvv"`
	CardGroupID      string           `json:"cardGroupId"`
	CreatedAt        time.Time        `json:"createdAt"`
	CardProductID    string           `json:"cardProductId"`
}

type OpenAPIListCardsRequest struct {
	OpenAPIListRequest
	FilterStatus slash.CardStatus `form:"filter:status"`
	Cursor       string           `form:"cursor"` // Invalid: cursor pagination is unsupported.
}

type OpenAPIListCardsData struct {
	Items    []*OpenAPICard  `json:"items"`
	Metadata OpenAPIMetadata `json:"metadata"`
}

func (s *SlashOpenAPIService) ListCards(ctx context.Context, req *OpenAPIListCardsRequest) (*OpenAPIListCardsData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	offset, limit := openAPIPagination(req.PageNumber, req.PageSize)
	items, err := s.usecase.ListCards(ctx, &biz.OpenAPIListCardsRequest{
		AccountID: accountID,
		Offset:    offset,
		Limit:     limit,
		Status:    slash.CardStatusToGeneric(req.FilterStatus),
	})
	if err != nil {
		return nil, err
	}

	return &OpenAPIListCardsData{
		Items: types.BulkConvertSlice(items, openAPICard),
		Metadata: OpenAPIMetadata{
			Count: len(items),
		},
	}, nil
}

type OpenAPICreateCardRequest struct {
	OpenAPIAccountRequest
	AccountID          string          `json:"accountId"`        // Invalid: mock has one generic account.
	VirtualAccountID   string          `json:"virtualAccountId"` // Invalid: virtual-account assignment is unsupported.
	Type               slash.CardType  `json:"type" binding:"required"`
	Name               string          `json:"name" binding:"required"` // Invalid: card names are not persisted.
	IsSingleUse        bool            `json:"isSingleUse"`             // Invalid: single-use cards are unsupported.
	SpendingConstraint json.RawMessage `json:"spendingConstraint"`      // Invalid: spending constraints are unsupported.
	UserData           OpenAPIUserData `json:"userData" binding:"required"`
	CardGroupID        string          `json:"cardGroupId"` // Invalid: card groups are unsupported.
	CardProductID      string          `json:"cardProductId" binding:"required"`
}

type OpenAPIUserData struct {
	RequestID string `json:"requestId" binding:"required"`
	CardID    string `json:"cardId" binding:"required"`
}

func (s *SlashOpenAPIService) CreateCard(ctx context.Context, req *OpenAPICreateCardRequest) (*OpenAPICard, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	cardProductID, err := slashID(req.CardProductID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.CreateCard(ctx, &biz.OpenAPICreateCardRequest{
		AccountID:     accountID,
		CardProductID: cardProductID,
		Currency:      enums.Currency_USD,
		RequestID:     req.UserData.RequestID,
	})
	if err != nil {
		return nil, err
	}
	s.webhookUsecase.Dispatch(ctx, slashWebhookDispatchRequest(
		item.AccountID,
		slash.WebhookEventCardCreate,
		item.ID,
	))

	return openAPICard(item), nil
}

type OpenAPIIDRequest struct {
	OpenAPIAccountRequest
	ID string `uri:"id" binding:"required"`
}

func (s *SlashOpenAPIService) GetCard(ctx context.Context, req *OpenAPIIDRequest) (*OpenAPICard, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetCard(ctx, &biz.ResourceRequest{AccountID: &accountID, ID: id})
	if err != nil {
		return nil, err
	}

	return openAPICard(item), nil
}

type OpenAPIUpdateCardRequest struct {
	OpenAPIIDRequest
	Name               *string          `json:"name"` // Invalid: card names are not persisted.
	Status             slash.CardStatus `json:"status" binding:"required"`
	CardGroupID        *string          `json:"cardGroupId"`        // Invalid: card groups are unsupported.
	SpendingConstraint json.RawMessage  `json:"spendingConstraint"` // Invalid: spending constraints are unsupported.
	UserData           *OpenAPIUserData `json:"userData"`           // Invalid: only creation request IDs are persisted.
}

func (s *SlashOpenAPIService) UpdateCard(ctx context.Context, req *OpenAPIUpdateCardRequest) (*OpenAPICard, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.UpdateCard(ctx, &biz.OpenAPIUpdateCardRequest{
		AccountID: accountID,
		ID:        id,
		Status:    slash.CardStatusToGeneric(req.Status),
	})
	if err != nil {
		return nil, err
	}
	event := slash.WebhookEventCardUpdate
	if req.Status == slash.CardStatus_Closed {
		event = slash.WebhookEventCardDelete
	}
	s.webhookUsecase.Dispatch(ctx, slashWebhookDispatchRequest(
		item.AccountID,
		event,
		item.ID,
	))

	return openAPICard(item), nil
}

type OpenAPICardProduct struct {
	ID     string                  `json:"id"`
	Prefix string                  `json:"prefix"`
	Status slash.CardProductStatus `json:"status"`
}

type OpenAPIListCardProductsRequest struct {
	OpenAPIAccountRequest
}

type OpenAPIListCardProductsData struct {
	Items []*OpenAPICardProduct `json:"items"`
}

func (s *SlashOpenAPIService) ListCardProducts(ctx context.Context, req *OpenAPIListCardProductsRequest) (*OpenAPIListCardProductsData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListCardProducts(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return &OpenAPIListCardProductsData{
		Items: types.BulkConvertSlice(items, func(item *model.CardProduct) *OpenAPICardProduct {
			return &OpenAPICardProduct{
				ID:     slashIDString(item.ID),
				Prefix: item.Prefix,
				Status: slash.CardProductStatus_Active,
			}
		}),
	}, nil
}

type OpenAPITransaction struct {
	ID                      string                  `json:"id"`
	Date                    time.Time               `json:"date"`
	Description             string                  `json:"description"`
	MerchantDescription     string                  `json:"merchantDescription"`
	AmountCents             int                     `json:"amountCents"`
	Status                  slash.TransactionStatus `json:"status"`
	DetailedStatus          slash.TransactionStatus `json:"detailedStatus"`
	AccountID               string                  `json:"accountId"`
	VirtualAccountID        string                  `json:"virtualAccountId"`
	CardID                  string                  `json:"cardId"`
	AuthorizedAt            time.Time               `json:"authorizedAt"`
	ProviderAuthorizationID string                  `json:"providerAuthorizationId"`
}

type OpenAPIListTransactionsRequest struct {
	OpenAPIListRequest
	FilterCardID    string `form:"filter:cardId"`
	Cursor          string `form:"cursor"` // Invalid: cursor pagination is unsupported.
	AuthorizationID string `form:"filter:providerAuthorizationId"`
}

type OpenAPIListTransactionsData struct {
	Items    []*OpenAPITransaction `json:"items"`
	Metadata OpenAPIMetadata       `json:"metadata"`
}

func (s *SlashOpenAPIService) ListTransactions(ctx context.Context, req *OpenAPIListTransactionsRequest) (*OpenAPIListTransactionsData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	offset, limit := openAPIPagination(req.PageNumber, req.PageSize)
	cardID, err := slashOptionalID(req.FilterCardID)
	if err != nil {
		return nil, err
	}
	authorizationID, err := slashOptionalID(req.AuthorizationID)
	if err != nil {
		return nil, err
	}
	items, err := s.usecase.ListTransactions(ctx, &biz.OpenAPIListTransactionsRequest{
		AccountID:       accountID,
		Offset:          offset,
		Limit:           limit,
		CardID:          cardID,
		AuthorizationID: authorizationID,
	})
	if err != nil {
		return nil, err
	}

	return &OpenAPIListTransactionsData{
		Items: types.BulkConvertSlice(items, openAPITransaction),
		Metadata: OpenAPIMetadata{
			Count: len(items),
		},
	}, nil
}

func (s *SlashOpenAPIService) GetTransaction(ctx context.Context, req *OpenAPIIDRequest) (*OpenAPITransaction, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	id, err := slashID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetTransaction(ctx, &biz.ResourceRequest{AccountID: &accountID, ID: id})
	if err != nil {
		return nil, err
	}

	return openAPITransaction(item), nil
}

func openAPICard(item *model.Card) *OpenAPICard {
	return &OpenAPICard{
		ID:               slashIDString(item.ID),
		AccountID:        slashIDString(item.AccountID),
		VirtualAccountID: virtualAccountIDString(item.VirtualAccountID),
		Last4:            item.CardNumber[len(item.CardNumber)-4:],
		ExpiryMonth:      item.ExpireAt.Format("01"),
		ExpiryYear:       item.ExpireAt.Format("2006"),
		Status:           slash.CardStatusFromGeneric(item.Status),
		IsPhysical:       item.FormType == enums.CardFormType_Physical,
		Pan:              item.CardNumber,
		Cvv:              item.Cvv,
		CreatedAt:        item.CreatedAt,
		CardProductID:    slashIDString(item.CardProductID),
	}
}

func openAPITransaction(item *model.CardTransaction) *OpenAPITransaction {
	authorizedAt := time.Time{}
	if item.Authorization != nil {
		authorizedAt = item.Authorization.CreatedAt.UTC()
	}

	return &OpenAPITransaction{
		ID:                      slashIDString(item.ID),
		Date:                    item.CreatedAt.UTC(),
		Description:             item.MerchantName,
		MerchantDescription:     item.MerchantName,
		AmountCents:             int(item.TxAmount.Mul(decimal.NewFromInt(100)).IntPart()),
		Status:                  slash.TransactionStatusFromGeneric(item.Status),
		DetailedStatus:          slash.TransactionStatusFromGeneric(item.Status),
		AccountID:               slashIDString(item.AccountID),
		CardID:                  slashIDString(item.CardID),
		AuthorizedAt:            authorizedAt,
		ProviderAuthorizationID: slashIDString(item.AuthorizationID),
	}
}

func virtualAccountIDString(id *model.ID) string {
	if id == nil {
		return ""
	}
	return slashIDString(*id)
}

func openAPIPagination(pageNumber int, pageSize int) (int, int) {
	page, size := types.NormalizePagination(pageNumber, pageSize)

	return (page - 1) * size, size
}
