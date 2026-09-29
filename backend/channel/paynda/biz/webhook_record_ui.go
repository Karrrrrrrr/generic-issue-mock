package biz

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	paynda "generic-mock/channel/paynda/enums"
	payndaerrors "generic-mock/channel/paynda/errors"
	"generic-mock/enums"
	"generic-mock/model"
	sharedbiz "generic-mock/shared/biz"

	"go.uber.org/zap"
)

type payndaEventPayload struct {
	CardTransactionWebhook payndaCardTransactionWebhook `json:"cardTransactionWebhook"`
}

type payndaCardTransactionWebhook struct {
	ID                                  string `json:"id"`
	CreateTime                          string `json:"createTime"`
	UpdateTime                          string `json:"updateTime"`
	MerchantID                          int64  `json:"merchantId"`
	BalanceAccountID                    int64  `json:"balanceAccountId"`
	CardholderID                        int64  `json:"cardholderId"`
	CardID                              int64  `json:"cardId"`
	MaskCardNo                          string `json:"maskCardNo"`
	Type                                string `json:"type"`
	ApprovalCode                        string `json:"approvalCode,omitempty"`
	PreAuthAmount                       string `json:"preAuthAmount"`
	PostedAmount                        string `json:"postedAmount"`
	Currency                            string `json:"currency"`
	OriginalCurrencyCode                string `json:"originalCurrencyCode"`
	TransactionAmountInOriginalCurrency string `json:"transactionAmountInOriginalCurrency"`
	ReversalFlag                        string `json:"reversalFlag"`
	TransactionTime                     string `json:"transactionTime"`
	AuthorizationTime                   string `json:"authorizationTime"`
	AcquirerID                          string `json:"acquirerId"`
	MerchantMCC                         string `json:"merchantMcc"`
	MerchantName                        string `json:"merchantName"`
	MerchantAddressAddressLine1         string `json:"merchantAddressAddressLine1"`
	MerchantAddressCity                 string `json:"merchantAddressCity"`
	MerchantAddressState                string `json:"merchantAddressState"`
	MerchantAddressCountry              string `json:"merchantAddressCountry"`
	MerchantAddressZIP                  string `json:"merchantAddressZip"`
	POSAcceptorID                       string `json:"posAcceptorId"`
	POSAcceptLocation                   string `json:"posAcceptLocation"`
	POSEntryDescription                 string `json:"posEntryDescription"`
	TransactionID                       string `json:"transactionId"`
	SupplierTransactionID               string `json:"supplierTransactionId"`
	SupplierTransactionLinkID           string `json:"supplierTransactionLinkId"`
	DeclineMessage                      string `json:"declineMessage"`
	WalletID                            string `json:"walletId"`
}

