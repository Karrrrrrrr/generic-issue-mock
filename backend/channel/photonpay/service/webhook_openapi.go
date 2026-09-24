package service

import (
	"context"

	photon "generic-mock/channel/photonpay/enums"
)

type OpenAPIWebhookNotificationRequest struct {
	OpenAPIAccountRequest
	TopicCode    *string `json:"topicCode" form:"topicCode"`       // Invalid: SDK subscription mutations do not change UI webhook configuration.
	TemplateCode *string `json:"templateCode" form:"templateCode"` // Invalid: SDK subscription mutations do not change UI webhook configuration.
}

type OpenAPIWebhookTemplate struct {
	Version       string `json:"version"`
	TemplateCode  string `json:"templateCode"`
	ContentSample string `json:"contentSample"`
	Active        bool   `json:"active"`
}

type OpenAPIWebhookTopic struct {
	TopicCode string                   `json:"topicCode"`
	Name      string                   `json:"name"`
	Templates []OpenAPIWebhookTemplate `json:"templates"`
}

type OpenAPIWebhookCategory struct {
	Category photon.WebhookNotificationCategory `json:"category"`
	Name     string                             `json:"name"`
	Topics   []OpenAPIWebhookTopic              `json:"topics"`
}

type OpenAPIWebhookNotifications struct {
	Categories []OpenAPIWebhookCategory `json:"categories"`
}

func (service *PhotonPayOpenAPIService) WebhookNotifications(ctx context.Context, req *OpenAPIWebhookNotificationRequest) (*OpenAPIWebhookNotifications, error) {
	if _, err := service.accountID(&req.OpenAPIAccountRequest); err != nil {
		return nil, err
	}
	return &OpenAPIWebhookNotifications{
		Categories: []OpenAPIWebhookCategory{{
			Category: photon.WebhookNotificationCategoryIssuing,
			Name:     "Card transactions",
			Topics: []OpenAPIWebhookTopic{{
				TopicCode: "vcc",
				Name:      "Card transactions",
				Templates: []OpenAPIWebhookTemplate{{
					Version:       "1",
					TemplateCode:  "transaction",
					ContentSample: "{}",
					Active:        false,
				}},
			}},
		}},
	}, nil
}
