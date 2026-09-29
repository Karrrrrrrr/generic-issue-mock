package biz

import "context"

type SlashAuthorizationRequestDeliveryRequest struct {
	TargetURL     string
	Payload       []byte
	TimeoutMillis int
}

type SlashAuthorizationRequestDeliveryResult struct {
	StatusCode      int
	ResponseBody    string
	RequestHeaders  []byte
	ResponseHeaders []byte
}

type SlashAuthorizationClient interface {
	RequestAuthorization(
		context.Context,
		*SlashAuthorizationRequestDeliveryRequest,
	) (*SlashAuthorizationRequestDeliveryResult, error)
}
