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

type payndaTransactionEventPayload struct {
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

type payndaCardStatusEventPayload struct {
	CardStatusWebhook payndaCardStatusWebhook `json:"cardStatusWebhook"`
}

type payndaCardStatusWebhook struct {
	ID               string `json:"id"`
	CreateTime       string `json:"createTime"`
	UpdateTime       string `json:"updateTime"`
	MerchantID       int64  `json:"merchantId"`
	BalanceAccountID int64  `json:"balanceAccountId"`
	CardholderID     int64  `json:"cardholderId"`
	CardID           int64  `json:"cardId"`
	MaskCardNo       string `json:"maskCardNo"`
	Status           string `json:"status"`
}

var _ sharedbiz.Notificator = (*PayndaWebhookNotificator)(nil)

func (n *PayndaWebhookNotificator) ReplayWebhookRecord(
	ctx context.Context,
	id model.ID,
) (*model.WebhookRecord, error) {
	exists, err := n.webhookRecordRepository.Exist(ctx, id)
	if err != nil {
		zap.S().Errorw("check paynda webhook record", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, payndaerrors.ErrResourceNotFound
	}

	original, err := n.webhookRecordRepository.Find(ctx, id)
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
	if err := n.webhookRecordRepository.Create(ctx, replay); err != nil {
		zap.S().Errorw("create paynda webhook replay record", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	n.deliverWebhookRecord(ctx, replay)
	if err := n.webhookRecordRepository.Save(ctx, replay); err != nil {
		zap.S().Errorw("save paynda webhook replay record", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return replay, nil
}

func (n *PayndaWebhookNotificator) NotifyIssueCard(ctx context.Context, req *sharedbiz.NotifyIssueCardReq) error {
	if req == nil || req.Channel != enums.Channel_Paynda || req.AccountID <= 0 || req.CardID <= 0 {
		return payndaerrors.ErrInvalidOperation
	}

	card, err := n.findNotificationCard(ctx, req.AccountID, req.CardID)
	if err != nil {
		return err
	}
	payload, err := n.payndaCardStatusWebhookPayload(ctx, card, card.Status)
	if err != nil {
		zap.S().Errorw("marshal paynda card status webhook payload", "error", err)
		return err
	}
	n.dispatchPayload(ctx, paynda.WebhookTypeCardStatus, card.AccountID, card.ID, payload)
	return nil
}

func (n *PayndaWebhookNotificator) NotifyCardStatus(ctx context.Context, req *sharedbiz.NotifyCardStatusReq) error {
	if req == nil || req.Channel != enums.Channel_Paynda || req.AccountID <= 0 || req.CardID <= 0 {
		return payndaerrors.ErrInvalidOperation
	}

	card, err := n.findNotificationCard(ctx, req.AccountID, req.CardID)
	if err != nil {
		return err
	}
	status := req.Status
	if status == "" {
		status = card.Status
	}
	payload, err := n.payndaCardStatusWebhookPayload(ctx, card, status)
	if err != nil {
		zap.S().Errorw("marshal paynda card status webhook payload", "error", err)
		return err
	}
	n.dispatchPayload(ctx, paynda.WebhookTypeCardStatus, card.AccountID, card.ID, payload)
	return nil
}

func (n *PayndaWebhookNotificator) NotifyCardTransaction(ctx context.Context, req *sharedbiz.NotifyCardTransactionReq) error {
	if req == nil || req.Channel != enums.Channel_Paynda || req.AccountID <= 0 || req.CardTransactionID <= 0 {
		return payndaerrors.ErrInvalidOperation
	}
	transaction, err := n.cardTransactionRepository.FindByAccountID(ctx, &CardTransactionFindByAccountIDRequest{
		AccountID: req.AccountID,
		ID:        req.CardTransactionID,
	})
	if err != nil {
		zap.S().Errorw(
			"find paynda notification transaction",
			"account_id",
			req.AccountID,
			"transaction_id",
			req.CardTransactionID,
			"error",
			err,
		)
		return err
	}
	payload, err := n.payndaTransactionWebhookPayload(ctx, transaction)
	if err != nil {
		zap.S().Errorw("marshal paynda transaction webhook payload", "error", err)
		return err
	}
	n.dispatchPayload(ctx, paynda.WebhookTypeCardTransaction, transaction.AccountID, transaction.ID, payload)
	return nil
}

func (n *PayndaWebhookNotificator) NotifyCardFunding(ctx context.Context, req *sharedbiz.NotifyCardFundingReq) error {
	if req == nil || req.Channel != enums.Channel_Paynda || req.AccountID <= 0 {
		return payndaerrors.ErrInvalidOperation
	}
	return nil
}

func (n *PayndaWebhookNotificator) findNotificationCard(
	ctx context.Context,
	accountID model.ID,
	cardID model.ID,
) (*model.Card, error) {
	card, err := n.cardRepository.FindByID(ctx, &CardFindByIDRequest{
		AccountID: &accountID,
		ID:        cardID,
	})
	if err != nil {
		zap.S().Errorw(
			"find paynda notification card",
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

func (n *PayndaWebhookNotificator) dispatchPayload(
	ctx context.Context,
	webhookType paynda.WebhookType,
	accountID model.ID,
	sourceID model.ID,
	payload []byte,
) {
	configs, err := n.webhookConfigRepository.ListByAccountIDs(ctx, &WebhookConfigListByAccountIDsRequest{
		AccountIDs: []model.ID{accountID},
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
		if err := n.webhookRecordRepository.Create(ctx, record); err != nil {
			zap.S().Errorw("create paynda webhook record", "error", err)
			continue
		}
		n.deliverWebhookRecord(ctx, record)
		if err := n.webhookRecordRepository.Save(ctx, record); err != nil {
			zap.S().Errorw("save paynda webhook record", "error", err)
		}
	}
}

func (n *PayndaWebhookNotificator) deliverWebhookRecord(ctx context.Context, record *model.WebhookRecord) {
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
	result, deliveryErr := n.webhookClient.Deliver(ctx, &PayndaWebhookDeliveryRequest{
		TargetURL:      record.TargetURL,
		Payload:        record.Payload,
		RequestHeaders: record.RequestHeaders,
		Category:       record.Event,
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
		if result != nil && result.StatusCode >= 200 && result.StatusCode < 300 {
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
}

func (n *PayndaWebhookNotificator) payndaCardStatusWebhookPayload(
	ctx context.Context,
	card *model.Card,
	status enums.CardStatus,
) ([]byte, error) {
	holder, err := n.cardHolderRepository.FindByID(ctx, card.CardHolderID)
	if err != nil {
		zap.S().Errorw("find paynda card status webhook card holder", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	account, err := n.accountRepository.FindByID(ctx, card.AccountID)
	if err != nil {
		zap.S().Errorw("find paynda card status webhook account", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	cardID := strconv.FormatInt(card.ID, 10)
	payload := payndaCardStatusEventPayload{
		CardStatusWebhook: payndaCardStatusWebhook{
			ID:               cardID,
			CreateTime:       card.CreatedAt.UTC().Format(time.RFC3339),
			UpdateTime:       card.UpdatedAt.UTC().Format(time.RFC3339),
			MerchantID:       account.ID,
			BalanceAccountID: account.ID,
			CardholderID:     holder.ID,
			CardID:           card.ID,
			MaskCardNo:       card.MaskedNumber(),
			Status:           string(paynda.ConvertGenericCardStatusToCardStatus(status)),
		},
	}
	return json.Marshal(payload)
}

func (n *PayndaWebhookNotificator) payndaTransactionWebhookPayload(
	ctx context.Context,
	transaction *model.CardTransaction,
) ([]byte, error) {
	card, err := n.cardRepository.FindByID(ctx, &CardFindByIDRequest{
		AccountID: &transaction.AccountID,
		ID:        transaction.CardID,
	})
	if err != nil {
		zap.S().Errorw("find paynda webhook card", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	holder, err := n.cardHolderRepository.FindByID(ctx, card.CardHolderID)
	if err != nil {
		zap.S().Errorw("find paynda webhook card holder", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	if card.WalletID == 0 {
		return nil, payndaerrors.ErrResourceNotFound
	}
	wallet, err := n.walletRepository.FindByID(ctx, &WalletFindByIDRequest{
		AccountID: &card.AccountID,
		ID:        card.WalletID,
	})
	if err != nil {
		zap.S().Errorw("find paynda webhook wallet", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	account, err := n.accountRepository.FindByID(ctx, card.AccountID)
	if err != nil {
		zap.S().Errorw("find paynda webhook account", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	authorizationTime := ""
	if transaction.AuthorizationID != 0 {
		authorization, err := n.authorizationRepository.FindByID(ctx, &PayndaFindAuthorizationRequest{
			ID: transaction.AuthorizationID,
		})
		if err != nil {
			zap.S().Errorw("find paynda webhook transaction authorization", "error", err)
			return nil, payndaerrors.ErrDatabaseOperation
		}
		authorizationTime = authorization.CreatedAt.UTC().Format(time.RFC3339)
	}
	transactionID := strconv.FormatInt(transaction.ID, 10)
	payload := payndaTransactionEventPayload{
		CardTransactionWebhook: payndaCardTransactionWebhook{
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
		},
	}
	if transaction.AuthorizationID != 0 {
		payload.CardTransactionWebhook.SupplierTransactionLinkID = strconv.FormatInt(transaction.AuthorizationID, 10)
	}
	return json.Marshal(payload)
}
