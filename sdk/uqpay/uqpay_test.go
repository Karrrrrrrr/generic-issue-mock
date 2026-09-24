package uqpay

import (
	"context"
	"encoding/json"
	"testing"

	"tman/enums"

	"github.com/shopspring/decimal"

	"github.com/google/uuid"
)

type uqConf struct {
	Url             string
	Debug           bool
	ApiKey          string
	ClientId        string
	CardLimitAmount float64
}

var (
	uqSDK *UqPaySDK
	ctx   = context.Background()
)

func (x *uqConf) GetToken() string {

	return ""
}

func (x *uqConf) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *uqConf) GetDebug() bool {
	if x != nil {
		return x.Debug
	}
	return false
}
func (x *uqConf) GetApiKey() string {
	return x.ApiKey
}
func (x *uqConf) GetClientId() string {
	return x.ClientId
}

func (x *uqConf) GetCardLimitAmount() float64 {
	return x.CardLimitAmount
}

func TestMain(m *testing.M) {
	conf := &uqConf{
		//Url: "http://localhost:8080",
		Url:             "http://test-api-uqpaytech.ptmlink.top",
		Debug:           true,
		ApiKey:          "2tXVF29q4N13bpRpg6mppmXJ7Y208Nm5Flta8kAuqyx9fvfgYhyvLUVx19MmS3ai",
		ClientId:        "lDRVA6v0xoxFT920axr95O",
		CardLimitAmount: 1000000,
	}

	uqSDK = New(conf, WithRequestMiddlewares(), WithResponseMiddlewares())
	m.Run()
}

func TestUqPaySDK_GetAccessToken(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}

	t.Logf("v: %+v\n", result)
}

func TestUqPaySDK_CreateCardholder(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	t.Logf("v: %+v\n", result)

	req := &CreateCardholderReq{
		Email:       "15969136919@gmail.com",
		FirstName:   "test first name",
		LastName:    "test last name",
		DateOfBirth: "1990-01-01",
		CountryCode: "CN",
		PhoneNumber: "15969136909",
		DeliveryAddress: &DeliveryAddress{
			City:       "City",
			Country:    "CN",
			Line1:      "浙江杭州",
			PostalCode: "310000",
		},
	}
	uuidKey := "96a2f04a-f23a-4bdf-bed0-11db64ea64e4"
	cardholder, err := uqSDK.CreateCardholder(ctx, result.AuthToken, uuidKey, req)
	if err != nil {
		t.Fatalf("创建持卡人失败: %v", err)
	}
	t.Logf("v: %+v\n", cardholder)
	// card_holder_id = ca5ddd54-3806-48ea-8726-5618d05f8e0e
	// card_holder_id = ca5ddd54-3806-48ea-8726-5618d05f8e0e
}

func TestUqPaySDK_UpdateCardholder(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	t.Logf("v: %+v\n", result)
	email := "test1@gmail.com"
	req := &UpdateCardholderReq{
		CardHolderID: "96a2f04a-f23a-4bdf-bed0-11db64ea64e4",
		Email:        &email,
		DeliveryAddress: &DeliveryAddress{
			City:       "City",
			Country:    "CN",
			Line1:      "浙江杭州",
			PostalCode: "310000",
		},
	}
	cardholder, err := uqSDK.UpdateCardholder(ctx, result.AuthToken, uuid.New().String(), req)
	if err != nil {
		msg := err.Error()
		result := &UpdateCardholderResp{}
		_ = json.Unmarshal([]byte(msg), &result)

		t.Fatalf("创建持卡人失败: %s", err.Error())
	}
	t.Logf("v: %+v\n", cardholder)
}

func TestUqPaySDK_GetCardholder(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	t.Logf("v: %+v\n", result)

	cardholder, err := uqSDK.GetCardholder(ctx, result.AuthToken, "ca5ddd54-3806-48ea-8726-5618d05f8e0e")
	if err != nil {
		msg := err.Error()
		result := &CardholderDetail{}
		_ = json.Unmarshal([]byte(msg), &result)

		t.Fatalf("获取持卡人失败: %s", err.Error())
	}
	t.Logf("v: %+v\n", cardholder)
}

