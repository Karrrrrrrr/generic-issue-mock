package payndapay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	CardTransferStatusInit    = "init"
	CardTransferStatusPending = "pending"
	CardTransferStatusSuccess = "success"
	CardTransferStatusFailed  = "failed"
)

type PayndaPaySDKInterface interface {
	CreateCardHolder(ctx context.Context, req *CardHolderRequest) (*Cardholder, error)
	CreateCard(ctx context.Context, req *CardCreateRequest) (*CardDetail, error)
	GetCardByID(ctx context.Context, cardID string) (*Card, error)
	GetCardSensitive(ctx context.Context, cardID string) (*CardSensitive, error)
	CardTransferIn(ctx context.Context, req *CardTransferRequest) (*CardTransferResult, error)
	CardTransferOut(ctx context.Context, req *CardTransferRequest) (*CardTransferResult, error)
	QueryCardTransfer(ctx context.Context, requestID string) (*QueryCardTransferResult, error)
	QueryTransfer(ctx context.Context, requestID string) (*CardBalanceTransferReuslt, error)
	QueryCreateCardResult(ctx context.Context, requestID string) (*CardCreateResult, error)
	CardFrozen(ctx context.Context, req *CardStatusUpdateRequest) error
	CardUnfrozen(ctx context.Context, req *CardStatusUpdateRequest) error
	QueryCardStatusUpdate(ctx context.Context, requestID string) (*CardStatusUpdateResult, error)
	ValidCardBin(ctx context.Context, cardBin string) bool
}

type CardTransferRequest struct {
	CardID    string
	CardBin   string
	Amount    decimal.Decimal
	RequestID string
}

type CardTransferResult struct {
	NeedRetry bool
	Status    string
}

type QueryCardTransferResult struct {
	Status            string
	Message           string
	NeedRetry         bool
	TransactionTime   time.Time
	Amount            decimal.Decimal
	CardBalance       decimal.Decimal
	CardBeforeBalance decimal.Decimal
}

func (p *PayndaPaySDK) ValidCardBin(_ context.Context, cardBin string) bool {
	if p.conf == nil {
		return true
	}
	for _, disabledBin := range strings.Split(p.conf.GetDisabledBins(), ",") {
		if disabledBin == cardBin {
			return false
		}
	}
	return true
}

func (p *PayndaPaySDK) GetCardByID(ctx context.Context, cardID string) (*Card, error) {
	return p.GetCard(ctx, cardID)
}

func (p *PayndaPaySDK) QueryCreateCardResult(ctx context.Context, requestID string) (*CardCreateResult, error) {
	if requestID == "" {
		return nil, ErrEmptyRequestID
	}

	r, err := p.QueryRequestResult(ctx, requestID)
	if err != nil {
		return nil, err
	}

	var rsp *APIResponse
	err = json.Unmarshal([]byte(r.Result), &rsp)
	if err != nil {
		return nil, err
	}

	result := &CardCreateResult{
		Code:       rsp.Code,
		Message:    rsp.Message,
		Success:    rsp.Success,
		CreateTime: r.CreateTime,
	}

	if rsp.Success {
		detail := CardDetail{}
		err = p.processResponseV0(rsp, &detail)
		if err != nil {
			return nil, err
		}
		result.Result = &detail
	}

	return result, nil
}

func (pp *PayndaPaySDK) CardFrozen(ctx context.Context, req *CardStatusUpdateRequest) error {
	// requestId 通过 query 透传到 Paynda，作为冻结请求的唯一幂等关联键。
	if err := req.Validate(); err != nil {
		return err
	}

	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s/status/frozen", balanceAccountID, req.CardID)
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method:    http.MethodPatch,
		Path:      apipath,
		Nonce:     genNonce(),
		RequestID: req.RequestID,
		Params:    struct{}{},
	})
	return err
}

func (pp *PayndaPaySDK) CardUnfrozen(ctx context.Context, req *CardStatusUpdateRequest) error {
	// 解冻必须沿用同一 requestId，供调用方查询 requestResults 并处理 MQ 重试。
	if err := req.Validate(); err != nil {
		return err
	}

	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s/status/unfrozen", balanceAccountID, req.CardID)
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method:    http.MethodPatch,
		Path:      apipath,
		Nonce:     genNonce(),
		RequestID: req.RequestID,
		Params:    struct{}{},
	})
	return err
}

func (pp *PayndaPaySDK) QueryCardStatusUpdate(ctx context.Context, requestID string) (*CardStatusUpdateResult, error) {
	// Paynda 将具体处理结果编码在 requestResults.data.result 的 JSON 字符串中。
	if requestID == "" {
		return nil, ErrEmptyRequestID
	}

	requestResult, err := pp.QueryRequestResult(ctx, requestID)
	if err != nil {
		return nil, err
	}

	var response APIResponse
	if err := json.Unmarshal([]byte(requestResult.Result), &response); err != nil {
		return nil, err
	}

	return &CardStatusUpdateResult{
		Code:    response.Code,
		Message: response.Message,
		Success: response.Success,
		Record: &CardStatusUpdateRecord{
			ID:         requestResult.ID,
			RequestID:  requestResult.RequestID,
			CreateTime: requestResult.CreateTime,
			UpdateTime: requestResult.UpdateTime,
		},
	}, nil
}

