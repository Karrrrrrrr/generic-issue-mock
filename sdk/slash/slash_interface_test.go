package slash

import (
	"context"
	"testing"
)

func TestCardStatusUpdateEmptyID(testContext *testing.T) {
	client := NewSlashSDK(false, SlashSDKConf{})
	testContext.Run("freeze", func(testContext *testing.T) {
		if err := client.CardFrozen(context.Background(), ""); err != ErrEmptyCardID {
			testContext.Fatalf("got %v, want %v", err, ErrEmptyCardID)
		}
	})
	testContext.Run("unfreeze", func(testContext *testing.T) {
		if err := client.CardUnfrozen(context.Background(), ""); err != ErrEmptyCardID {
			testContext.Fatalf("got %v, want %v", err, ErrEmptyCardID)
		}
	})
}
