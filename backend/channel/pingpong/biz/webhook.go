package biz

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"time"

	ping "generic-mock/channel/pingpong/enums"
	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

type PingPongWebhookUsecase struct {
	accountRepo         sharedbiz.AccountRepo
	cardRepo            sharedbiz.CardRepo
	cardTransactionRepo sharedbiz.CardTransactionRepo
	walletTransferRepo  sharedbiz.WalletTransferRepo
	webhookConfigRepo   sharedbiz.WebhookConfigRepo
	webhookRecordRepo   sharedbiz.WebhookRecordRepo
	tx                  sharedbiz.Transaction
	client              WebhookClient
}

func NewWebhookUsecase(injector do.Injector) (*PingPongWebhookUsecase, error) {
	return &PingPongWebhookUsecase{
		accountRepo:         do.MustInvoke[sharedbiz.AccountRepo](injector),
		cardRepo:            do.MustInvoke[sharedbiz.CardRepo](injector),
		cardTransactionRepo: do.MustInvoke[sharedbiz.CardTransactionRepo](injector),
		walletTransferRepo:  do.MustInvoke[sharedbiz.WalletTransferRepo](injector),
		webhookConfigRepo:   do.MustInvoke[sharedbiz.WebhookConfigRepo](injector),
		webhookRecordRepo:   do.MustInvoke[sharedbiz.WebhookRecordRepo](injector),
		tx:                  do.MustInvoke[sharedbiz.Transaction](injector),
		client:              do.MustInvoke[WebhookClient](injector),
	}, nil
}

type LoadWebhookSourceRequest struct {
	AccountID model.ID
	SourceID  model.ID
	Event     ping.WebhookEvent
}

func (req *LoadWebhookSourceRequest) Validate() error {
	if req == nil || req.AccountID <= 0 || req.SourceID <= 0 || !req.Event.Valid() {
		return pingerrors.ErrInvalid
	}
	return nil
}

type WebhookSource struct {
	Card        *model.Card
	Transaction *model.CardTransaction
	Transfer    *model.WalletTransfer
}

