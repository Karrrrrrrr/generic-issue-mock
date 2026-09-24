package service

import (
	"context"
	"encoding/json"
	"time"

	slash "generic-mock/channel/slash/enums"
)

type OpenAPIWebhookRequest struct {
	OpenAPIAccountPathRequest
	Config     *json.RawMessage                  `json:"config"`     // Invalid: authorization fallback is not persisted.
	Name       *string                           `json:"name"`       // Invalid: SDK webhook configuration is not persisted by this protocol-only endpoint.
	URL        *string                           `json:"url"`        // Invalid: SDK webhook configuration is not persisted by this protocol-only endpoint.
	WebhookURL *string                           `json:"webhookUrl"` // Invalid: UI manages operational authorization callbacks.
	Status     *slash.AuthorizationWebhookStatus `json:"status"`     // Invalid: UI manages operational authorization callbacks.
}

type OpenAPIWebhook struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createAt"`
	UpdatedAt time.Time `json:"updateAt"`
}

func (service *SlashOpenAPIService) Webhook(ctx context.Context, req *OpenAPIWebhookRequest) (*OpenAPIWebhook, error) {
	item, err := service.protocolAccount(ctx, &req.OpenAPIAccountPathRequest)
	if err != nil {
		return nil, err
	}
	result := &OpenAPIWebhook{
		ID:        slashIDString(item.ID),
		Name:      item.Name,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
	if req.Name != nil {
		result.Name = *req.Name
	}
	if req.URL != nil {
		result.URL = *req.URL
	}
	return result, nil
}

func (service *SlashOpenAPIService) ListOpenAPIWebhooks(ctx context.Context, req *OpenAPIWebhookRequest) (*OpenAPIItems[*OpenAPIWebhook], error) {
	item, err := service.Webhook(ctx, req)
	if err != nil {
		return nil, err
	}
	return &OpenAPIItems[*OpenAPIWebhook]{
		Items:    []*OpenAPIWebhook{item},
		Metadata: OpenAPIMetadata{Count: 1},
	}, nil
}

type OpenAPIAuthWebhook struct {
	WebhookURL        string                           `json:"webhookUrl"`
	SigningSecret     string                           `json:"signingSecret"` // Invalid: signatures are not validated.
	Status            slash.AuthorizationWebhookStatus `json:"status"`
	TimeoutDurationMs int                              `json:"timeoutDurationMs"`
	Config            struct{}                         `json:"config"` // Invalid: authorization fallback is unsupported.
	CreatedAt         time.Time                        `json:"createdAt"`
	UpdatedAt         time.Time                        `json:"updatedAt"`
}

func (service *SlashOpenAPIService) AuthWebhook(ctx context.Context, req *OpenAPIWebhookRequest) (*OpenAPIAuthWebhook, error) {
	item, err := service.protocolAccount(ctx, &req.OpenAPIAccountPathRequest)
	if err != nil {
		return nil, err
	}
	result := &OpenAPIAuthWebhook{
		Status:            slash.AuthorizationWebhookDisabled,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		TimeoutDurationMs: 1000,
	}
	if req.WebhookURL != nil {
		result.WebhookURL = *req.WebhookURL
	}
	if req.Status != nil {
		result.Status = *req.Status
	}
	return result, nil
}
