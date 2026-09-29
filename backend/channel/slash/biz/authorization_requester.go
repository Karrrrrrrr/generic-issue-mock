package biz

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"generic-mock/channel/slash/pkg/idconv"
	"generic-mock/enums"
	"generic-mock/model"
	sharedbiz "generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/samber/do/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const slashAuthorizationRequestEventType = "authorization.request"

type slashAuthorizationPayload struct {
	Event slashAuthorizationEvent `json:"event"`
	Data  slashAuthorizationData  `json:"data"`
}

type slashAuthorizationEvent struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type slashAuthorizationData struct {
	Transaction        slashAuthorizationTransaction        `json:"transaction"`
	Card               slashAuthorizationCard               `json:"card"`
	Account            slashAuthorizationAccount            `json:"account"`
	VirtualAccount     slashAuthorizationVirtualAccount     `json:"virtualAccount"`
	Merchant           slashAuthorizationMerchant           `json:"merchant"`
	Amount             slashAuthorizationAmount             `json:"amount"`
	CurrencyConversion slashAuthorizationCurrencyConversion `json:"currencyConversion"`
}

type slashAuthorizationTransaction struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
}

type slashAuthorizationCard struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type slashAuthorizationAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type slashAuthorizationVirtualAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type slashAuthorizationMerchant struct {
	Name         string                     `json:"name"`
	CategoryCode string                     `json:"categoryCode"`
	Location     slashAuthorizationLocation `json:"location"`
}

type slashAuthorizationLocation struct {
	Country string `json:"country"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type slashAuthorizationAmount struct {
	AmountCents int64 `json:"amountCents"`
}

type slashAuthorizationCurrencyConversion struct {
	ConversionRate       float64 `json:"conversionRate"`
	OriginalAmountCents  int64   `json:"originalAmountCents"`
	OriginalCurrencyCode string  `json:"originalCurrencyCode"`
}

type slashVirtualAccountRequest struct {
	VirtualAccountID *model.ID
	VirtualAccount   *model.VirtualAccount
}

type SlashAuthorizationRequester struct {
	authorizationConfigRepo SlashAuthorizationConfigRepository
	authorizationClient     SlashAuthorizationClient
}

var _ sharedbiz.AuthorizationRequester = (*SlashAuthorizationRequester)(nil)

func NewSlashAuthorizationRequester(injector do.Injector) (*SlashAuthorizationRequester, error) {
	return &SlashAuthorizationRequester{
		authorizationConfigRepo: do.MustInvoke[SlashAuthorizationConfigRepository](injector),
		authorizationClient:     do.MustInvoke[SlashAuthorizationClient](injector),
	}, nil
}

func (r *SlashAuthorizationRequester) RequestAuthorization(
	ctx context.Context,
	req *sharedbiz.AuthorizationRequest,
) (*sharedbiz.AuthorizationDecision, error) {
	if req == nil || req.Channel != enums.Channel_Slash ||
		req.AccountID <= 0 || req.Card == nil ||
		req.Authorization == nil || req.CardTransaction == nil {
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}

	exists, err := r.authorizationConfigRepo.ExistByAccountID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw(
			"check slash authorization config",
			"account_id",
			req.AccountID,
			"error",
			err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		zap.S().Warnw(
			"slash authorization config missing",
			"account_id",
			req.AccountID,
			"transaction_id",
			req.CardTransaction.ID,
		)
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}

	config, err := r.authorizationConfigRepo.FindByAccountID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw(
			"find slash authorization config",
			"account_id",
			req.AccountID,
			"error",
			err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !config.Enabled || config.TargetURL == "" {
		zap.S().Warnw(
			"slash authorization config disabled",
			"account_id",
			req.AccountID,
			"authorization_config_id",
			config.ID,
			"transaction_id",
			req.CardTransaction.ID,
		)
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}

	payload, err := json.Marshal(r.buildPayload(req))
	if err != nil {
		zap.S().Errorw("marshal slash authorization request", "error", err)
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}

	result, err := r.authorizationClient.RequestAuthorization(ctx, &SlashAuthorizationRequestDeliveryRequest{
		TargetURL:     config.TargetURL,
		Payload:       payload,
		TimeoutMillis: config.TimeoutMillis,
	})
	if err != nil {
		zap.S().Errorw(
			"send slash authorization request",
			"account_id",
			req.AccountID,
			"authorization_config_id",
			config.ID,
			"target_url",
			config.TargetURL,
			"transaction_id",
			req.CardTransaction.ID,
			"error",
			err,
		)
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}
	if result.StatusCode < http.StatusOK || result.StatusCode >= http.StatusMultipleChoices {
		zap.S().Errorw(
			"slash authorization request rejected by http status",
			"account_id",
			req.AccountID,
			"authorization_config_id",
			config.ID,
			"target_url",
			config.TargetURL,
			"transaction_id",
			req.CardTransaction.ID,
			"status_code",
			result.StatusCode,
			"response_body",
			result.ResponseBody,
		)
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}

	var response struct {
		Approved *bool  `json:"approved"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(result.ResponseBody), &response); err != nil {
		zap.S().Errorw(
			"decode slash authorization response",
			"account_id",
			req.AccountID,
			"authorization_config_id",
			config.ID,
			"transaction_id",
			req.CardTransaction.ID,
			"response_body",
			result.ResponseBody,
			"error",
			err,
		)
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}
	if response.Approved == nil {
		zap.S().Errorw(
			"slash authorization response missing approved",
			"account_id",
			req.AccountID,
			"authorization_config_id",
			config.ID,
			"transaction_id",
			req.CardTransaction.ID,
			"response_body",
			result.ResponseBody,
		)
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}

	return &sharedbiz.AuthorizationDecision{
		Approved: *response.Approved,
		Reason:   response.Reason,
	}, nil
}