func (u *PayndaUIUsecase) ReplayWebhookRecord(
	ctx context.Context,
	id model.ID,
) (*model.WebhookRecord, error) {
	exists, err := u.webhookRecordRepository.Exist(ctx, id)
	if err != nil {
		zap.S().Errorw("check paynda webhook record", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, payndaerrors.ErrResourceNotFound
	}

	original, err := u.webhookRecordRepository.Find(ctx, id)
	if err != nil {
		zap.S().Errorw("find paynda webhook record", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	replay := &model.WebhookRecord{
		Account:         original.Account,
		WebhookConfigID: original.WebhookConfigID,
		AccountID:       original.AccountID,
		Channel:         enums.Channel_Paynda,
		Event:           original.Event,
		TargetURL:       original.TargetURL,
		SourceID:        original.SourceID,
		Payload:         original.Payload,
		RequestHeaders:  original.RequestHeaders,
		Status:          enums.WebhookDeliveryStatus_Pending,
		AttemptCount:    original.AttemptCount + 1,
	}
	if err := u.webhookRecordRepository.Create(ctx, replay); err != nil {
		zap.S().Errorw("create paynda webhook replay record", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
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
	result, deliveryErr := u.webhookClient.Deliver(ctx, &PayndaWebhookDeliveryRequest{
		TargetURL:      replay.TargetURL,
		Payload:        replay.Payload,
		RequestHeaders: replay.RequestHeaders,
		Category:       replay.Event,
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
		if result.StatusCode >= 200 && result.StatusCode < 300 {
			deliveredAt := time.Now().UTC()
			replay.Status = enums.WebhookDeliveryStatus_Succeeded
			replay.DeliveredAt = &deliveredAt
		} else {
			replay.Status = enums.WebhookDeliveryStatus_Failed
			replay.ErrorMessage = "unexpected Paynda webhook response"
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
	if err := u.webhookRecordRepository.Save(ctx, replay); err != nil {
		zap.S().Errorw("save paynda webhook replay record", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return replay, nil
}

var _ sharedbiz.Notificator = (*PayndaUIUsecase)(nil)

func (u *PayndaUIUsecase) NotifyCardTransaction(ctx context.Context, req *sharedbiz.NotifyCardTransactionReq) error {
	if req == nil || req.Channel != enums.Channel_Paynda || req.AccountID <= 0 || req.CardTransactionID <= 0 {
		return payndaerrors.ErrInvalidOperation
	}
	transaction, err := u.cardTransactionRepository.FindByAccountID(ctx, &CardTransactionFindByAccountIDRequest{
		AccountID: req.AccountID,
		ID:        req.CardTransactionID,
	})
	if err != nil {
		zap.S().Errorw("find paynda notification transaction", "account_id", req.AccountID, "transaction_id", req.CardTransactionID, "error", err)
		return err
	}
	u.dispatch(ctx, paynda.WebhookTypeCardTransaction, transaction.ID, transaction)
	return nil
}

func (u *PayndaUIUsecase) dispatch(
	ctx context.Context,
	webhookType paynda.WebhookType,
	sourceID model.ID,
	transaction *model.CardTransaction,
) {
	card, err := u.cardRepository.FindByID(ctx, &CardFindByIDRequest{
		AccountID: &transaction.AccountID,
		ID:        transaction.CardID,
	})
	if err != nil {
		zap.S().Errorw("find paynda webhook card for account", "error", err)
		return
	}
	if card.AccountID == 0 {
		return
	}
	payload, err := u.payndaWebhookPayload(ctx, transaction)
	if err != nil {
		zap.S().Errorw("marshal paynda webhook payload", "error", err)
		return
	}
	configs, err := u.webhookConfigRepository.ListByAccountIDs(ctx, &WebhookConfigListByAccountIDsRequest{
		AccountIDs: []model.ID{card.AccountID},
	})
	if err != nil {
		zap.S().Errorw("list paynda webhook configs", "error", err)
		return
	}
	for _, config := range configs {
		if !config.Enabled || config.Event != string(webhookType) {
			continue
		}
		record := &model.WebhookRecord{
			WebhookConfigID: config.ID,
			AccountID:       config.AccountID,
			Channel:         enums.Channel_Paynda,
			Event:           string(webhookType),
			TargetURL:       config.TargetURL,
			SourceID:        strconv.FormatInt(int64(sourceID), 10),
			Payload:         payload,
			Status:          enums.WebhookDeliveryStatus_Pending,
			AttemptCount:    1,
		}
		if err := u.webhookRecordRepository.Create(ctx, record); err != nil {
			zap.S().Errorw("create paynda webhook record", "error", err)
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
		result, deliveryErr := u.webhookClient.Deliver(ctx, &PayndaWebhookDeliveryRequest{
			TargetURL: config.TargetURL,
			Payload:   payload,
			Category:  string(webhookType),
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
			if result.StatusCode >= 200 && result.StatusCode < 300 {
				deliveredAt := time.Now().UTC()
				record.Status = enums.WebhookDeliveryStatus_Succeeded
				record.DeliveredAt = &deliveredAt
			} else {
				record.Status = enums.WebhookDeliveryStatus_Failed
				record.ErrorMessage = "unexpected Paynda webhook response"
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
		if err := u.webhookRecordRepository.Save(ctx, record); err != nil {
			zap.S().Errorw("save paynda webhook record", "error", err)
		}
	}
}

func (u *PayndaUIUsecase) payndaWebhookPayload(ctx context.Context, transaction *model.CardTransaction) ([]byte, error) {
	card, err := u.cardRepository.FindByID(ctx, &CardFindByIDRequest{
		AccountID: &transaction.AccountID,
		ID:        transaction.CardID,
	})
	if err != nil {
		zap.S().Errorw("find paynda webhook card", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	holder, err := u.cardHolderRepository.FindByID(ctx, card.CardHolderID)
	if err != nil {
		zap.S().Errorw("find paynda webhook card holder", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if card.WalletID == 0 {
		return nil, payndaerrors.ErrResourceNotFound
	}
	wallet, err := u.walletRepository.FindByID(ctx, &WalletFindByIDRequest{
		AccountID: &card.AccountID,
		ID:        card.WalletID,
	})
	if err != nil {
		zap.S().Errorw("find paynda webhook wallet", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	account, err := u.accountRepository.FindByChannel(ctx)
	if err != nil {
		zap.S().Errorw("find paynda webhook account", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	authorizationTime := ""
	if transaction.AuthorizationID != 0 {
		authorization, err := u.authorizationRepository.FindByID(ctx, &PayndaFindAuthorizationRequest{
			ID: transaction.AuthorizationID,
		})
		if err != nil {
			zap.S().Errorw("find paynda webhook transaction authorization", "error", err)
			return nil, payndaerrors.ErrDatabaseOperation
		}
		authorizationTime = authorization.CreatedAt.UTC().Format(time.RFC3339)
	}
	transactionID := strconv.FormatInt(transaction.ID, 10)
	payload := payndaEventPayload{CardTransactionWebhook: payndaCardTransactionWebhook{
		ID:               transactionID,
		CreateTime:       transaction.CreatedAt.UTC().Format(time.RFC3339),
		UpdateTime:       transaction.UpdatedAt.UTC().Format(time.RFC3339),
		MerchantID:       account.ID,
		BalanceAccountID: account.ID,
		CardholderID:     holder.ID,
		CardID:           card.ID,
		MaskCardNo:       card.MaskedNumber(),
		Type: string(paynda.ConvertGenericCardTransactionToTransactionType(
			transaction.Type,
			transaction.Status,
		)),
		ApprovalCode:                        transaction.AuthorizationCode,
		PreAuthAmount:                       transaction.TxAmount.String(),
		PostedAmount:                        transaction.TxAmount.String(),
		Currency:                            string(transaction.TxCurrency),
		OriginalCurrencyCode:                string(transaction.TxCurrency),
		TransactionAmountInOriginalCurrency: transaction.TxAmount.String(),
		ReversalFlag:                        strconv.FormatBool(transaction.Type == enums.CardTransactionType_VOID),
		TransactionTime:                     transaction.CreatedAt.UTC().Format(time.RFC3339),
		AuthorizationTime:                   authorizationTime,
		MerchantMCC:                         transaction.MerchantMCC,
		MerchantName:                        transaction.MerchantName,
		MerchantAddressCountry:              transaction.MerchantCountry,
		TransactionID:                       transactionID,
		SupplierTransactionID:               transactionID,
		WalletID:                            strconv.FormatInt(wallet.ID, 10),
	}}
	if transaction.AuthorizationID != 0 {
		payload.CardTransactionWebhook.SupplierTransactionLinkID = strconv.FormatInt(transaction.AuthorizationID, 10)
	}
	return json.Marshal(payload)
}