func TestUqPaySDK_ListCardProducts(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}

	products, err := uqSDK.ListCardProducts(ctx, result.AuthToken, "10", "1")
	if err != nil {
		t.Fatalf("获取product list失败: %v", err)
	}
	t.Logf("v: %+v\n", products)
}

func TestUqPaySDK_CreateCard(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	t.Logf("v: %+v\n", result)

	cardLimit := uqSDK.GetDefaultCardLimitAmount(ctx)
	req := &CreateCardReq{
		CardCurrency:  "USD",
		CardholderID:  "f6dee940-8480-4341-9db5-65453fe7f4c9",
		CardProductID: "e374ef17-5e37-4269-bd84-6591e96ab716",
		CardLimit:     &cardLimit,
	}
	idempotencyKey := "e374ef17-5e37-4269-bd84-6591e96ab716"

	card, err := uqSDK.CreateCard(ctx, result.AuthToken, idempotencyKey, req)
	if err != nil {
		t.Fatalf("创建card失败: %v", err)
	}
	t.Logf("v: %+v\n", card)
}

func TestUqPaySDK_UpdateCard(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	t.Logf("v: %+v\n", result)
	cardID := "7e72d97e-55a2-4c44-ae62-486c2f83c736"
	cardLimit := 1123.02
	var req = &UpdateCardReq{
		CardID:             cardID,
		CardLimit:          &cardLimit,
		NoPinPaymentAmount: new(float64),
		SpendingControls: []*SpendingControl{
			{
				Amount:   "20000",
				Interval: "PER_TRANSACTION",
			},
		},

		RiskControls: nil,
	}
	card, err := uqSDK.UpdateCard(ctx, result.AuthToken, uuid.NewString(), req)
	if err != nil {
		t.Fatalf("更新card失败: %v", err)
	}
	t.Logf("v: %+v\n", card)
}

/*
	{
	   "available_balance": "0.00",
	   "card_bin": "40963608",
	   "card_currency": "USD",
	   "card_id": "5e4908d6-6a99-447b-b6cf-e60dc3c6d297",
	   "card_limit": 0,
	   "card_number": "************",
	   "card_product_id": "e374ef17-5e37-4269-bd84-6591e96ab716",
	   "card_scheme": "VISA",
	   "card_status": "PENDING",
	   "cardholder": {
	      "cardholder_id": "96a2f04a-f23a-4bdf-bed0-11db64ea64e4",
	      "cardholder_status": "SUCCESS",
	      "create_time": "2025-12-09T17:36:00+08:00",
	      "email": "test1@gmail.com",
	      "first_name": "test first name",
	      "last_name": "test last name"
	   },
	   "consumed_amount": "0.00",
	   "form_factor": "PHYSICAL",
	   "metadata": null,
	   "mode_type": "SHARE",
	   "no_pin_payment_amount": "USD",
	   "risk_controls": {
	      "allow_3ds_transactions": "Y"
	   },
	   "spending_controls": [
	      {
	         "interval": "PER_TRANSACTION",
	         "amount": "20000"
	      }
	   ],
	   "update_reason": ""
	}
*/
func TestUqPaySDK_GetCardInfo(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	t.Logf("v: %+v\n", result)
	cardID := "9b2ee4e1-3eed-41de-aa89-d4661bcbcf2a"

	card, err := uqSDK.GetCardInfo(ctx, result.AuthToken, cardID)
	if err != nil {
		t.Fatalf("获取card信息失败: %v", err)
	}
	t.Logf("v: %+v\n", card)
}

func TestUqPaySDK_UpdateCardStatus(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	t.Logf("v: %+v\n", result)
	req := &UpdateCardStatusReq{
		CardStatus: THIRD_PARTY_UP_CARD_STATUS_ENUM_CANCELLED,
		CardID:     "af018c19-ced4-4e76-ade3-cd61f81acb45",
	}

	card, err := uqSDK.UpdateCardStatus(ctx, result.AuthToken, uuid.New().String(), req)
	if err != nil {
		t.Fatalf("更新card 状态失败: %v", err)
	}
	t.Logf("v: %+v\n", card)
}

