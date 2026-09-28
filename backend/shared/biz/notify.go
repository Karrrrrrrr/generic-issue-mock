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
}

type NoopNotificator struct{}

func (NoopNotificator) NotifyIssueCard(context.Context, *NotifyIssueCardReq) error {
	return nil
}
