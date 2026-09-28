package service

import (
	"context"

	"generic-mock/channel/slash/biz"
	slash "generic-mock/channel/slash/enums"
	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/channel/slash/pkg/idconv"
	common "generic-mock/enums"
	sharedbiz "generic-mock/shared/biz"
)

var _ sharedbiz.CardTransactionNotificator = (*SlashUIService)(nil)

func (s *SlashUIService) NotifyCardTransaction(ctx context.Context, req *sharedbiz.NotifyCardTransactionReq) error {
	if req == nil || req.Channel != common.Channel_Slash || req.AccountID <= 0 || req.CardTransactionID <= 0 {
		return slasherrors.ErrInvalidOperation
	}
	transactionID := idconv.ToUUID(req.CardTransactionID)
	s.webhookUsecase.Dispatch(ctx, &biz.DispatchWebhookRequest{
		AccountID: req.AccountID,
		Event:     slash.WebhookEventTransactionCreate,
		EntityID:  transactionID,
		EventID:   transactionID,
	})
	return nil
}