func TestUqPaySDK_GetCardPrivateInfo(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	t.Logf("v: %+v\n", result)
	cardID := "aabf8d60-1936-4299-bdb3-ecba5650b680"

	card, err := uqSDK.GetCardPrivateInfo(ctx, result.AuthToken, cardID)
	if err != nil {
		t.Fatalf("获取card信息失败: %v", err)
	}
	t.Logf("v: %+v\n", card)
}

/*
成功

	{
	   "card_order_id": "ec92a9ed-c846-4549-9997-4f6997f82f31",
	   "card_id": "273336dd-e747-4f66-8a5f-a93b7b39bb02",
	   "card_status": "PENDING",
	   "order_status": "PENDING",
	   "create_time": "2026-01-14T13:56:06+08:00",
	   "risk_controls": {}
	}

	{
	   "card_order_id": "64ae684d-290c-4d84-a089-22e2a3635a49",
	   "card_id": "5e4908d6-6a99-447b-b6cf-e60dc3c6d297",
	   "card_status": "PENDING",
	   "order_status": "PENDING",
	   "create_time": "2026-01-14T20:43:34+08:00",
	   "risk_controls": {}
	}

失败(重复绑定)

	{
	   "type": "invalid_request_error",
	   "code": "invalid_request_error",
	   "message": "card has been assigned"
	}
*/

/*
4096360811125926
4096360811126510
4096360811126684
4096360811127039
4096360811127971
4096360811128409
4096360811128474
4096360811129423
4096360811129753
4096360811124002
4096360811124580
4096360811125538
4096360811125660
*/
func TestUUID(t *testing.T) {
	println(uuid.New().String())

}
func TestUqPaySDK_AssignCard(t *testing.T) {
	token, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}

	result, err := uqSDK.AssignCard(context.Background(), token.AuthToken, "47122ff8-b7e9-44dd-b191-1c2464b150f4", &AssignCardReq{
		CardHolderID: "96a2f04a-f23a-4bdf-bed0-11db64ea64e4",
		CardNumber:   "4096360811126684",
		CardCurrency: enums.CURRENCY_ENUM_USD,
		CardMode:     THIRD_PARTY_UP_CARD_MODE_ENUM_SHARE, // 必须选SHARE
	})
	if err != nil {
		t.Fatalf("开实体卡失败: %v", err)
	}
	t.Logf("result: %+v\n", result)
}

/*
成功

	{
	   "request_status": "SUCCESS"
	}

失败1

	{
	   "code": "400",
	   "message": "Card does not exists"
	}

失败2

	{
	   "code": "400",
	   "message": "Invalid activation code."
	}
*/
func TestUqPaySDK_ActivateCard(t *testing.T) {
	token, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}

	result, err := uqSDK.ActivateCard(context.Background(), token.AuthToken, uuid.New().String(), &ActivateCardReq{
		ActivationCode:     "76090556",
		CardID:             "9b2ee4e1-3eed-41de-aa89-d4661bcbcf2a",
		NoPinPaymentAmount: nil,
		Pin:                "666666",
	})
	if err != nil {
		t.Fatalf("激活实体卡失败: %v", err)
	}
	t.Logf("result: %+v\n", result)
}

func TestUqPaySDK_ListTransaction(t *testing.T) {
	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	cardID := "7e72d97e-55a2-4c44-ae62-486c2f83c736"

	transaction, err := uqSDK.ListTransaction(ctx, result.AuthToken, &ListTransactionReq{
		PageNumber: 1,
		PageSize:   10,
		CardID:     &cardID,
	})
	if err != nil {
		t.Fatalf("获取card信息失败: %v", err)
	}
	t.Logf("v: %+v\n", transaction)
}

