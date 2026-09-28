package service

import (
	"generic-mock/channel/slash/biz"
	slash "generic-mock/channel/slash/enums"
	"generic-mock/channel/slash/pkg/idconv"
	"generic-mock/model"
)

type webhookDispatchRequest struct {
	AccountID  model.ID
	Event      slash.WebhookEvent
	ResourceID model.ID
}

func toWebhookDispatchRequest(req *webhookDispatchRequest) *biz.DispatchWebhookRequest {
	resourceIDString := idconv.ToUUID(req.ResourceID)
	return &biz.DispatchWebhookRequest{
		AccountID: req.AccountID,
		Event:     req.Event,
		EntityID:  resourceIDString,
		EventID:   resourceIDString,
	}
}
