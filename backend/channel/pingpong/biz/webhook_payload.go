package biz

import (
	"strings"

	ping "generic-mock/channel/pingpong/enums"
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
	"generic-mock/model"
	datetime "generic-mock/pkg/types/time"

	"github.com/shopspring/decimal"
)

// TODO: 时间暂按示例 yyyy-MM-dd HH:mm:ss，待确认分隔符和时区。
// TODO: 字段大小写及文档拼写差异待确认，目前保留字段表的 record_Id、card_Id、unique_oder_id。
type OpenCardWebhook struct {
	OrderID  string   `json:"order_id"`
	BudgetID *string  `json:"budget_id,omitempty"`
	CardIDs  []string `json:"card_id_list"`
}

type CardOperateWebhook struct {
	RecordID      string                  `json:"record_Id"`
	UniqueOrderID string                  `json:"unique_oder_id"`
	Created       datetime.DateTime       `json:"created"`
	ChangeAmount  webhookNumber           `json:"change_amount"`
	Currency      common.Currency         `json:"currency"`
	CardNumber    string                  `json:"card_number"`
	CardID        string                  `json:"card_Id"`
	Remark        string                  `json:"remark"`
	Status        ping.FundingStatus      `json:"status"`
	OperateType   ping.WebhookOperateType `json:"operate_type"`
	BudgetID      *string                 `json:"budget_id,omitempty"`
	OperateReason string                  `json:"operate_reason"`
}

type AuthorizationWebhook struct {
	AuthorizationDate       datetime.DateTime               `json:"authorization_date"`
	AuthorizationID         string                          `json:"authorization_id"`
	OriginalAuthorizationID *string                         `json:"original_authorization_id,omitempty"`
	CardID                  string                          `json:"card_id"`
	CardNumber              string                          `json:"card_number"`
	ApproveCode             string                          `json:"approve_code"`
	MCC                     string                          `json:"mcc"`
	MerchantName            string                          `json:"merchant_name"`
	MerchantCountry         string                          `json:"merchant_country"`
	BillingAmount           webhookNumber                   `json:"billing_amount"`
	BillingCurrency         common.Currency                 `json:"billing_currency"`
	TransactionAmount       webhookNumber                   `json:"transaction_amount"`
	TransactionCurrency     common.Currency                 `json:"transaction_currency"`
	AuthorizationType       ping.WebhookAuthorizationType   `json:"authorization_type"`
	AuthorizationStatus     ping.WebhookAuthorizationStatus `json:"authorization_status"`
	FailReason              *string                         `json:"fail_reason"` // Invalid: 没有渠道拒绝原因模型；成功时为 null。
	BudgetID                *string                         `json:"budget_id,omitempty"`
	RateFee                 webhookNumber                   `json:"rate_fee"` // Invalid: mock 暂无渠道费率模型，返回 0。
	RateFeeCurrency         common.Currency                 `json:"rate_fee_currency"`
	ThreeDSFee              webhookNumber                   `json:"threeds_fee"` // Invalid: 未实现 3DS 费用，返回 0。
	ThreeDSFeeCurrency      common.Currency                 `json:"threeds_fee_currency"`
	ReconciliationInfo      *WebhookReconciliationInfo      `json:"reconciliation_info"` // Invalid: 未实现 OTA 卡，返回 null。
}

type ClearingWebhook struct {
	ClearDate          datetime.DateTime          `json:"clear_date"`
	CardID             string                     `json:"card_id"`
	CardNumber         string                     `json:"card_number"`
	OriginalAuthID     *string                    `json:"original_auth_id,omitempty"`
	ClearID            string                     `json:"clear_id"`
	ClearType          ping.WebhookClearType      `json:"clear_type"`
	ClearStatus        ping.WebhookClearStatus    `json:"clear_status"`
	SettlementAmount   webhookNumber              `json:"settlement_amount"`
	SettlementCurrency common.Currency            `json:"settlement_currency"`
	BudgetID           *string                    `json:"budget_id,omitempty"`
	MerchantName       string                     `json:"merchant_name"`
	MerchantCountry    string                     `json:"merchant_country"`
	MCC                string                     `json:"mcc"`
	ReconciliationInfo *WebhookReconciliationInfo `json:"reconciliation_info"` // Invalid: 未实现 OTA 卡，返回 null。
}

