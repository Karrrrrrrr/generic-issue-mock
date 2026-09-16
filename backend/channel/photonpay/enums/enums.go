package enums

import generic "generic-mock/enums"

type ResponseCode string

const (
	ResponseCode_Success  ResponseCode = "0000"
	ResponseCode_BadInput ResponseCode = "4000"
	ResponseCode_NotFound ResponseCode = "VCC1039"
)

type AccountType string

const (
	AccountType_Available AccountType = "FT10001"
)

type CardType string

const (
	CardType_Share    CardType = "share"
	CardType_Recharge CardType = "recharge"
)

type CardFormFactor string

const (
	CardFormFactor_Virtual  CardFormFactor = "virtual_card"
	CardFormFactor_Physical CardFormFactor = "physical_card"
)

type CardStatus string

const (
	CardStatus_Normal    CardStatus = "normal"
	CardStatus_Freezing  CardStatus = "freezing"
	CardStatus_Frozen    CardStatus = "frozen"
	CardStatus_Cancelled CardStatus = "cancelled"
)

type OperationStatus string

const (
	OperationStatus_Succeed OperationStatus = "succeed"
)

const (
	MemberID       = "photonpay-mock-member"
	AccountNumber  = "photonpay-mock-account"
	DefaultCardBin = "543210"
	CardScheme     = "MasterCard"
)

func CardTypeFromGeneric(value generic.CardType) CardType {
	if value == generic.CardType_Share {
		return CardType_Share
	}

	return CardType_Recharge
}

func CardTypeToGeneric(value CardType) generic.CardType {
	if value == CardType_Share {
		return generic.CardType_Share
	}

	return generic.CardType_Single
}

func CardFormFactorFromGeneric(value generic.CardFormType) CardFormFactor {
	if value == generic.CardFormType_Physical {
		return CardFormFactor_Physical
	}

	return CardFormFactor_Virtual
}

func CardFormFactorToGeneric(value CardFormFactor) generic.CardFormType {
	if value == CardFormFactor_Physical {
		return generic.CardFormType_Physical
	}

	return generic.CardFormType_Virtual
}

func CardStatusFromGeneric(value generic.CardStatus) CardStatus {
	switch value {
	case generic.CardStatus_Frozing:
		return CardStatus_Freezing
	case generic.CardStatus_Frozen:
		return CardStatus_Frozen
	case generic.CardStatus_Deleted:
		return CardStatus_Cancelled
	default:
		return CardStatus_Normal
	}
}
