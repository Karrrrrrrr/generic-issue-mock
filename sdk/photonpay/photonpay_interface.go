package photonpay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"tman/enums"

	kratosErr "github.com/go-kratos/kratos/v2/errors"
)

type PhotonpayCreateCardRequest struct {
	RequestID          string // 幂等id
	CardBin            string // 卡bin
	CardCurrency       string // 卡币种
	CardExpirationDate *int   // 卡有效期(月)
	CardScheme         string // 卡组织 MasterCard/Discover
	CardType           string // 卡类型 share/recharge
	CardFormFactor     string // 卡介质 virtual_card/physical_card
	CardholderID       string // 持卡人ID
}

// PhotonpayGetRequestResultRequest 查询请求结果参数
type PhotonpayGetRequestResultRequest struct {
	RequestID string // 请求流水号
	Type      string // 请求类型 Enum: "apply_card" "card_update" "card_freeze"
}

// PhotonpayGetRequestResultResponse 查询请求结果响应
type PhotonpayGetRequestResultResponse struct {
	CardDetail *CardDetail
	Status     enums.PhotonPayRequestResultStatus // 请求结果状态
}

type photonpaySDKInterface struct {
	sdk *PhotonPaySDK
}

type PhotonpaySDKInterface interface {
	// GetAccessToken 获取access token
	GetAccessToken(ctx context.Context) (*AccessTokenResponse, error)
	// CreateCardPre 创建卡片参数处理
	CreateCardPre(params *PhotonpayCreateCardRequest) (json.RawMessage, error)
	// CreateCard 创建卡
	CreateCard(ctx context.Context, token string, params *PhotonpayCreateCardRequest) (*OpenCardResponse, error)
	// GetCardInfo 获取卡信息
	GetCardInfo(ctx context.Context, token string, cardID string) (*GetCardDetailResponse, error)
	// GetRequestResult 查询请求结果（包含状态映射）
	GetRequestResult(ctx context.Context, token string, req *PhotonpayGetRequestResultRequest) (*PhotonpayGetRequestResultResponse, error)
	// UpdateCardStatus 冻结/解冻
	UpdateCardStatus(ctx context.Context, token string, param *FreezeCardRequest) error
	// CancelCard 删除卡
	CancelCard(ctx context.Context, token string, param *CancelCardRequest) error
	// GetCvv 获取卡CVV
	GetCvv(ctx context.Context, token string, cardID string) (*GetCvvResponse, error)
}

func NewPhotonpaySDK(sdk *PhotonPaySDK) PhotonpaySDKInterface {
	return &photonpaySDKInterface{
		sdk: sdk,
	}
}
func (ge *photonpaySDKInterface) CreateCardPre(params *PhotonpayCreateCardRequest) (json.RawMessage, error) {
	cardPostParams, err := ge.cardPostParams(params)
	if err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(cardPostParams)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(marshal), nil
}
func (ge *photonpaySDKInterface) cardPostParams(params *PhotonpayCreateCardRequest) (*OpenCardRequest, error) {
	return &OpenCardRequest{
		CardBin:            params.CardBin,            // 您需要填入你想开卡的卡bin信息，目前支持的卡bin信息可在卡bin接口中查询
		CardCurrency:       params.CardCurrency,       // 卡本币
		CardExpirationDate: params.CardExpirationDate, // 卡有效期。您可填写此卡的有效期，以月为计数单位。如：您需要虚拟卡的有效期为1年，则输入“12”即可。如不上传，则由系统自动分配虚拟卡有效期。（最小12个月，最大35个月）
		CardScheme:         params.CardScheme,         // 卡组织 Enum: "MasterCard" "Discover"
		CardType:           params.CardType,           // 卡类型。share： 共享卡； recharge： 常规卡
		CardFormFactor:     params.CardFormFactor,     // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。不填默认虚拟卡
		CardholderID:       params.CardholderID,       // 填入用卡人id后，卡将属于此用卡人。如不填则默认使用默认持卡人信息进行开卡。
		RequestID:          params.RequestID,          // 幂等id,必填 商户请求流水号，每笔交易的唯一请求号，不可重复
	}, nil
}
func (p *photonpaySDKInterface) CreateCard(ctx context.Context, token string, params *PhotonpayCreateCardRequest) (*OpenCardResponse, error) {
	req := &OpenCardRequest{
		CardBin:            params.CardBin,
		CardCurrency:       params.CardCurrency,
		CardExpirationDate: params.CardExpirationDate,
		CardScheme:         params.CardScheme,
		CardType:           params.CardType,
		CardFormFactor:     params.CardFormFactor,
		CardholderID:       params.CardholderID,
		RequestID:          params.RequestID,
	}
	resp, err := p.sdk.OpenCard(ctx, token, req)
	if err != nil {
		return nil, fmt.Errorf("photonpay open card failed: %w", err)
	}
	return resp, nil
}