type TransferWebhook struct {
	TransferReason   string                   `json:"transfer_reason"`
	TransferInType   ping.WebhookTransferType `json:"transfer_in_type"`
	TransferOutType  ping.WebhookTransferType `json:"transfer_out_type"`
	CardID           string                   `json:"card_id"`
	BudgetID         *string                  `json:"budget_id,omitempty"`
	TransferDate     datetime.DateTime        `json:"transfer_date"`
	TransferAmount   webhookNumber            `json:"transfer_amount"`
	TransferCurrency common.Currency          `json:"transfer_currency"`
}

type WebhookReconciliationInfo struct {
	ServiceType       *ping.WebhookServiceType `json:"service_type"`
	SupplierReference *string                  `json:"supplier_reference"`
	BookingID         *string                  `json:"booking_id"`
	Routing           *string                  `json:"routing"`
	TicketNo          *string                  `json:"ticket_no"`
	EmployeeNo        *string                  `json:"employee_no"`
	BookingClass      *string                  `json:"booking_class"`
	AirlineCode       *string                  `json:"airline_code"`
	Department        *string                  `json:"department"`
	CostCentre        *string                  `json:"cost_centre"`
	CustomField1      *string                  `json:"custom_field1"`
	CustomField2      *string                  `json:"custom_field2"`
	CustomField3      *string                  `json:"custom_field3"`
	CustomField4      *string                  `json:"custom_field4"`
	CustomField5      *string                  `json:"custom_field5"`
}

type webhookPayloadRequest struct {
	Event  ping.WebhookEvent
	Source *WebhookSource
}

