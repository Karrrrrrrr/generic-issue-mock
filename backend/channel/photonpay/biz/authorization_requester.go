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

type photonPayAuthorizationFailureRequest struct {
	Request   *sharedbiz.AuthorizationRequest
	Side      sharedbiz.AuthorizationFailureSide
	Message   string
	Reason    string
	TargetURL string
	Payload   []byte
	Result    *AuthorizationRequestDeliveryResult
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
		return nil, r.authorizationFailure(&photonPayAuthorizationFailureRequest{
			Request: req,
			Side:    sharedbiz.AuthorizationFailureSideMock,
			Message: "PhotonPay authorization config missing",
		})
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
		return nil, r.authorizationFailure(&photonPayAuthorizationFailureRequest{
			Request: req,
			Side:    sharedbiz.AuthorizationFailureSideMock,
			Message: "PhotonPay authorization config disabled or target URL is empty",
		})
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
		return nil, r.authorizationFailure(&photonPayAuthorizationFailureRequest{
			Request: req,
			Side:    sharedbiz.AuthorizationFailureSideMock,
			Message: "marshal PhotonPay authorization request failed",
		})
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
		return nil, r.authorizationFailure(&photonPayAuthorizationFailureRequest{
			Request:   req,
			Side:      sharedbiz.AuthorizationFailureSideThirdParty,
			Message:   "send PhotonPay authorization request failed",
			TargetURL: config.TargetURL,
			Payload:   payload,
			Result:    result,
		})
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
		return nil, r.authorizationFailure(&photonPayAuthorizationFailureRequest{
			Request:   req,
			Side:      sharedbiz.AuthorizationFailureSideThirdParty,
			Message:   "PhotonPay authorization endpoint returned non-success status",
			TargetURL: config.TargetURL,
			Payload:   payload,
			Result:    result,
		})
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
		return nil, r.authorizationFailure(&photonPayAuthorizationFailureRequest{
			Request:   req,
			Side:      sharedbiz.AuthorizationFailureSideThirdParty,
			Message:   "decode PhotonPay authorization response failed",
			TargetURL: config.TargetURL,
			Payload:   payload,
			Result:    result,
		})
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
		return nil, r.authorizationFailure(&photonPayAuthorizationFailureRequest{
			Request:   req,
			Side:      sharedbiz.AuthorizationFailureSideThirdParty,
			Message:   "PhotonPay authorization response missing code",
			TargetURL: config.TargetURL,
			Payload:   payload,
			Result:    result,
		})
	}

	message := photon.ConvertAuthorizationResultCodeToMessage(response.Code)
	return &sharedbiz.AuthorizationDecision{
		Approved:    response.Code == string(photon.AuthorizationResultCodePass),
		FailureSide: sharedbiz.AuthorizationFailureSideThirdParty,
		Reason:      response.Code,
		Message:     message,
		Exchange: r.authorizationExchange(&photonPayAuthorizationFailureRequest{
			TargetURL: config.TargetURL,
			Payload:   payload,
			Result:    result,
		}),
	}, nil
}

func (r *PhotonPayAuthorizationRequester) authorizationFailure(
	req *photonPayAuthorizationFailureRequest,
) *sharedbiz.AuthorizationFailureError {
	return sharedbiz.NewAuthorizationFailureError(
		&sharedbiz.AuthorizationResult{
			Attempted:       req.Payload != nil,
			Approved:        false,
			FailureSide:     req.Side,
			Reason:          req.Reason,
			Message:         req.Message,
			AccountID:       req.Request.AccountID,
			CardID:          req.Request.Card.ID,
			Amount:          req.Request.Authorization.Amount,
			Currency:        req.Request.Authorization.Currency,
			MerchantName:    req.Request.CardTransaction.MerchantName,
			MerchantCountry: req.Request.CardTransaction.MerchantCountry,
			MerchantMCC:     req.Request.CardTransaction.MerchantMCC,
			Exchange:        r.authorizationExchange(req),
		},
		sharederrors.ErrAuthorizationRequestFailed,
	)
}

func (r *PhotonPayAuthorizationRequester) authorizationExchange(
	req *photonPayAuthorizationFailureRequest,
) *sharedbiz.AuthorizationExchange {
	if req.Payload == nil && req.Result == nil {
		return nil
	}
	exchange := &sharedbiz.AuthorizationExchange{
		TargetURL:      req.TargetURL,
		RequestPayload: req.Payload,
	}
	if req.Result == nil {
		return exchange
	}
	exchange.StatusCode = req.Result.StatusCode
	exchange.RequestHeaders = req.Result.RequestHeaders
	exchange.ResponseBody = req.Result.ResponseBody
	exchange.ResponseHeaders = req.Result.ResponseHeaders
	return exchange
}
