package biz

import (
	"context"
	"encoding/json"

	ping "generic-mock/channel/pingpong/enums"
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
	sharedbiz "generic-mock/shared/biz"

	"go.uber.org/zap"
)

var _ sharedbiz.Notificator = (*PingPongWebhookUsecase)(nil)
var _ sharedbiz.UIWebhookEventCatalog = (*PingPongWebhookUsecase)(nil)
var _ sharedbiz.UIWebhookReplayer = (*PingPongWebhookUsecase)(nil)

func (uc *PingPongWebhookUsecase) NotifyCardTransaction(ctx context.Context, req *sharedbiz.NotifyCardTransactionReq) error {
	if req == nil || req.Channel != common.Channel_PingPong {
		return pingerrors.ErrInvalid
	}
	var event ping.WebhookEvent
	switch req.Type {
	case common.CardTransactionType_AUTH, common.CardTransactionType_VOID:
		event = ping.WebhookAuthorization
	case common.CardTransactionType_CLEAR, common.CardTransactionType_REFUND:
		event = ping.WebhookClearing
	default:
		return pingerrors.ErrInvalid
	}
	return uc.Dispatch(ctx, &LoadWebhookSourceRequest{AccountID: req.AccountID, SourceID: req.CardTransactionID, Event: event})
}

func (uc *PingPongWebhookUsecase) NotifyIssueCard(ctx context.Context, req *sharedbiz.NotifyIssueCardReq) error {
	if req == nil || req.Channel != common.Channel_PingPong {
		return pingerrors.ErrInvalid
	}
	return uc.Dispatch(ctx, &LoadWebhookSourceRequest{AccountID: req.AccountID, SourceID: req.CardID, Event: ping.WebhookOpenCard})
}

func (uc *PingPongWebhookUsecase) NotifyCardFunding(ctx context.Context, req *sharedbiz.NotifyCardFundingReq) error {
	if req == nil || req.Channel != common.Channel_PingPong {
		return pingerrors.ErrInvalid
	}
	source, err := uc.LoadFundingSource(ctx, &LoadWebhookFundingRequest{AccountID: req.AccountID, TransferID: req.WalletTransferID})
	if err != nil {
		return err
	}
	return uc.Dispatch(ctx, source)
}

func (uc *PingPongWebhookUsecase) ListEvents(ctx context.Context, req *sharedbiz.UIWebhookEventsRequest) ([]string, error) {
	if req == nil || req.Channel != common.Channel_PingPong {
		return nil, pingerrors.ErrInvalid
	}
	events := ping.WebhookEvents()
	result := make([]string, 0, len(events))
	for _, event := range events {
		result = append(result, string(event))
	}
	return result, nil
}

func (uc *PingPongWebhookUsecase) Replay(ctx context.Context, req *sharedbiz.UIWebhookReplayRequest) error {
	if req == nil || req.Record == nil || req.Record.Channel != common.Channel_PingPong {
		return pingerrors.ErrInvalid
	}
	return uc.ReplayRecord(ctx, &ReplayWebhookRequest{AccountID: req.Record.AccountID, ID: req.Record.ID})
}

func (uc *PingPongWebhookUsecase) Dispatch(ctx context.Context, req *LoadWebhookSourceRequest) error {
	source, err := uc.LoadSource(ctx, req)
	if err != nil {
		return err
	}
	payload, err := toWebhookPayload(&webhookPayloadRequest{Event: req.Event, Source: source})
	if err != nil {
		return err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		zap.S().Errorw("encode pingpong webhook payload", "event", req.Event, "error", err)
		return pingerrors.ErrWebhookPayload
	}
	// TODO: 确认事件所在 header/外层封装，暂直接发送文档业务 JSON，不自行添加协议字段；签名暂不实现。
	return uc.Queue(ctx, &QueueWebhookRequest{
		AccountID: req.AccountID, Event: req.Event, SourceID: idconv.ToString(req.SourceID), Payload: raw,
	})
}