func toWebhookPayload(req *webhookPayloadRequest) (any, error) {
	if req == nil || req.Source == nil || req.Source.Card == nil {
		return nil, pingerrors.ErrInvalid
	}
	card := req.Source.Card
	cardID := idconv.ToString(card.ID)
	budgetID := toWebhookOptionalID(card.VirtualAccountID)
	masked := maskWebhookCardNumber(card.CardNumber)
	switch req.Event {
	case ping.WebhookOpenCard:
		return &OpenCardWebhook{OrderID: cardID, BudgetID: budgetID, CardIDs: []string{cardID}}, nil
	case ping.WebhookCardOperate:
		transfer := req.Source.Transfer
		if transfer == nil || transfer.Kind != common.WalletTransfer_CardTopUp {
			return nil, pingerrors.ErrInvalid
		}
		return &CardOperateWebhook{
			RecordID: idconv.ToString(transfer.ID), UniqueOrderID: transfer.RequestID,
			Created: datetime.DateTime(transfer.CreatedAt.UTC()), ChangeAmount: webhookNumber{transfer.Amount.Round(2)},
			Currency: transfer.Currency, CardNumber: masked, CardID: cardID, Remark: "Card funding",
			Status: ping.FundingSuccess, OperateType: ping.WebhookCardIn, BudgetID: budgetID, OperateReason: "Card funding",
		}, nil
	case ping.WebhookTransfer:
		transfer := req.Source.Transfer
		if transfer == nil || transfer.Kind != common.WalletTransfer_CardWithdraw {
			return nil, pingerrors.ErrInvalid
		}
		// TODO: 确认 notify_transfer 是否仅用于销卡余额回退；暂按实际已提交的卡到 VA 转出生成。
		reason := "Card withdrawal"
		if card.Status == common.CardStatus_Deleted {
			reason = "Card deactivation balance transfer"
		}
		return &TransferWebhook{
			TransferReason: reason, TransferInType: ping.WebhookTransferBudget, TransferOutType: ping.WebhookTransferCard,
			CardID: cardID, BudgetID: budgetID, TransferDate: datetime.DateTime(transfer.CreatedAt.UTC()),
			TransferAmount: webhookNumber{transfer.Amount.Round(2)}, TransferCurrency: transfer.Currency,
		}, nil
	case ping.WebhookAuthorization:
		transaction := req.Source.Transaction
		if transaction == nil || transaction.AuthorizationID <= 0 {
			return nil, pingerrors.ErrInvalid
		}
		payload := &AuthorizationWebhook{
			AuthorizationDate: datetime.DateTime(transaction.CreatedAt.UTC()),
			AuthorizationID:   idconv.ToString(transaction.AuthorizationID),
			CardID:            cardID, CardNumber: masked, ApproveCode: transaction.AuthorizationCode,
			MCC: transaction.MerchantMCC, MerchantName: transaction.MerchantName, MerchantCountry: transaction.MerchantCountry,
			BillingAmount: webhookNumber{transaction.TxAmount.Round(2)}, BillingCurrency: transaction.Currency,
			TransactionAmount: webhookNumber{transaction.TxAmount.Round(2)}, TransactionCurrency: transaction.TxCurrency,
			AuthorizationType: ping.WebhookAuth, AuthorizationStatus: ping.WebhookApproved,
			BudgetID: budgetID, RateFee: webhookNumber{decimal.Zero}, RateFeeCurrency: transaction.Currency,
			ThreeDSFee: webhookNumber{decimal.Zero}, ThreeDSFeeCurrency: transaction.Currency,
		}
		if transaction.Type == common.CardTransactionType_VOID {
			payload.AuthorizationType = ping.WebhookReversal
			payload.AuthorizationID = idconv.ToString(transaction.ID)
			originalID := idconv.ToString(transaction.AuthorizationID)
			payload.OriginalAuthorizationID = &originalID
		} else if transaction.Type != common.CardTransactionType_AUTH {
			return nil, pingerrors.ErrInvalid
		}
		if transaction.Status == common.TransactionStatus_FAILED {
			payload.AuthorizationStatus = ping.WebhookDeclined
		} else if transaction.Status != common.TransactionStatus_AUTHORIZED && transaction.Status != common.TransactionStatus_VOID && transaction.Status != common.TransactionStatus_SUCCEED {
			return nil, pingerrors.ErrInvalid
		}
		return payload, nil
	case ping.WebhookClearing:
		transaction := req.Source.Transaction
		if transaction == nil || transaction.Status != common.TransactionStatus_SUCCEED {
			return nil, pingerrors.ErrInvalid
		}
		clearType := ping.WebhookDebit
		if transaction.Type == common.CardTransactionType_REFUND {
			clearType = ping.WebhookCredit
		} else if transaction.Type != common.CardTransactionType_CLEAR {
			return nil, pingerrors.ErrInvalid
		}
		payload := &ClearingWebhook{
			ClearDate: datetime.DateTime(transaction.CreatedAt.UTC()), CardID: cardID, CardNumber: masked,
			ClearID: idconv.ToString(transaction.ID), ClearType: clearType, ClearStatus: ping.WebhookSettled,
			SettlementAmount: webhookNumber{transaction.TxAmount.Round(2)}, SettlementCurrency: transaction.Currency,
			BudgetID: budgetID, MerchantName: transaction.MerchantName, MerchantCountry: transaction.MerchantCountry, MCC: transaction.MerchantMCC,
		}
		if transaction.AuthorizationID > 0 {
			originalID := idconv.ToString(transaction.AuthorizationID)
			payload.OriginalAuthID = &originalID
		}
		return payload, nil
	default:
		return nil, pingerrors.ErrInvalid
	}
}

func toWebhookOptionalID(value *model.ID) *string {
	if value == nil {
		return nil
	}
	result := idconv.ToString(*value)
	return &result
}

func maskWebhookCardNumber(value string) string {
	if len(value) < 10 {
		return strings.Repeat("*", len(value))
	}
	return value[:6] + strings.Repeat("*", len(value)-10) + value[len(value)-4:]
}

type webhookNumber struct{ decimal.Decimal }

func (value webhookNumber) MarshalJSON() ([]byte, error) { return []byte(value.String()), nil }
