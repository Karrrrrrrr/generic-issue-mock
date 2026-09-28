package biz

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/enums"
	"generic-mock/model"
	sharedbiz "generic-mock/shared/biz"

	"go.uber.org/zap"
)

type photonPayEventPayload struct {
	MemberID                   string `json:"memberId"`
	MatrixAccount              string `json:"matrixAccount"`
	CreatedAt                  string `json:"createdAt"`
	UpdatedAt                  string `json:"updatedAt"`
	CardID                     string `json:"cardId"`
	CardType                   string `json:"cardType"`
	TransactionID              string `json:"transactionId"`
	OriginTransactionID        string `json:"originTransactionId,omitempty"`
	RequestID                  string `json:"requestId,omitempty"`
	TransactionType            string `json:"transactionType"`
	Status                     string `json:"status"`
	Code                       string `json:"code"`
	Message                    string `json:"msg"`
	MCC                        string `json:"mcc,omitempty"`
	AuthCode                   string `json:"authCode,omitempty"`
	TransactionAmount          string `json:"transactionAmount"`
	TransactionCurrency        string `json:"transactionCurrency"`
	TxnPrincipalChangeAccount  string `json:"txnPrincipalChangeAccount"`
	TxnPrincipalChangeAmount   string `json:"txnPrincipalChangeAmount"`
	TxnPrincipalChangeCurrency string `json:"txnPrincipalChangeCurrency"`
	MerchantName               string `json:"merchantName,omitempty"`
	MerchantLocation           string `json:"merchantLocation,omitempty"`
	TransactionStatus          string `json:"transactionStatus"`
	TransactionCountry         string `json:"transactionCountry,omitempty"`
	TransactionHappenedAt      string `json:"transactionHappenedAt"`
	SettleAmount               string `json:"settleAmount,omitempty"`
	SettleCurrency             string `json:"settleCurrency,omitempty"`
}

func photonPayWebhookCategory(transaction *model.CardTransaction) photon.WebhookNotificationCategory {
	if transaction.Status == enums.TransactionStatus_SUCCEED {
		return photon.WebhookNotificationCategoryIssuingSettlement
	}
	return photon.WebhookNotificationCategoryIssuing
}

func photonPayWebhookAcknowledged(body string) bool {
	var response struct {
		Roger bool `json:"roger"`
	}
	return json.Unmarshal([]byte(body), &response) == nil && response.Roger
}

