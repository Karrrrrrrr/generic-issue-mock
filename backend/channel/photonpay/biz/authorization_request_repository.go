package biz

import "context"

type AuthorizationRequestDeliveryRequest struct {
	TargetURL     string
	Payload       []byte
	TimeoutMillis int
}

type AuthorizationRequestDeliveryResult struct {
	StatusCode      int
	ResponseBody    string
	RequestHeaders  []byte
	ResponseHeaders []byte
}

type AuthorizationClient interface {
	RequestAuthorization(
		context.Context,
		*AuthorizationRequestDeliveryRequest,
	) (*AuthorizationRequestDeliveryResult, error)
}