func (pp *PayndaPaySDK) CardTransferIn(ctx context.Context, req *CardTransferRequest) (*CardTransferResult, error) {
	if req == nil || req.CardID == "" || req.RequestID == "" {
		return &CardTransferResult{Status: CardTransferStatusFailed}, fmt.Errorf("paynda card transfer cardID and requestID are required")
	}
	if req.CardBin != "" && !pp.ValidCardBin(ctx, req.CardBin) {
		return &CardTransferResult{Status: CardTransferStatusFailed}, fmt.Errorf("paynda card BIN does not support transfers")
	}
	_, err := pp.CardTransfer(ctx, &CardBalanceTransferRequest{
		CardID:    req.CardID,
		Amount:    req.Amount.String(),
		Nonce:     req.RequestID,
		RequestID: req.RequestID,
		Type:      TransferType_IN,
	})
	return &CardTransferResult{NeedRetry: isNeedRetryCardTransferError(err)}, err
}

func (pp *PayndaPaySDK) CardTransferOut(ctx context.Context, req *CardTransferRequest) (*CardTransferResult, error) {
	if req == nil || req.CardID == "" || req.RequestID == "" {
		return &CardTransferResult{Status: CardTransferStatusFailed}, fmt.Errorf("paynda card transfer cardID and requestID are required")
	}
	if req.CardBin != "" && !pp.ValidCardBin(ctx, req.CardBin) {
		return &CardTransferResult{Status: CardTransferStatusFailed}, fmt.Errorf("paynda card BIN does not support transfers")
	}
	cardBalance, err := pp.GetCardBalance(ctx, req.CardID)
	if err != nil {
		return nil, err
	}
	amount, err := decimal.NewFromString(cardBalance.Amount)
	if err != nil {
		return &CardTransferResult{Status: CardTransferStatusFailed}, err
	}
	availableAmount, err := decimal.NewFromString(cardBalance.AvailableAmount)
	if err != nil {
		return &CardTransferResult{Status: CardTransferStatusFailed}, err
	}
	availableOut := decimal.Min(amount, availableAmount)
	if req.Amount.GreaterThan(availableOut) {
		return &CardTransferResult{Status: CardTransferStatusFailed}, fmt.Errorf("transfer amount exceeds available balance: %s", availableOut)
	}
	_, err = pp.CardTransfer(ctx, &CardBalanceTransferRequest{
		CardID:    req.CardID,
		Amount:    req.Amount.String(),
		Nonce:     req.RequestID,
		RequestID: req.RequestID,
		Type:      TransferType_OUT,
	})
	return &CardTransferResult{NeedRetry: isNeedRetryCardTransferError(err)}, err
}

func (pp *PayndaPaySDK) QueryCardTransfer(ctx context.Context, requestID string) (*QueryCardTransferResult, error) {
	if requestID == "" {
		return nil, fmt.Errorf("paynda card transfer query requestID is required")
	}
	response, err := pp.QueryTransfer(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return &QueryCardTransferResult{Status: CardTransferStatusPending}, nil
	}
	if !response.Success {
		if response.Message == "transfer record not found" {
			return &QueryCardTransferResult{Status: CardTransferStatusInit}, nil
		}
		transactionTime := time.Time{}
		if response.CreateTime != "" {
			transactionTime, err = time.Parse(time.DateTime, response.CreateTime)
			if err != nil {
				return nil, err
			}
		}
		return &QueryCardTransferResult{
			Status:          CardTransferStatusFailed,
			Message:         response.Message,
			NeedRetry:       isNeedRetryCardTransferMessage(response.Message),
			TransactionTime: transactionTime.Local(),
		}, nil
	}
	if response.TransferRecord == nil {
		return &QueryCardTransferResult{Status: CardTransferStatusPending}, nil
	}

	record := response.TransferRecord
	amount, err := decimal.NewFromString(record.Amount)
	if err != nil {
		return nil, err
	}
	cardBalance, err := decimal.NewFromString(record.NewAmount)
	if err != nil {
		return nil, err
	}
	cardBeforeBalance, err := decimal.NewFromString(record.OldAmount)
	if err != nil {
		return nil, err
	}
	transactionTime, err := time.Parse(time.DateTime, record.CreateTime)
	if err != nil {
		return nil, err
	}

	return &QueryCardTransferResult{
		Status:            CardTransferStatusSuccess,
		TransactionTime:   transactionTime.Local(),
		Amount:            amount,
		CardBalance:       cardBalance,
		CardBeforeBalance: cardBeforeBalance,
	}, nil
}

func isNeedRetryCardTransferError(err error) bool {
	return err != nil && isNeedRetryCardTransferMessage(err.Error())
}

func isNeedRetryCardTransferMessage(message string) bool {
	return strings.Contains(message, "INTERNAL_SERVER_ERROR") || strings.Contains(message, "Data conflict, please try again")
}

// IsRequestIDNotFound 判断 requestResults 查询是否因未使用过 requestId 而返回 not found。
// 该错误只表示不存在历史请求记录，调用方可使用相同 requestId 首次发起冻结或解冻。
func IsRequestIDNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
