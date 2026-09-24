package cardwallet

import (
	"generic-mock/enums"
	"generic-mock/model"
)

type PrepareRequest struct {
	AccountID      model.ID
	Channel        enums.Channel
	CardType       enums.CardType
	Currency       enums.Currency
	VirtualAccount *model.VirtualAccount
}

type PrepareResult struct {
	VirtualAccountID *model.ID
	Wallet           *model.Wallet
	CreateWallet     bool
}

func Prepare(req PrepareRequest) (PrepareResult, bool) {
	if req.AccountID <= 0 || req.Channel == "" || req.Currency == "" {
		return PrepareResult{}, false
	}

	result := PrepareResult{
		Wallet: &model.Wallet{
			AccountID: req.AccountID,
			Channel:   req.Channel,
			Type:      enums.WalletType_Card,
			Currency:  req.Currency,
		},
		CreateWallet: true,
	}
	switch req.CardType {
	case enums.CardType_Single:
		if req.VirtualAccount != nil {
			return PrepareResult{}, false
		}
		return result, true
	case enums.CardType_Share, enums.CardType_VirtualAccountSingle:
		virtualAccount := req.VirtualAccount
		if virtualAccount == nil || virtualAccount.ID <= 0 ||
			virtualAccount.AccountID != req.AccountID || virtualAccount.Channel != req.Channel {
			return PrepareResult{}, false
		}
		wallet := virtualAccount.Wallet
		if wallet == nil || wallet.ID <= 0 || wallet.ID != virtualAccount.WalletID ||
			wallet.AccountID != req.AccountID || wallet.Channel != req.Channel ||
			wallet.Currency != req.Currency || wallet.Type != enums.WalletType_VirtualAccount {
			return PrepareResult{}, false
		}
		virtualAccountID := virtualAccount.ID
		result.VirtualAccountID = &virtualAccountID
		if req.CardType == enums.CardType_Share {
			result.Wallet = wallet
			result.CreateWallet = false
		}
		return result, true
	default:
		return PrepareResult{}, false
	}
}

func FundingWalletID(card *model.Card) (model.ID, bool) {
	if card == nil || card.AccountID <= 0 || card.Channel == "" || card.WalletID <= 0 {
		return 0, false
	}
	if card.CardType == enums.CardType_VirtualAccountSingle {
		virtualAccount := card.VirtualAccount
		if card.VirtualAccountID == nil || *card.VirtualAccountID <= 0 || virtualAccount == nil ||
			virtualAccount.ID != *card.VirtualAccountID || virtualAccount.AccountID != card.AccountID ||
			virtualAccount.Channel != card.Channel || virtualAccount.WalletID <= 0 ||
			virtualAccount.WalletID == card.WalletID {
			return 0, false
		}
		return virtualAccount.WalletID, true
	}
	if card.CardType != enums.CardType_Single && card.CardType != enums.CardType_Share {
		return 0, false
	}
	account := card.Account
	if account == nil || account.ID != card.AccountID || account.Channel != card.Channel ||
		account.WalletID <= 0 || account.WalletID == card.WalletID {
		return 0, false
	}
	return account.WalletID, true
}