func (p *photonpaySDKInterface) GetAccessToken(ctx context.Context) (*AccessTokenResponse, error) {
	resp, err := p.sdk.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("photonpay get access token failed: %w", err)
	}
	return resp, nil
}

func (p *photonpaySDKInterface) UpdateCardStatus(ctx context.Context, token string, param *FreezeCardRequest) error {
	if err := p.sdk.FreezeCard(ctx, token, param); err != nil {
		return fmt.Errorf("photonpay freeze card failed: %w", err)
	}
	return nil
}

func (p *photonpaySDKInterface) CancelCard(ctx context.Context, token string, param *CancelCardRequest) error {
	if err := p.sdk.CancelCard(ctx, token, param); err != nil {
		return fmt.Errorf("photonpay cancel card failed: %w", err)
	}
	return nil
}

func (p *photonpaySDKInterface) GetCardInfo(ctx context.Context, token string, cardID string) (*GetCardDetailResponse, error) {
	resp, err := p.sdk.GetCardDetail(ctx, token, &GetCardDetailRequest{
		CardID: cardID,
	})
	if err != nil {
		return nil, fmt.Errorf("photonpay get card detail failed: %w", err)
	}
	return resp, nil
}

// GetRequestResult 查询请求结果
func (p *photonpaySDKInterface) GetRequestResult(ctx context.Context, token string, req *PhotonpayGetRequestResultRequest) (*PhotonpayGetRequestResultResponse, error) {
	resp, err := p.sdk.GetRequestResult(ctx, token, &GetRequestResultRequest{
		RequestID: req.RequestID,
		Type:      req.Type,
	})
	if err != nil {
		return nil, fmt.Errorf("photonpay get request result failed: %w", err)
	}

	if resp == nil {
		return &PhotonpayGetRequestResultResponse{
			Status: enums.PHOTON_PAY_REQUEST_RESULT_STATUS_PENDING,
		}, nil
	}

	return &PhotonpayGetRequestResultResponse{
		Status:     enums.PhotonPayRequestResultStatusFromString(resp.Status),
		CardDetail: &resp.CardDetail,
	}, nil
}

// IsVCC1039 判断是否为VCC1039错误（无效requestId，说明未调用过）

func (p *photonpaySDKInterface) GetCvv(ctx context.Context, token string, cardID string) (*GetCvvResponse, error) {
	return p.sdk.GetCvv(ctx, token, &GetCvvRequest{
		CardID: cardID,
	})
}

// IsVCC1039 判断是否为VCC1039错误（无效requestId，说明未调用过）
func IsVCC1039(err error) bool {
	var ke *kratosErr.Error
	if errors.As(err, &ke) {
		return ke.Reason == "VCC1039"
	}
	return false
}

// IsInvalidToken 判断是否为token失效错误（1002）
func IsInvalidToken(err error) bool {
	var ke *kratosErr.Error
	if errors.As(err, &ke) {
		return ke.Reason == "1002"
	}
	return false
}
