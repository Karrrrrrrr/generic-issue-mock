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
	NotifyCardStatus(context.Context, *NotifyCardStatusReq) error
	NotifyCardTransaction(context.Context, *NotifyCardTransactionReq) error
	NotifyCardFunding(context.Context, *NotifyCardFundingReq) error
}

type NoopNotificator struct{}

type NotifyCardStatusReq struct {
	AccountID model.ID
	Channel   enums.Channel
	CardID    model.ID
	Status    enums.CardStatus
}

type NotifyCardTransactionReq struct {
	AccountID         model.ID
	Channel           enums.Channel
	CardID            model.ID
	AuthorizationID   model.ID
	CardTransactionID model.ID
	Type              enums.CardTransactionType
}

type NotifyCardFundingReq struct {
	AccountID        model.ID
	Channel          enums.Channel
	WalletTransferID model.ID
}