func (u *PhotonPayUIUsecase) ReplayWebhookRecord(
	ctx context.Context,
	id model.ID,
) (*model.WebhookRecord, error) {
	exists, err := u.webhookRecordRepo.Exist(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay webhook record", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}

	original, err := u.webhookRecordRepo.Find(ctx, id)
	if err != nil {
		zap.S().Errorw("find photonpay webhook record", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	replay := &model.WebhookRecord{
		Account:         original.Account,
		WebhookConfigID: original.WebhookConfigID,
		AccountID:       original.AccountID,
		Channel:         enums.Channel_PhotonPay,
		Event:           original.Event,
		TargetURL:       original.TargetURL,
		SourceID:        original.SourceID,
		Payload:         original.Payload,
		RequestHeaders:  original.RequestHeaders,
		Status:          enums.WebhookDeliveryStatus_Pending,
		AttemptCount:    original.AttemptCount + 1,
	}
	if err := u.webhookRecordRepo.Create(ctx, replay); err != nil {
		zap.S().Errorw("create photonpay webhook replay record", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}

	startedAt := time.Now()
	zap.S().Infow("webhook delivery started",
		"channel", replay.Channel,
		"account_id", replay.AccountID,
		"webhook_record_id", replay.ID,
		"event", replay.Event,
		"source_id", replay.SourceID,
		"attempt", replay.AttemptCount,
		"method", "POST",
		"url", replay.TargetURL,
		"request_body", string(replay.Payload),
	)
	result, deliveryErr := u.webhookClient.Deliver(ctx, &PhotonPayWebhookDeliveryRequest{
		TargetURL:      replay.TargetURL,
		Payload:        replay.Payload,
		RequestHeaders: replay.RequestHeaders,
	})
	if result != nil {
		replay.StatusCode = result.StatusCode
		replay.ResponseBody = result.ResponseBody
		replay.RequestHeaders = result.RequestHeaders
		replay.ResponseHeaders = result.ResponseHeaders
	}
	if deliveryErr != nil {
		replay.Status = enums.WebhookDeliveryStatus_Failed
		replay.ErrorMessage = deliveryErr.Error()
	} else {
		if result.StatusCode >= 200 && result.StatusCode < 300 && photonPayWebhookAcknowledged(result.ResponseBody) {
			deliveredAt := time.Now().UTC()
			replay.Status = enums.WebhookDeliveryStatus_Succeeded
			replay.DeliveredAt = &deliveredAt
		} else {
			replay.Status = enums.WebhookDeliveryStatus_Failed
			replay.ErrorMessage = "unexpected PhotonPay webhook response"
		}
	}
	logFields := []any{
		"channel", replay.Channel,
		"account_id", replay.AccountID,
		"webhook_record_id", replay.ID,
		"event", replay.Event,
		"source_id", replay.SourceID,
		"attempt", replay.AttemptCount,
		"method", "POST",
		"url", replay.TargetURL,
		"status", replay.Status,
		"status_code", replay.StatusCode,
		"request_headers", string(replay.RequestHeaders),
		"response_headers", string(replay.ResponseHeaders),
		"response_body", replay.ResponseBody,
		"duration_ms", time.Since(startedAt).Milliseconds(),
		"error", replay.ErrorMessage,
	}
	if replay.Status == enums.WebhookDeliveryStatus_Failed {
		zap.S().Errorw("webhook delivery failed", logFields...)
	} else {
		zap.S().Infow("webhook delivery succeeded", logFields...)
	}
	if err := u.webhookRecordRepo.Save(ctx, replay); err != nil {
		zap.S().Errorw("save photonpay webhook replay record", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return replay, nil
}

var _ sharedbiz.CardTransactionNotificator = (*PhotonPayUIUsecase)(nil)

func (u *PhotonPayUIUsecase) NotifyCardTransaction(ctx context.Context, req *sharedbiz.NotifyCardTransactionReq) error {
	if req == nil || req.Channel != enums.Channel_PhotonPay || req.AccountID <= 0 || req.CardTransactionID <= 0 {
		return photonpayerrors.ErrInvalidOperation
	}
	transaction, err := u.cardTransactionRepo.FindByAccountID(ctx, &CardTransactionFindByAccountIDRequest{
		AccountID: &req.AccountID,
		ID:        req.CardTransactionID,
	})
	if err != nil {
		return err
	}
	u.dispatch(ctx, photon.WebhookEventFromGenericTransactionType(transaction.Type), transaction.ID, transaction)
	return nil
}

func (u *PhotonPayUIUsecase) dispatch(ctx context.Context, event photon.WebhookEvent, sourceID model.ID, transaction *model.CardTransaction) {
	payload, err := u.photonPayWebhookPayload(ctx, transaction)
	if err != nil {
		zap.S().Errorw("marshal photonpay webhook payload", "error", err)
		return
	}
	configs, err := u.webhookRepo.List(ctx, &WebhookConfigListRequest{
		AccountIDs: []model.ID{transaction.AccountID},
	})
	if err != nil {
		zap.S().Errorw("list photonpay webhook configs", "error", err)
		return
	}
	for _, config := range configs {
		if !config.Enabled || config.Event != string(event) {
			continue
		}
		record := &model.WebhookRecord{
			WebhookConfigID: config.ID,
			AccountID:       config.AccountID,
			Channel:         enums.Channel_PhotonPay,
			Event:           string(event),
			TargetURL:       config.TargetURL,
			SourceID:        strconv.FormatInt(int64(sourceID), 10),
			Payload:         payload,
			Status:          enums.WebhookDeliveryStatus_Pending,
			AttemptCount:    1,
		}
		if err := u.webhookRecordRepo.Create(ctx, record); err != nil {
			zap.S().Errorw("create photonpay webhook record", "error", err)
			continue
		}
		startedAt := time.Now()
		zap.S().Infow("webhook delivery started",
			"channel", record.Channel,
			"account_id", record.AccountID,
			"webhook_record_id", record.ID,
			"event", record.Event,
			"source_id", record.SourceID,
			"attempt", record.AttemptCount,
			"method", "POST",
			"url", record.TargetURL,
			"request_body", string(record.Payload),
		)
		result, deliveryErr := u.webhookClient.Deliver(ctx, &PhotonPayWebhookDeliveryRequest{
			TargetURL:      config.TargetURL,
			Payload:        payload,
			NotifyCategory: string(photonPayWebhookCategory(transaction)),
			NotifyType:     string(event),
			PublishedAt:    transaction.CreatedAt.UTC().Format(time.RFC3339),
		})
		if result != nil {
			record.StatusCode = result.StatusCode
			record.ResponseBody = result.ResponseBody
			record.RequestHeaders = result.RequestHeaders
			record.ResponseHeaders = result.ResponseHeaders
		}
		if deliveryErr != nil {
			record.Status = enums.WebhookDeliveryStatus_Failed
			record.ErrorMessage = deliveryErr.Error()
		} else {
			if result.StatusCode >= 200 && result.StatusCode < 300 && photonPayWebhookAcknowledged(result.ResponseBody) {
				deliveredAt := time.Now().UTC()
				record.Status = enums.WebhookDeliveryStatus_Succeeded
				record.DeliveredAt = &deliveredAt
			} else {
				record.Status = enums.WebhookDeliveryStatus_Failed
				record.ErrorMessage = "unexpected PhotonPay webhook response"
			}
		}
		logFields := []any{
			"channel", record.Channel,
			"account_id", record.AccountID,
			"webhook_record_id", record.ID,
			"event", record.Event,
			"source_id", record.SourceID,
			"attempt", record.AttemptCount,
			"method", "POST",
			"url", record.TargetURL,
			"status", record.Status,
			"status_code", record.StatusCode,
			"request_headers", string(record.RequestHeaders),
			"response_headers", string(record.ResponseHeaders),
			"response_body", record.ResponseBody,
			"duration_ms", time.Since(startedAt).Milliseconds(),
			"error", record.ErrorMessage,
		}
		if record.Status == enums.WebhookDeliveryStatus_Failed {
			zap.S().Errorw("webhook delivery failed", logFields...)
		} else {
			zap.S().Infow("webhook delivery succeeded", logFields...)
		}
		if err := u.webhookRecordRepo.Save(ctx, record); err != nil {
			zap.S().Errorw("save photonpay webhook record", "error", err)
		}
	}
}

func (u *PhotonPayUIUsecase) photonPayWebhookPayload(ctx context.Context, transaction *model.CardTransaction) ([]byte, error) {
	card, err := u.cardRepo.FindCardByID(ctx, transaction.CardID)
	if err != nil {
		zap.S().Errorw("find photonpay webhook card", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	payload := photonPayEventPayload{
		MemberID:                   photon.MemberID,
		CreatedAt:                  transaction.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                  transaction.UpdatedAt.UTC().Format(time.RFC3339),
		CardID:                     strconv.FormatInt(card.ID, 10),
		CardType:                   string(photon.CardTypeFromGeneric(card.CardType)),
		TransactionID:              strconv.FormatInt(transaction.ID, 10),
		RequestID:                  transaction.RequestID,
		TransactionType:            string(photon.WebhookEventFromGenericTransactionType(transaction.Type)),
		Status:                     string(photon.TransactionStatusFromGeneric(transaction.Status)),
		Code:                       photon.WebhookSuccessCode,
		Message:                    photon.WebhookSuccessMessage,
		MCC:                        transaction.MerchantMCC,
		AuthCode:                   transaction.AuthorizationCode,
		TransactionAmount:          transaction.TxAmount.String(),
		TransactionCurrency:        string(transaction.TxCurrency),
		TxnPrincipalChangeAccount:  photon.WebhookBalanceAccountCard,
		TxnPrincipalChangeAmount:   transaction.TxAmount.String(),
		TxnPrincipalChangeCurrency: string(transaction.TxCurrency),
		MerchantName:               transaction.MerchantName,
		MerchantLocation:           transaction.MerchantCountry,
		TransactionStatus:          string(photon.TransactionStatusFromGeneric(transaction.Status)),
		TransactionCountry:         transaction.MerchantCountry,
		TransactionHappenedAt:      transaction.CreatedAt.UTC().Format(time.RFC3339),
	}
	if transaction.OriginCardTransactionID != 0 {
		payload.OriginTransactionID = strconv.FormatInt(transaction.OriginCardTransactionID, 10)
	}
	if transaction.Type == enums.CardTransactionType_CLEAR {
		payload.SettleAmount = transaction.TxAmount.String()
		payload.SettleCurrency = string(transaction.TxCurrency)
	}
	return json.Marshal(payload)
}
