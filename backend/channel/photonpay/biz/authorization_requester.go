package biz

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	"generic-mock/channel/photonpay/pkg/idconv"
	"generic-mock/enums"
	sharedbiz "generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

type photonPayAuthorizationPayload struct {
	TransactionID         string `json:"transactionId"`
	OriginTransactionID   string `json:"originTransactionId"`
	CardID                string `json:"cardId"`
	TransactionType       string `json:"transactionType"`
	TransactionHappenedAt string `json:"transactionHappenedAt"`
	TransactionCurrency   string `json:"transactionCurrency"`
	TransactionAmount     string `json:"transactionAmount"`
	AuthCurrency          string `json:"authCurrency"`
	AuthAmount            string `json:"authAmount"`
	FeeCurrency           string `json:"feeCurrency"`
	FeeAmount             string `json:"feeAmount"`
	MCC                   string `json:"mcc"`
	MerchantName          string `json:"merchantName"`
	TransactionCountry    string `json:"transactionCountry"`
	AuthCode              string `json:"authCode"`
}

type PhotonPayAuthorizationRequester struct {
	authorizationConfigRepo AuthorizationConfigRepository
	authorizationClient     AuthorizationClient
}

var _ sharedbiz.AuthorizationRequester = (*PhotonPayAuthorizationRequester)(nil)

func NewPhotonPayAuthorizationRequester(injector do.Injector) (*PhotonPayAuthorizationRequester, error) {
	return &PhotonPayAuthorizationRequester{
		authorizationConfigRepo: do.MustInvoke[AuthorizationConfigRepository](injector),
		authorizationClient:     do.MustInvoke[AuthorizationClient](injector),
	}, nil
}

func (r *PhotonPayAuthorizationRequester) RequestAuthorization(
	ctx context.Context,
	req *sharedbiz.AuthorizationRequest,
) (*sharedbiz.AuthorizationDecision, error) {
	if req == nil || req.Channel != enums.Channel_PhotonPay ||
		req.AccountID <= 0 || req.Card == nil ||
		req.Authorization == nil || req.CardTransaction == nil {
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}

	exists, err := r.authorizationConfigRepo.ExistByAccountID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw(
			"check photonpay authorization config",
			"account_id",
			req.AccountID,
			"error",
			err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		zap.S().Warnw(
			"photonpay authorization config missing",
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
			"find photonpay authorization config",
			"account_id",
			req.AccountID,
			"error",
			err,
		)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !config.Enabled || config.TargetURL == "" {
		zap.S().Warnw(
			"photonpay authorization config disabled",
			"account_id",
			req.AccountID,
			"authorization_config_id",
			config.ID,
			"transaction_id",
			req.CardTransaction.ID,
		)
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}

	payload, err := json.Marshal(photonPayAuthorizationPayload{
		TransactionID:         idconv.ToString(req.CardTransaction.ID),
		OriginTransactionID:   "",
		CardID:                idconv.ToString(req.Card.ID),
		TransactionType:       string(photon.TransactionType_Auth),
		TransactionHappenedAt: req.CardTransaction.CreatedAt.UTC().Format(time.RFC3339),
		TransactionCurrency:   string(req.CardTransaction.TxCurrency),
		TransactionAmount:     req.CardTransaction.TxAmount.String(),
		AuthCurrency:          string(req.Authorization.Currency),
		AuthAmount:            req.Authorization.Amount.String(),
		FeeCurrency:           string(req.Authorization.Currency),
		FeeAmount:             "0",
		MCC:                   req.CardTransaction.MerchantMCC,
		MerchantName:          req.CardTransaction.MerchantName,
		TransactionCountry:    req.CardTransaction.MerchantCountry,
		AuthCode:              req.CardTransaction.AuthorizationCode,
	})
	if err != nil {
		zap.S().Errorw("marshal photonpay authorization request", "error", err)
		return nil, sharederrors.ErrAuthorizationRequestFailed
	}

	result, err := r.authorizationClient.RequestAuthorization(ctx, &AuthorizationRequestDeliveryRequest{
		TargetURL:     config.TargetURL,
		Payload:       payload,
		TimeoutMillis: config.TimeoutMillis,
	})
	if err != nil {
		zap.S().Errorw(
			"send photonpay authorization request",
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
			"photonpay authorization request rejected by http status",
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
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(result.ResponseBody), &response); err != nil {
		zap.S().Errorw(
			"decode photonpay authorization response",
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
	if response.Code == "" {
		zap.S().Errorw(
			"photonpay authorization response missing code",
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
		Approved: response.Code == string(photon.ResponseCode_Success),
		Reason:   response.Code,
	}, nil
}