/*
	{
	   "authorization_code": "IFRG6A",
	   "billing_amount": "10",
	   "billing_currency": "USD",
	   "card_available_balance": "1039.52",
	   "card_id": "7e72d97e-55a2-4c44-ae62-486c2f83c736",
	   "card_number": "************8193",
	   "cardholder_id": "6a812eb3-e59c-49fc-bfd2-77ffee7c8cd3",
	   "failure_reason": "",
	   "merchant_data": {
	      "category_code": "5734",
	      "city": "",
	      "country": "",
	      "name": "Orchard Central"
	   },
	   "posted_time": "2026-02-12T14:25:07.114400253+08:00",
	   "transaction_amount": "10",
	   "transaction_currency": "USD",
	   "transaction_id": "1a917072-83d9-4e7f-a8db-214a69be9caf",
	   "transaction_status": "APPROVED",
	   "transaction_time": "2026-02-12T14:25:07.050236438+08:00",
	   "transaction_type": "AUTHORIZATION"
	}
*/

func TestUqPaySDK_SimulateAuthorization(t *testing.T) {

	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}
	cardID := "7e72d97e-55a2-4c44-ae62-486c2f83c736"

	authorization, err := uqSDK.SimulateAuthorization(ctx, result.AuthToken, uuid.NewString(), &SimulateAuthorizationReq{
		CardID:               cardID,
		TransactionAmount:    decimal.NewFromInt(10),
		TransactionCurrency:  "USD",
		MerchantName:         "Orchard Central",
		MerchantCategoryCode: "5734",
	})
	if err != nil {
		t.Fatalf("获取authorization信息失败: %v", err)
	}
	t.Logf("v: %+v\n", authorization)
}

/*
{
   "authorization_code": "PECMHY",
   "billing_amount": "10",
   "billing_currency": "USD",
   "card_available_balance": "1049.52",
   "card_id": "7e72d97e-55a2-4c44-ae62-486c2f83c736",
   "card_number": "************8193",
   "cardholder_id": "6a812eb3-e59c-49fc-bfd2-77ffee7c8cd3",
   "failure_reason": "",
   "merchant_data": {
      "category_code": "5734",
      "city": "",
      "country": "",
      "name": "Orchard Central"
   },
   "origin_transaction_id": "1a917072-83d9-4e7f-a8db-214a69be9caf",
   "posted_time": "2026-02-12T14:46:25.466308377+08:00",
   "transaction_amount": "10",
   "transaction_currency": "USD",
   "transaction_id": "2c6b696a-c81f-44af-a2d1-57d7aff90d0b",
   "transaction_status": "APPROVED",
   "transaction_time": "2026-02-12T14:46:25.466313736+08:00",
   "transaction_type": "REVERSAL"
}
*/

/*
	{
		"code": "400",
		"message": "Card reversal transaction is error"
	}
*/
func TestUqPaySDK_SimulateReversal(t *testing.T) {

	result, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}

	authorization, err := uqSDK.SimulateReversal(ctx, result.AuthToken, uuid.NewString(), &SimulateReversalReq{
		TransactionID: "1a917072-83d9-4e7f-a8db-214a69be9caf",
	})
	if err != nil {
		t.Fatalf("获取authorization信息失败: %v", err)
	}
	t.Logf("v: %+v\n", authorization)
}

func TestUqPaySDK_RetrieveBalance(t *testing.T) {
	token, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}

	result, err := uqSDK.RetrieveBalance(context.Background(), token.AuthToken, uuid.New().String(), &RetrieveBalanceReq{
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("取回余额失败: %v", err)
	}
	t.Logf("result: %+v\n", result)
}

func TestUqPaySDK_RetrieveIssuingBalance(t *testing.T) {
	token, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}

	result, err := uqSDK.RetrieveIssuingBalance(context.Background(), token.AuthToken, uuid.New().String(), &RetrieveIssuingBalanceReq{
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("取回余额失败: %v", err)
	}
	t.Logf("result: %+v\n", result)
}

func TestUqPaySDK_CreatePanToken(t *testing.T) {
	token, err := uqSDK.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}

	cardID := "e42a1901-19d4-4483-81fb-f10c3cbb397b"
	result, err := uqSDK.CreatePanToken(context.Background(), token.AuthToken, cardID)
	if err != nil {
		t.Fatalf("CreatePanToken失败: %v", err)
	}

	t.Logf("result: %+v\n", result)
}
