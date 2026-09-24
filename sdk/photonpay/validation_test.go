package photonpay

import "testing"

func TestListSDKValidation(testContext *testing.T) {
	cardType := "share"
	formFactor := "virtual_card"
	testContext.Run("card-type-filter", func(testContext *testing.T) {
		request := &PagingVccCardRequest{CardType: &cardType}
		if err := request.Validate(); err == nil {
			testContext.Fatal("SDK no longer rejects this filter; add a live filtered-call case")
		}
	})
	testContext.Run("card-form-filter", func(testContext *testing.T) {
		request := &PagingVccCardRequest{CardFormFactor: &formFactor}
		if err := request.Validate(); err == nil {
			testContext.Fatal("SDK no longer rejects this filter; add a live filtered-call case")
		}
	})
	testContext.Run("funds-form-filter", func(testContext *testing.T) {
		request := &PagingRechargeCardFundsDetailRequest{CardFormFactor: &formFactor}
		if err := request.Validate(); err == nil {
			testContext.Fatal("SDK no longer rejects this filter; add a live filtered-call case")
		}
	})
	testContext.Run("trade-form-filter", func(testContext *testing.T) {
		request := &PagingVccTradeOrderRequest{CardFormFactor: &formFactor}
		if err := request.Validate(); err == nil {
			testContext.Fatal("SDK no longer rejects this filter; add a live filtered-call case")
		}
	})
	testContext.Run("trade-type-filter", func(testContext *testing.T) {
		transactionType := "refund"
		request := &PagingVccTradeOrderRequest{TransactionType: &transactionType}
		if err := request.Validate(); err == nil {
			testContext.Fatal("SDK no longer rejects this filter; add a live filtered-call case")
		}
	})
}