func (uc *PingPongWebhookUsecase) LoadSource(ctx context.Context, req *LoadWebhookSourceRequest) (*WebhookSource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	result := &WebhookSource{}
	cardID := req.SourceID
	switch req.Event {
	case ping.WebhookAuthorization, ping.WebhookClearing:
		exists, err := uc.cardTransactionRepo.Exist(ctx, &sharedbiz.CardTransactionExistRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: req.SourceID})
		if err != nil {
			zap.S().Errorw("check pingpong webhook transaction", "error", err)
			return nil, pingerrors.ErrDatabase
		}
		if !exists {
			return nil, pingerrors.ErrNotFound
		}
		result.Transaction, err = uc.cardTransactionRepo.Find(ctx, &sharedbiz.CardTransactionFindRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: req.SourceID})
		if err != nil {
			zap.S().Errorw("find pingpong webhook transaction", "error", err)
			return nil, pingerrors.ErrDatabase
		}
		item := result.Transaction
		if req.Event == ping.WebhookAuthorization && item.Type != common.CardTransactionType_AUTH && item.Type != common.CardTransactionType_VOID {
			return nil, pingerrors.ErrInvalid
		}
		if req.Event == ping.WebhookClearing && item.Type != common.CardTransactionType_CLEAR && item.Type != common.CardTransactionType_REFUND {
			return nil, pingerrors.ErrInvalid
		}
		if item.AuthorizationID != 0 {
			auth := item.Authorization
			if auth == nil || auth.AccountID != req.AccountID || auth.Channel != common.Channel_PingPong || auth.CardID != item.CardID {
				return nil, pingerrors.ErrInvalid
			}
		} else if item.Type != common.CardTransactionType_REFUND {
			return nil, pingerrors.ErrInvalid
		}
		cardID = item.CardID
	case ping.WebhookCardOperate, ping.WebhookTransfer:
		exists, err := uc.walletTransferRepo.Exist(ctx, &sharedbiz.WalletTransferExistRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: req.SourceID})
		if err != nil {
			zap.S().Errorw("check pingpong webhook transfer", "error", err)
			return nil, pingerrors.ErrDatabase
		}
		if !exists {
			return nil, pingerrors.ErrNotFound
		}
		result.Transfer, err = uc.walletTransferRepo.Find(ctx, &sharedbiz.WalletTransferFindRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: req.SourceID})
		if err != nil {
			zap.S().Errorw("find pingpong webhook transfer", "error", err)
			return nil, pingerrors.ErrDatabase
		}
		if result.Transfer.CardID == nil {
			return nil, pingerrors.ErrInvalid
		}
		if req.Event == ping.WebhookCardOperate && result.Transfer.Kind != common.WalletTransfer_CardTopUp {
			return nil, pingerrors.ErrInvalid
		}
		if req.Event == ping.WebhookTransfer && result.Transfer.Kind != common.WalletTransfer_CardWithdraw {
			return nil, pingerrors.ErrInvalid
		}
		cardID = *result.Transfer.CardID
	}
	exists, err := uc.cardRepo.Exist(ctx, &sharedbiz.CardExistRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: cardID})
	if err != nil {
		zap.S().Errorw("check pingpong webhook card", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	result.Card, err = uc.cardRepo.Find(ctx, &sharedbiz.CardFindRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: cardID})
	if err != nil {
		zap.S().Errorw("find pingpong webhook card", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if result.Transfer != nil {
		if result.Transfer.Currency != result.Card.CardCurrency || result.Card.VirtualAccountID == nil || result.Card.VirtualAccount == nil {
			return nil, pingerrors.ErrInvalid
		}
		virtual := result.Card.VirtualAccount
		if virtual.AccountID != req.AccountID || virtual.Channel != common.Channel_PingPong {
			return nil, pingerrors.ErrInvalid
		}
		if req.Event == ping.WebhookCardOperate && (result.Transfer.SourceWalletID != virtual.WalletID || result.Transfer.TargetWalletID != result.Card.WalletID) {
			return nil, pingerrors.ErrInvalid
		}
		if req.Event == ping.WebhookTransfer && (result.Transfer.SourceWalletID != result.Card.WalletID || result.Transfer.TargetWalletID != virtual.WalletID) {
			return nil, pingerrors.ErrInvalid
		}
	}
	return result, nil
}

type QueueWebhookRequest struct {
	AccountID model.ID
	Event     ping.WebhookEvent
	SourceID  string
	Payload   []byte
}

func (uc *PingPongWebhookUsecase) Queue(ctx context.Context, req *QueueWebhookRequest) error {
	if req == nil || req.AccountID <= 0 || !req.Event.Valid() || req.SourceID == "" || !json.Valid(req.Payload) {
		return pingerrors.ErrInvalid
	}
	if uc.tx.IsInTx(ctx) {
		return pingerrors.ErrWebhookInTransaction
	}
	var pending []*model.WebhookRecord
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		exists, err := uc.accountRepo.Exist(ctx, &sharedbiz.AccountExistRequest{ID: req.AccountID, Channel: common.Channel_PingPong})
		if err != nil {
			zap.S().Errorw("check pingpong webhook owner", "error", err)
			return pingerrors.ErrDatabase
		}
		if !exists {
			return pingerrors.ErrNotFound
		}
		if _, err := uc.accountRepo.FindByIDWithLock(ctx, &sharedbiz.AccountFindByIDWithLockRequest{ID: req.AccountID, Channel: common.Channel_PingPong}); err != nil {
			zap.S().Errorw("lock pingpong webhook owner", "error", err)
			return pingerrors.ErrDatabase
		}
		for offset := 0; ; offset += 100 {
			configs, err := uc.webhookConfigRepo.List(ctx, &sharedbiz.WebhookConfigListRequest{
				WebhookConfigFilters: sharedbiz.WebhookConfigFilters{AccountIDs: []model.ID{req.AccountID}, Channel: common.Channel_PingPong, Events: []string{string(req.Event)}},
				Offset:               offset, Limit: 100,
			})
			if err != nil {
				zap.S().Errorw("list pingpong webhook subscriptions", "error", err)
				return pingerrors.ErrDatabase
			}
			for _, config := range configs {
				if !config.Enabled {
					continue
				}
				exists, err := uc.webhookRecordRepo.ExistBySource(ctx, &sharedbiz.WebhookRecordExistBySourceRequest{
					AccountID: req.AccountID, Channel: common.Channel_PingPong, WebhookConfigID: config.ID, Event: string(req.Event), SourceID: req.SourceID,
				})
				if err != nil {
					zap.S().Errorw("check pingpong webhook duplicate", "error", err)
					return pingerrors.ErrDatabase
				}
				if exists {
					continue
				}
				record := &model.WebhookRecord{
					AccountID: req.AccountID, Channel: common.Channel_PingPong, WebhookConfigID: config.ID,
					Event: string(req.Event), SourceID: req.SourceID, TargetURL: config.TargetURL,
					Payload: slices.Clone(req.Payload), RequestHeaders: []byte(`{"Content-Type":["application/json"]}`),
					ResponseHeaders: []byte(`{}`), Status: common.WebhookDeliveryStatus_Pending, AttemptCount: 1,
				}
				if err := uc.webhookRecordRepo.Create(ctx, &sharedbiz.WebhookRecordCreateRequest{Record: record}); err != nil {
					zap.S().Errorw("create pingpong webhook delivery", "error", err)
					return pingerrors.ErrDatabase
				}
				pending = append(pending, record)
			}
			if len(configs) < 100 {
				break
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, record := range pending {
		go func() {
			deliveryCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = uc.deliver(deliveryCtx, record)
		}()
	}
	return nil
}

type ReplayWebhookRequest struct {
	AccountID model.ID
	ID        model.ID
}

func (uc *PingPongWebhookUsecase) ReplayRecord(ctx context.Context, req *ReplayWebhookRequest) error {
	if req == nil || req.AccountID <= 0 || req.ID <= 0 {
		return pingerrors.ErrInvalid
	}
	if uc.tx.IsInTx(ctx) {
		return pingerrors.ErrWebhookInTransaction
	}
	exists, err := uc.webhookRecordRepo.Exist(ctx, &sharedbiz.WebhookRecordExistRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: req.ID})
	if err != nil {
		zap.S().Errorw("check pingpong replay source", "error", err)
		return pingerrors.ErrDatabase
	}
	if !exists {
		return pingerrors.ErrNotFound
	}
	source, err := uc.webhookRecordRepo.Find(ctx, &sharedbiz.WebhookRecordFindRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: req.ID})
	if err != nil {
		zap.S().Errorw("find pingpong replay source", "error", err)
		return pingerrors.ErrDatabase
	}
	replay := &model.WebhookRecord{
		AccountID: source.AccountID, Channel: source.Channel, WebhookConfigID: source.WebhookConfigID,
		Event: source.Event, SourceID: source.SourceID, TargetURL: source.TargetURL,
		Payload: slices.Clone(source.Payload), RequestHeaders: slices.Clone(source.RequestHeaders),
		ResponseHeaders: []byte(`{}`), Status: common.WebhookDeliveryStatus_Pending, AttemptCount: source.AttemptCount + 1,
	}
	if err := uc.webhookRecordRepo.Create(ctx, &sharedbiz.WebhookRecordCreateRequest{Record: replay}); err != nil {
		zap.S().Errorw("create pingpong webhook replay", "error", err)
		return pingerrors.ErrDatabase
	}
	deliveryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	return uc.deliver(deliveryCtx, replay)
}

func (uc *PingPongWebhookUsecase) deliver(ctx context.Context, record *model.WebhookRecord) error {
	headers := make(http.Header)
	err := json.Unmarshal(record.RequestHeaders, &headers)
	var result *SendWebhookResult
	if err == nil {
		result, err = uc.client.Send(ctx, &SendWebhookRequest{TargetURL: record.TargetURL, Payload: record.Payload, Headers: headers})
	}
	update := &sharedbiz.WebhookRecordUpdateDeliveryRequest{
		AccountID: record.AccountID, Channel: record.Channel, ID: record.ID,
		Status: common.WebhookDeliveryStatus_Failed, ResponseHeaders: []byte(`{}`),
	}
	if result != nil {
		update.StatusCode = result.StatusCode
		update.ResponseBody = result.Body
		update.ResponseHeaders, _ = json.Marshal(result.Headers)
	}
	// TODO: 暂定 HTTP 200 即成功，待确认 ACK body 及正式成功条件。
	if err == nil && result != nil && result.StatusCode == http.StatusOK {
		deliveredAt := time.Now().UTC()
		update.Status = common.WebhookDeliveryStatus_Succeeded
		update.DeliveredAt = &deliveredAt
	} else {
		update.ErrorMessage = "expected HTTP 200"
		if err != nil {
			update.ErrorMessage = err.Error()
		}
	}
	saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := uc.webhookRecordRepo.UpdateDelivery(saveCtx, update); err != nil {
		zap.S().Errorw("save pingpong webhook delivery result", "record_id", record.ID, "error", err)
		return pingerrors.ErrDatabase
	}
	zap.S().Infow("pingpong webhook delivery", "record_id", record.ID, "account_id", record.AccountID,
		"event", record.Event, "url", record.TargetURL, "status_code", update.StatusCode, "status", update.Status, "error", update.ErrorMessage)
	if update.Status == common.WebhookDeliveryStatus_Failed {
		return pingerrors.ErrWebhookDelivery
	}
	return nil
}

type LoadWebhookFundingRequest struct {
	AccountID  model.ID
	TransferID model.ID
}

func (uc *PingPongWebhookUsecase) LoadFundingSource(ctx context.Context, req *LoadWebhookFundingRequest) (*LoadWebhookSourceRequest, error) {
	if req == nil || req.AccountID <= 0 || req.TransferID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.walletTransferRepo.Exist(ctx, &sharedbiz.WalletTransferExistRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: req.TransferID})
	if err != nil {
		zap.S().Errorw("check pingpong funding notification source", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	transfer, err := uc.walletTransferRepo.Find(ctx, &sharedbiz.WalletTransferFindRequest{AccountID: req.AccountID, Channel: common.Channel_PingPong, ID: req.TransferID})
	if err != nil {
		zap.S().Errorw("find pingpong funding notification source", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	event := ping.WebhookCardOperate
	if transfer.Kind == common.WalletTransfer_CardWithdraw {
		event = ping.WebhookTransfer
	} else if transfer.Kind != common.WalletTransfer_CardTopUp {
		return nil, pingerrors.ErrInvalid
	}
	return &LoadWebhookSourceRequest{AccountID: req.AccountID, SourceID: req.TransferID, Event: event}, nil
}
