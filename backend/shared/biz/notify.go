package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type NotifyIssueCardReq struct {
	AccountID model.ID
	Channel   enums.Channel
	CardID    model.ID
}

// Notificator shared暴露接口 在各渠道实现
type Notificator interface {
	NotifyIssueCard(context.Context, *NotifyIssueCardReq) error
	NotifyCardTransaction(context.Context, *NotifyCardTransactionReq) error
	NotifyCardFunding(context.Context, *NotifyCardFundingReq) error
}

type NoopNotificator struct{}

var _ Notificator = NoopNotificator{}

func (NoopNotificator) NotifyIssueCard(context.Context, *NotifyIssueCardReq) error {
	return nil
}

type NotifyCardTransactionReq struct {
	AccountID         model.ID
	Channel           enums.Channel
	CardID            model.ID
	AuthorizationID   model.ID
	CardTransactionID model.ID
	Type              enums.CardTransactionType
}

func (NoopNotificator) NotifyCardTransaction(context.Context, *NotifyCardTransactionReq) error {
	return nil
}

type NotifyCardFundingReq struct {
	AccountID        model.ID
	Channel          enums.Channel
	WalletTransferID model.ID
}

func (NoopNotificator) NotifyCardFunding(context.Context, *NotifyCardFundingReq) error {
	return nil
}