func (r *SlashAuthorizationRequester) buildPayload(
	req *sharedbiz.AuthorizationRequest,
) *slashAuthorizationPayload {
	transactionID := idconv.ToUUID(req.CardTransaction.ID)
	amountCents := decimalAmountCents(req.CardTransaction.TxAmount)
	virtualAccount := slashVirtualAccount(&slashVirtualAccountRequest{
		VirtualAccountID: req.Card.VirtualAccountID,
		VirtualAccount:   req.Card.VirtualAccount,
	})

	return &slashAuthorizationPayload{
		Event: slashAuthorizationEvent{
			Type: slashAuthorizationRequestEventType,
			ID:   transactionID,
		},
		Data: slashAuthorizationData{
			Transaction: slashAuthorizationTransaction{
				ID:        transactionID,
				Timestamp: req.CardTransaction.CreatedAt.UTC().Format(time.RFC3339),
			},
			Card: slashAuthorizationCard{
				ID:   idconv.ToUUID(req.Card.ID),
				Name: req.Card.MaskedNumber(),
			},
			Account: slashAuthorizationAccount{
				ID:   idconv.ToUUID(req.AccountID),
				Name: req.Card.Account.GetName(),
			},
			VirtualAccount: virtualAccount,
			Merchant: slashAuthorizationMerchant{
				Name:         req.CardTransaction.MerchantName,
				CategoryCode: req.CardTransaction.MerchantMCC,
				Location: slashAuthorizationLocation{
					Country: req.CardTransaction.MerchantCountry,
					City:    "",
					State:   "",
					Zip:     "",
				},
			},
			Amount: slashAuthorizationAmount{
				AmountCents: amountCents,
			},
			CurrencyConversion: slashAuthorizationCurrencyConversion{
				ConversionRate:       1,
				OriginalAmountCents:  amountCents,
				OriginalCurrencyCode: string(req.CardTransaction.TxCurrency),
			},
		},
	}
}

func slashVirtualAccount(req *slashVirtualAccountRequest) slashAuthorizationVirtualAccount {
	if req.VirtualAccountID == nil {
		return slashAuthorizationVirtualAccount{}
	}
	name := ""
	if req.VirtualAccount != nil {
		name = req.VirtualAccount.Name
	}
	return slashAuthorizationVirtualAccount{
		ID:   idconv.ToUUID(*req.VirtualAccountID),
		Name: name,
	}
}

func decimalAmountCents(amount decimal.Decimal) int64 {
	return amount.Mul(decimal.NewFromInt(100)).Round(0).IntPart()
}
