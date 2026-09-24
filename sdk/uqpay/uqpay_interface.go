package uqpay

import "context"

// TODO: UQPay 调用三方时，request_id 必须使用 UUID 作为 X-Idempotency-Key。
type UQPayInterface interface {
	GetAccessToken(ctx context.Context) (*GetAccessTokenResp, error)
	CreateCard(ctx context.Context, token, idempotencyKey string, req *CreateCardReq) (*CreateCardResp, error)
	GetCardInfo(ctx context.Context, token string, cardID string) (*GetCardInfo, error)
	RetrieveCardOrder(ctx context.Context, token string, orderID string) (*CardOrder, error)
	UpdateCardStatus(ctx context.Context, token, idempotencyKey string, req *UpdateCardStatusReq) (*UpdateCardStatusResp, error)
	GetDefaultCardLimitAmount(ctx context.Context) float64
}
