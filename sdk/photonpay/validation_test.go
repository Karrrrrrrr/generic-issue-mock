package photonpay

import "testing"

func TestListSDKValidation(testContext *testing.T) {
	cardType := "share"
	formFactor := "virtual_card"
	requests := []struct {
		name     string
		validate func() error
	}{
		{
			name:     "card type filter",
			validate: (&PagingVccCardRequest{CardType: &cardType}).Validate,
		},
		{
			name:     "card form filter",
			validate: (&PagingVccCardRequest{CardFormFactor: &formFactor}).Validate,
		},
		{
			name:     "funds form filter",
			validate: (&PagingRechargeCardFundsDetailRequest{CardFormFactor: &formFactor}).Validate,
		},
		{
			name:     "trade form filter",
			validate: (&PagingVccTradeOrderRequest{CardFormFactor: &formFactor}).Validate,
		},
	}
	for _, request := range requests {
		testContext.Run(request.name, func(testContext *testing.T) {
			if err := request.validate(); err == nil {
				testContext.Fatal("SDK no longer rejects this filter; add a live filtered-call case")
			}
		})
	}
}
