package enums

type WalletTransferKind string

const (
	WalletTransfer_CardTopUp              WalletTransferKind = "card_top_up"
	WalletTransfer_CardWithdraw           WalletTransferKind = "card_withdraw"
	WalletTransfer_VirtualAccountTopUp    WalletTransferKind = "virtual_account_top_up"
	WalletTransfer_VirtualAccountTransfer WalletTransferKind = "virtual_account_transfer"
)
