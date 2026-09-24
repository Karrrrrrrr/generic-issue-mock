package slash

import (
	"context"
	"testing"
)

func TestCardStatusUpdateEmptyID(testContext *testing.T) {
	client := NewSlashSDK(false, SlashSDKConf{})
	for _, name := range []string{"freeze", "unfreeze"} {
		testContext.Run(name, func(testContext *testing.T) {
			operation := client.CardFrozen
			if name == "unfreeze" {
				operation = client.CardUnfrozen
			}
			if err := operation(context.Background(), ""); err != ErrEmptyCardID {
				testContext.Fatalf("got %v, want %v", err, ErrEmptyCardID)
			}
		})
	}
}
