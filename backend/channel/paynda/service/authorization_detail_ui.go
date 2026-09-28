package service

import (
	"context"
	"encoding/json"
	"time"

	"generic-mock/channel/paynda/biz"
	channelEnums "generic-mock/channel/paynda/enums"
	"generic-mock/channel/paynda/pkg/idconv"
	common "generic-mock/enums"

	"github.com/shopspring/decimal"
)

type GetAuthorizationDetailRequest struct {
	ManagementAccountRequest
	ID string `uri:"id" binding:"required"`
}

type ReverseAuthorizationRequest struct {
	ID     string          `uri:"id" binding:"required"`
	Amount decimal.Decimal `json:"amount"`
}

type RefundAuthorizationRequest struct {
	ID     string          `uri:"id" binding:"required"`
	Amount decimal.Decimal `json:"amount"`
}

type AuthorizationDetailData struct {
	AuthorizationBalanceData
	CardNumber        string                         `json:"card_number"`
	AuthorizationCode string                         `json:"authorization_code"`
	MerchantCountry   string                         `json:"merchant_country"`
	MerchantMCC       string                         `json:"merchant_mcc"`
	RawPayload        json.RawMessage                `json:"raw_payload"`
	Transactions      []AuthorizationTransactionData `json:"transactions"`
}

type AuthorizationTransactionData struct {
	ID              string                         `json:"id"`
	AccountID       string                         `json:"account_id"`
	AccountName     string                         `json:"account_name"`
	TransactionType channelEnums.TransactionType   `json:"transaction_type"`
	Status          channelEnums.TransactionStatus `json:"status"`
	Amount          string                         `json:"amount"`
	Currency        common.Currency                `json:"currency"`
	CreatedAt       time.Time                      `json:"created_at"`
}

func (s *PayndaUIService) GetAuthorizationDetail(ctx context.Context, req *GetAuthorizationDetailRequest) (*AuthorizationDetailData, error) {
	accountID, err := idconv.FromAccountString(req.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.GetAuthorizationDetail(ctx, &biz.GetAuthorizationDetailRequest{
		AccountID: accountID,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}
	auth := item.Balance.Authorization
	result := &AuthorizationDetailData{
		AuthorizationBalanceData: authorizationBalanceData(item.Balance),
		CardNumber:               item.Card.CardNumber,
		AuthorizationCode:        auth.AuthorizationCode,
		MerchantCountry:          auth.MerchantCountry,
		MerchantMCC:              auth.MerchantMCC,
		Transactions:             make([]AuthorizationTransactionData, 0, len(item.Transactions)),
	}
	if json.Valid(auth.RawPayload) {
		result.RawPayload = json.RawMessage(auth.RawPayload)
	}
	for _, transaction := range item.Transactions {
		result.Transactions = append(result.Transactions, AuthorizationTransactionData{
			ID:              idconv.ToString(transaction.ID),
			AccountID:       idconv.ToString(auth.AccountID),
			AccountName:     uiAccountName(auth.Account),
			TransactionType: channelEnums.TransactionTypeFromGeneric(transaction.Type),
			Status:          channelEnums.TransactionStatusFromGeneric(transaction.Status),
			Amount:          transaction.TxAmount.String(),
			Currency:        transaction.TxCurrency,
			CreatedAt:       transaction.CreatedAt,
		})
	}
	return result, nil
}

func (s *PayndaUIService) ReverseAuthorization(ctx context.Context, req *ReverseAuthorizationRequest) (*ClearAuthorizationData, error) {
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.ReverseAuthorization(ctx, &biz.ReverseAuthorizationRequest{
		ID:     id,
		Amount: req.Amount,
	})
	if err != nil {
		return nil, err
	}

	return &ClearAuthorizationData{ID: idconv.ToString(item.ID)}, nil
}

func (s *PayndaUIService) RefundAuthorization(ctx context.Context, req *RefundAuthorizationRequest) (*ClearAuthorizationData, error) {
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.RefundAuthorization(ctx, &biz.RefundAuthorizationRequest{
		ID:     id,
		Amount: req.Amount,
	})
	if err != nil {
		return nil, err
	}

	return &ClearAuthorizationData{ID: idconv.ToString(item.ID)}, nil
}
