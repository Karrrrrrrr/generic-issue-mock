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

type photonPayCardStatusPayload struct {
	CardID     string `json:"cardId"`
	CardStatus string `json:"cardStatus"`
}

type photonPayDispatchPayloadRequest struct {
	AccountID   model.ID
	SourceID    model.ID
	Category    photon.WebhookNotificationCategory
	Event       photon.WebhookEvent
	Payload     []byte
	PublishedAt string
}

var _ sharedbiz.Notificator = (*PhotonPayWebhookNotificator)(nil)

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

func (n *PhotonPayWebhookNotificator) ReplayWebhookRecord(
	ctx context.Context,
	id model.ID,
) (*model.WebhookRecord, error) {
	exists, err := n.webhookRecordRepo.Exist(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay webhook record", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}

	original, err := n.webhookRecordRepo.Find(ctx, id)
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
	if err := n.webhookRecordRepo.Create(ctx, replay); err != nil {
		zap.S().Errorw("create photonpay webhook replay record", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}

	n.deliverWebhookRecord(ctx, replay, nil)
	if err := n.webhookRecordRepo.Save(ctx, replay); err != nil {
		zap.S().Errorw("save photonpay webhook replay record", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return replay, nil
}

func (n *PhotonPayWebhookNotificator) NotifyIssueCard(ctx context.Context, req *sharedbiz.NotifyIssueCardReq) error {
	if req == nil || req.Channel != enums.Channel_PhotonPay || req.AccountID <= 0 || req.CardID <= 0 {
		return photonpayerrors.ErrInvalidOperation
	}

	card, err := n.findNotificationCard(ctx, req.AccountID, req.CardID)
	if err != nil {
		return err
	}
	payload, err := n.photonPayCardStatusWebhookPayload(card, card.Status)
	if err != nil {
		zap.S().Errorw("marshal photonpay card status webhook payload", "error", err)
		return err
	}
	n.dispatchPayload(ctx, &photonPayDispatchPayloadRequest{
		AccountID:   card.AccountID,
		SourceID:    card.ID,
		Category:    photon.WebhookNotificationCategoryIssuingCard,
		Event:       photon.WebhookEventCardStatusUpdate,
		Payload:     payload,
		PublishedAt: card.UpdatedAt.UTC().Format(time.RFC3339),
	})
	return nil
}

func (n *PhotonPayWebhookNotificator) NotifyCardStatus(ctx context.Context, req *sharedbiz.NotifyCardStatusReq) error {
	if req == nil || req.Channel != enums.Channel_PhotonPay || req.AccountID <= 0 || req.CardID <= 0 {
		return photonpayerrors.ErrInvalidOperation
	}

	card, err := n.findNotificationCard(ctx, req.AccountID, req.CardID)
	if err != nil {
		return err
	}
	status := req.Status
	if status == "" {
		status = card.Status
	}
	payload, err := n.photonPayCardStatusWebhookPayload(card, status)
	if err != nil {
		zap.S().Errorw("marshal photonpay card status webhook payload", "error", err)
		return err
	}
	n.dispatchPayload(ctx, &photonPayDispatchPayloadRequest{
		AccountID:   card.AccountID,
		SourceID:    card.ID,
		Category:    photon.WebhookNotificationCategoryIssuingCard,
		Event:       photon.WebhookEventCardStatusUpdate,
		Payload:     payload,
		PublishedAt: time.Now().UTC().Format(time.RFC3339),
	})
	return nil
}

func (n *PhotonPayWebhookNotificator) NotifyCardTransaction(ctx context.Context, req *sharedbiz.NotifyCardTransactionReq) error {
	if req == nil || req.Channel != enums.Channel_PhotonPay || req.AccountID <= 0 || req.CardTransactionID <= 0 {
		return photonpayerrors.ErrInvalidOperation
	}

	transaction, err := n.cardTransactionRepo.FindByAccountID(ctx, &CardTransactionFindByAccountIDRequest{
		AccountID: &req.AccountID,
		ID:        req.CardTransactionID,
	})
	if err != nil {
		zap.S().Errorw(
			"find photonpay notification transaction",
			"account_id",
			req.AccountID,
			"transaction_id",
			req.CardTransactionID,
			"error",
			err,
		)
		return err
	}
	payload, err := n.photonPayWebhookPayload(ctx, transaction)
	if err != nil {
		zap.S().Errorw("marshal photonpay webhook payload", "error", err)
		return err
	}
	n.dispatchPayload(ctx, &photonPayDispatchPayloadRequest{
		AccountID:   transaction.AccountID,
		SourceID:    transaction.ID,
		Category:    photonPayWebhookCategory(transaction),
		Event:       photon.ConvertGenericTransactionTypeToWebhookEvent(transaction.Type),
		Payload:     payload,
		PublishedAt: transaction.CreatedAt.UTC().Format(time.RFC3339),
	})
	return nil
}

func (n *PhotonPayWebhookNotificator) dispatchPayload(ctx context.Context, req *photonPayDispatchPayloadRequest) {
	configs, err := n.webhookRepo.List(ctx, &WebhookConfigListRequest{
		AccountIDs: []model.ID{req.AccountID},
	})
	if err != nil {
		zap.S().Errorw("list photonpay webhook configs", "error", err)
		return
	}
	for _, config := range configs {
		if !config.Enabled || config.Event != string(req.Event) {
			continue
		}
		record := &model.WebhookRecord{
			WebhookConfigID: config.ID,
			AccountID:       config.AccountID,
			Channel:         enums.Channel_PhotonPay,
			Event:           string(req.Event),
			TargetURL:       config.TargetURL,
			SourceID:        strconv.FormatInt(int64(req.SourceID), 10),
			Payload:         req.Payload,
			Status:          enums.WebhookDeliveryStatus_Pending,
			AttemptCount:    1,
		}
		if err := n.webhookRecordRepo.Create(ctx, record); err != nil {
			zap.S().Errorw("create photonpay webhook record", "error", err)
			continue
		}
		n.deliverWebhookRecord(ctx, record, req)
		if err := n.webhookRecordRepo.Save(ctx, record); err != nil {
			zap.S().Errorw("save photonpay webhook record", "error", err)
		}
	}
}

func (n *PhotonPayWebhookNotificator) deliverWebhookRecord(
	ctx context.Context,
	record *model.WebhookRecord,
	req *photonPayDispatchPayloadRequest,
) {
	startedAt := time.Now()
	zap.S().Infow(
		"webhook delivery started",
		"channel",
		record.Channel,
		"account_id",
		record.AccountID,
		"webhook_record_id",
		record.ID,
		"event",
		record.Event,
		"source_id",
		record.SourceID,
		"attempt",
		record.AttemptCount,
		"method",
		"POST",
		"url",
		record.TargetURL,
		"request_body",
		string(record.Payload),
	)
	deliveryReq := &PhotonPayWebhookDeliveryRequest{
		TargetURL:      record.TargetURL,
		Payload:        record.Payload,
		RequestHeaders: record.RequestHeaders,
	}
	if req != nil {
		deliveryReq.NotifyCategory = string(req.Category)
		deliveryReq.NotifyType = string(req.Event)
		deliveryReq.PublishedAt = req.PublishedAt
	}
	result, deliveryErr := n.webhookClient.Deliver(ctx, deliveryReq)
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
		if result != nil && result.StatusCode >= 200 && result.StatusCode < 300 && photonPayWebhookAcknowledged(result.ResponseBody) {
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
}

func (n *PhotonPayWebhookNotificator) findNotificationCard(
	ctx context.Context,
	accountID model.ID,
	cardID model.ID,
) (*model.Card, error) {
	card, err := n.cardRepo.FindCardByAccountID(ctx, &CardFindCardByAccountIDRequest{
		AccountID: &accountID,
		ID:        cardID,
	})
	if err != nil {
		zap.S().Errorw(
			"find photonpay notification card",
			"account_id",
			accountID,
			"card_id",
			cardID,
			"error",
			err,
		)
		return nil, err
	}
	return card, nil
}

func (n *PhotonPayWebhookNotificator) photonPayCardStatusWebhookPayload(
	card *model.Card,
	status enums.CardStatus,
) ([]byte, error) {
	payload := photonPayCardStatusPayload{
		CardID:     strconv.FormatInt(int64(card.ID), 10),
		CardStatus: string(photon.ConvertGenericCardStatusToCardStatus(status)),
	}
	return json.Marshal(payload)
}

func (n *PhotonPayWebhookNotificator) photonPayWebhookPayload(ctx context.Context, transaction *model.CardTransaction) ([]byte, error) {
	card, err := n.cardRepo.FindCardByID(ctx, transaction.CardID)
	if err != nil {
		zap.S().Errorw("find photonpay webhook card", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	payload := photonPayEventPayload{
		MemberID:                   photon.MemberID,
		CreatedAt:                  transaction.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                  transaction.UpdatedAt.UTC().Format(time.RFC3339),
		CardID:                     strconv.FormatInt(int64(card.ID), 10),
		CardType:                   string(photon.ConvertGenericCardTypeToCardType(card.CardType)),
		TransactionID:              strconv.FormatInt(int64(transaction.ID), 10),
		RequestID:                  transaction.RequestID,
		TransactionType:            string(photon.ConvertGenericTransactionTypeToWebhookEvent(transaction.Type)),
		Status:                     string(photon.ConvertGenericTransactionStatusToTransactionStatus(transaction.Status)),
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
		TransactionStatus:          string(photon.ConvertGenericTransactionStatusToTransactionStatus(transaction.Status)),
		TransactionCountry:         transaction.MerchantCountry,
		TransactionHappenedAt:      transaction.CreatedAt.UTC().Format(time.RFC3339),
	}
	if transaction.OriginCardTransactionID != 0 {
		payload.OriginTransactionID = strconv.FormatInt(int64(transaction.OriginCardTransactionID), 10)
	}
	if transaction.Type == enums.CardTransactionType_CLEAR {
		payload.SettleAmount = transaction.TxAmount.String()
		payload.SettleCurrency = string(transaction.TxCurrency)
	}
	return json.Marshal(payload)
}
