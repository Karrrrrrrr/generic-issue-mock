package service

import (
	"context"
	"encoding/base64"
	stderrors "errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"generic-mock/channel/photonpay/biz"
	photon "generic-mock/channel/photonpay/enums"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/channel/photonpay/pkg/idconv"
	"generic-mock/channel/photonpay/pkg/queryconv"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/timefmt"
	"generic-mock/pkg/types"
	timeTypes "generic-mock/pkg/types/time"
	sharedbiz "generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/samber/do/v2"
)

type PhotonPayOpenAPIService struct {
	usecase     *biz.PhotonPayOpenAPIUsecase
	cardIssuer  *sharedbiz.CardIssuer
	notificator *biz.PhotonPayWebhookNotificator
}

type OpenAPIAccountRequest struct {
	Token string `header:"X-PD-TOKEN" binding:"required"`
}

func NewPhotonPayOpenAPIService(injector do.Injector) (*PhotonPayOpenAPIService, error) {
	return &PhotonPayOpenAPIService{
		usecase:     do.MustInvoke[*biz.PhotonPayOpenAPIUsecase](injector),
		cardIssuer:  do.MustInvoke[*sharedbiz.CardIssuer](injector),
		notificator: do.MustInvoke[*biz.PhotonPayWebhookNotificator](injector),
	}, nil
}

type AccessTokenRequest struct {
	AppID         *string `form:"app_id" json:"app_id"`
	Secret        *string `form:"secret" json:"secret"` // Invalid: secrets are not validated by the mock.
	Authorization *string `header:"Authorization"`
}

type AccessTokenData struct {
	ExpiresIn        int64  `json:"expiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	Token            string `json:"token"`
}

func (s *PhotonPayOpenAPIService) AccessToken(_ context.Context, req *AccessTokenRequest) (*AccessTokenData, error) {
	selector := types.Value(req.AppID)
	if req.AppID == nil && req.Authorization != nil {
		scheme, credentials, valid := strings.Cut(*req.Authorization, " ")
		if !valid || !strings.EqualFold(scheme, "basic") {
			return nil, photonpayerrors.ErrInvalidOperation
		}
		decoded, err := base64.StdEncoding.DecodeString(credentials)
		if err != nil {
			return nil, photonpayerrors.ErrInvalidOperation
		}
		selector, _, valid = strings.Cut(string(decoded), "/")
		if !valid {
			return nil, photonpayerrors.ErrInvalidOperation
		}
	}
	accountID, err := idconv.FromAccountString(selector)
	if err != nil {
		return nil, err
	}
	token := idconv.ToString(accountID)
	return &AccessTokenData{
		ExpiresIn:        3600,
		RefreshExpiresIn: 7200,
		RefreshToken:     token,
		Token:            token,
	}, nil
}

type AccountSingleRequest struct {
	OpenAPIAccountRequest
	Currency      *string             `form:"currency" json:"currency"`
	AccountNo     *string             `form:"accountNo" json:"accountNo"` // Invalid: mock has one generic account.
	MemberID      *string             `form:"memberId" json:"memberId"`   // Invalid: mock does not partition by member.
	AccountType   *photon.AccountType `form:"accountType" json:"accountType"`
	MatrixAccount *string             `form:"matrixAccount" json:"matrixAccount"` // Invalid: Matrix is unsupported.
}

type AccountSingleData struct {
	MemberID        string                `json:"memberId"`
	AccountNo       string                `json:"accountNo"`
	AccountType     photon.AccountType    `json:"accountType"`
	Currency        common.Currency       `json:"currency"`
	RealTimeBalance float64               `json:"realTimeBalance"`
	ReturnedAt      timeTypes.ISODateTime `json:"returnedAt"`
}

func (s *PhotonPayOpenAPIService) AccountSingle(_ context.Context, req *AccountSingleRequest) (*AccountSingleData, error) {
	currency := common.Currency_USD
	if req.Currency != nil {
		currency = common.Currency(*req.Currency)
	}
	accountType := photon.AccountType_Available
	if req.AccountType != nil {
		accountType = *req.AccountType
	}
	return &AccountSingleData{
		MemberID:        photon.MemberID,
		AccountNo:       photon.AccountNumber,
		AccountType:     accountType,
		Currency:        currency,
		RealTimeBalance: biz.DefaultBalance().InexactFloat64(),
		ReturnedAt:      timeTypes.ISODateTime(time.Now().UTC()),
	}, nil
}

type CreateCardHolderRequest struct {
	OpenAPIAccountRequest
	MemberID                   *string `json:"memberId"`      // Invalid: mock does not partition by member.
	MatrixAccount              *string `json:"matrixAccount"` // Invalid: Matrix is unsupported.
	FirstName                  string  `json:"firstName" binding:"required"`
	LastName                   string  `json:"lastName" binding:"required"`
	CardholderNameAbbreviation *string `json:"cardholderNameAbbreviation"` // Invalid: physical card printing is unsupported.
	Email                      string  `json:"email" binding:"required"`
	Mobile                     string  `json:"mobile" binding:"required"`
	MobilePrefix               string  `json:"mobilePrefix" binding:"required"`
	DateOfBirth                string  `json:"dateOfBirth" binding:"required"`
	CertType                   *string `json:"certType"`
	Portrait                   *string `json:"portrait"`
	ReverseSide                *string `json:"reverseSide"`
	NationalityCountryCode     string  `json:"nationalityCountryCode" binding:"required"`
	ResidentialAddress         *string `json:"residentialAddress"`
	ResidentialCity            *string `json:"residentialCity"`
	ResidentialCountryCode     *string `json:"residentialCountryCode"`
	ResidentialPostalCode      *string `json:"residentialPostalCode"`
	ResidentialState           *string `json:"residentialState"`
	CertCountryCode            *string `json:"certCountryCode"`
	CertID                     *string `json:"certId"`
}

type CardHolderData struct {
	CardholderID           string                        `json:"cardholderId"`
	MemberID               string                        `json:"memberId"`
	Status                 photon.CardHolderStatus       `json:"status"`
	CardholderReviewStatus photon.CardHolderReviewStatus `json:"cardholderReviewStatus"`
	IdInfoRequirement      string                        `json:"idInfoRequirement"`
	Reason                 string                        `json:"reason"`
}

func (s *PhotonPayOpenAPIService) CreateCardHolder(ctx context.Context, req *CreateCardHolderRequest) (*CardHolderData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	dateOfBirth, valid := queryconv.RequiredDate(req.DateOfBirth)
	if !valid {
		return nil, photonpayerrors.ErrInvalidDateOfBirth
	}

	holder, err := s.usecase.CreateCardHolder(ctx, &biz.CreateCardHolderRequest{
		AccountID:              accountID,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		Mobile:                 req.Mobile,
		MobilePrefix:           req.MobilePrefix,
		DateOfBirth:            dateOfBirth,
		NationalityCountryCode: req.NationalityCountryCode,
		ResidentialAddress:     types.Value(req.ResidentialAddress),
		ResidentialCity:        types.Value(req.ResidentialCity),
		ResidentialCountryCode: types.Value(req.ResidentialCountryCode),
		ResidentialPostalCode:  types.Value(req.ResidentialPostalCode),
		ResidentialState:       types.Value(req.ResidentialState),
		CertType:               types.Value(req.CertType),
		CertCountryCode:        types.Value(req.CertCountryCode),
		CertID:                 types.Value(req.CertID),
		Portrait:               types.Value(req.Portrait),
		ReverseSide:            types.Value(req.ReverseSide),
	})
	if err != nil {
		return nil, err
	}

	return cardHolderData(holder), nil
}

type EditCardHolderRequest struct {
	OpenAPIAccountRequest
	CardholderID string  `json:"cardholderId" binding:"required"`
	Email        *string `json:"email"`
	Mobile       *string `json:"mobile"`
	MobilePrefix *string `json:"mobilePrefix"`
}

func (s *PhotonPayOpenAPIService) EditCardHolder(ctx context.Context, req *EditCardHolderRequest) (*CardHolderData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	cardholderID, err := idconv.FromString(req.CardholderID)
	if err != nil {
		return nil, photonpayerrors.ErrInvalidCardholderID
	}
	holder, err := s.usecase.UpdateCardHolder(ctx, &biz.UpdateCardHolderRequest{
		AccountID:    accountID,
		CardholderID: cardholderID,
		Email:        req.Email,
		Mobile:       req.Mobile,
		MobilePrefix: req.MobilePrefix,
	})
	if err != nil {
		return nil, err
	}

	return cardHolderData(holder), nil
}

type ListCardHolderRequest struct {
	OpenAPIAccountRequest
	PageIndex *int `form:"pageIndex" binding:"omitempty,min=1"`
	PageSize  *int `form:"pageSize" binding:"omitempty,min=1"`
}
type CardHolderListItem struct {
	CardholderID           string                        `json:"cardholderId"`
	CreatedAt              timeTypes.ISODateTime         `json:"createdAt"`
	FirstName              string                        `json:"firstName"`
	LastName               string                        `json:"lastName"`
	Email                  string                        `json:"email"`
	Mobile                 string                        `json:"mobile"`
	MobilePrefix           string                        `json:"mobilePrefix"`
	Status                 photon.CardHolderStatus       `json:"status"`
	CardholderReviewStatus photon.CardHolderReviewStatus `json:"cardholderReviewStatus"`
}

func (s *PhotonPayOpenAPIService) ListCardHolders(ctx context.Context, req *ListCardHolderRequest) (*[]CardHolderListItem, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	page, size := types.NormalizePagination(types.Value(req.PageIndex), types.Value(req.PageSize))
	holders, err := s.usecase.ListCardHolders(ctx, &biz.ListRequest{
		AccountID: &accountID,
		Offset:    (page - 1) * size,
		Limit:     size,
	})
	if err != nil {
		return nil, err
	}
	items := make([]CardHolderListItem, 0, len(holders))
	for _, holder := range holders {
		items = append(items, CardHolderListItem{
			CardholderID:           idconv.ToString(holder.ID),
			CreatedAt:              timeTypes.ISODateTime(holder.CreatedAt.UTC()),
			FirstName:              holder.FirstName,
			LastName:               holder.LastName,
			Email:                  holder.Email,
			Mobile:                 holder.Mobile,
			MobilePrefix:           holder.MobilePrefix,
			Status:                 photon.ConvertGenericCardHolderStatusToCardHolderStatus(holder.Status),
			CardholderReviewStatus: photon.ConvertGenericCardHolderReviewStatusToCardHolderReviewStatus(holder.ReviewStatus),
		})
	}
	return &items, nil
}

type CardBinRequest struct {
	Token          string                 `header:"X-PD-TOKEN"` // Invalid: products are channel-level; no account selection or token validation.
	CardType       *photon.CardType       `form:"cardType"`
	CardFormFactor *photon.CardFormFactor `form:"cardFormFactor"`
	CardCurrency   *common.Currency       `form:"cardCurrency"`
}
type CardBinData struct {
	CardBin                string                `json:"cardBin"`
	CardCurrency           common.Currency       `json:"cardCurrency"`
	CardScheme             common.CardScheme     `json:"cardScheme"`
	CardType               photon.CardType       `json:"cardType"`
	CardFormFactor         photon.CardFormFactor `json:"cardFormFactor"`
	RemainingAvailableCard string                `json:"remainingAvailableCard"`
}

func (s *PhotonPayOpenAPIService) CardBins(ctx context.Context, req *CardBinRequest) (*[]CardBinData, error) {
	products, err := s.usecase.ListCardProducts(ctx)
	if err != nil {
		return nil, err
	}
	data := types.BulkConvertSlice(products, func(item *model.CardProduct) CardBinData {
		return CardBinData{
			CardBin:                item.Prefix,
			CardCurrency:           common.Currency_USD,
			CardScheme:             photon.CardScheme,
			CardType:               photon.CardType_Share,
			CardFormFactor:         photon.CardFormFactor_Virtual,
			RemainingAvailableCard: photon.RemainingAvailableCardUnlimited,
		}
	})

	return &data, nil
}

type OpenCardRequest struct {
	OpenAPIAccountRequest
	MemberID             *string               `json:"memberId"`      // Invalid: mock does not partition by member.
	MatrixAccount        *string               `json:"matrixAccount"` // Invalid: Matrix is unsupported.
	CardBin              string                `json:"cardBin" binding:"required"`
	CardCurrency         common.Currency       `json:"cardCurrency" binding:"required"`
	CardExpirationDate   *int                  `json:"cardExpirationDate"`
	CardScheme           common.CardScheme     `json:"cardScheme"`
	CardType             photon.CardType       `json:"cardType" binding:"required"`
	CardFormFactor       photon.CardFormFactor `json:"cardFormFactor"`
	CardholderID         string                `json:"cardholderId" binding:"required"`
	CardDesignID         *string               `json:"cardDesignId"`   // Invalid: custom card artwork is unsupported.
	CardLogoID           *string               `json:"cardLogoId"`     // Invalid: custom card artwork is unsupported.
	MaxOnDaily           *int64                `json:"maxOnDaily"`     // Invalid: velocity limits are unsupported.
	MaxOnMonthly         *int64                `json:"maxOnMonthly"`   // Invalid: velocity limits are unsupported.
	MaxOnPercent         *int64                `json:"maxOnPercent"`   // Invalid: velocity limits are unsupported.
	RechargeAmount       *float64              `json:"rechargeAmount"` // Invalid: recharging is unsupported.
	RequestID            string                `json:"requestId" binding:"required"`
	TransactionLimit     *float64              `json:"transactionLimit"`     // Invalid: velocity limits are unsupported.
	TransactionLimitType *string               `json:"transactionLimitType"` // Invalid: velocity limits are unsupported.
	AccountID            *string               `json:"accountId"`            // Invalid: mock has one generic account.
	ArrivalAmount        *float64              `json:"arrivalAmount"`        // Invalid: recharging is unsupported.
	RecipientID          *string               `json:"recipientId"`          // Invalid: delivery is unsupported.
}

type CardData struct {
	CardBalance    float64               `json:"cardBalance"`
	CardID         string                `json:"cardId"`
	CardNo         string                `json:"cardNo"`
	CVV            string                `json:"cvv"`
	ExpirationDate string                `json:"expirationDate"`
	CardCurrency   common.Currency       `json:"cardCurrency"`
	CardScheme     common.CardScheme     `json:"cardScheme"`
	CardStatus     photon.CardStatus     `json:"cardStatus"`
	CardFormFactor photon.CardFormFactor `json:"cardFormFactor"`
	CardType       photon.CardType       `json:"cardType"`
	CardholderID   string                `json:"cardholderId"`
	CreatedAt      timeTypes.ISODateTime `json:"createdAt"`
	MaskCardNo     string                `json:"maskCardNo"`
}
type OpenCardData struct {
	CardDetail *CardData              `json:"cardDetail"`
	RequestID  string                 `json:"requestId"`
	Status     photon.OperationStatus `json:"status"`
}

func (s *PhotonPayOpenAPIService) OpenCard(ctx context.Context, req *OpenCardRequest) (*OpenCardData, error) {
	fmt.Printf("%+v\n", req)
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	cardholderID, err := idconv.FromString(req.CardholderID)
	if err != nil {
		return nil, photonpayerrors.ErrInvalidCardholderID
	}
	formFactor := req.CardFormFactor
	if formFactor == "" {
		formFactor = photon.CardFormFactor_Virtual
	}
	card, err := s.usecase.GetRequestResult(ctx, &biz.RequestResultResourceRequest{
		AccountID: accountID,
		RequestID: req.RequestID,
	})
	if err != nil {
		if !kratosErrors.IsNotFound(err) {
			return nil, err
		}
		product, err := s.usecase.FindCardProductByBinPrefix(ctx, req.CardBin)
		if err != nil {
			return nil, err
		}
		virtualAccount, err := s.usecase.FindDefaultVirtualAccount(ctx, accountID)
		if err != nil {
			return nil, err
		}
		cardScheme := req.CardScheme
		if cardScheme == "" {
			cardScheme = photon.CardScheme
		}
		months := types.Value(req.CardExpirationDate)
		if months == 0 {
			months = 24
		}
		result, err := s.cardIssuer.Issue(ctx, &sharedbiz.IssueCardReq{
			IssueCardHolder: &sharedbiz.IssueCardHolder{
				ID: &cardholderID,
			},
			Channel:          common.Channel_PhotonPay,
			AccountID:        accountID,
			CardType:         common.CardType_Share,
			VirtualAccountID: &virtualAccount.ID,
			CardProductID:    product.ID,
			Currency:         req.CardCurrency,
			CardScheme:       cardScheme,
			FormType:         photon.ConvertCardFormFactorToGenericCardFormType(formFactor),
			Status:           common.CardStatus_Active,
			ExpireAt:         time.Now().UTC().AddDate(0, months, 0),
			RequestID:        &req.RequestID,
			Notificator:      s.notificator,
		})
		if err != nil {
			return nil, convertPhotonPayIssueCardError(err)
		}
		card = result.Card
	}

	return &OpenCardData{
		CardDetail: cardData(card),
		RequestID:  card.RequestID,
		Status:     photon.OperationStatus_Succeed,
	}, nil
}

func convertPhotonPayIssueCardError(err error) error {
	switch {
	case stderrors.Is(err, sharederrors.ErrCardHolderNotFound):
		return photonpayerrors.ErrInvalidCardholderID
	case stderrors.Is(err, sharederrors.ErrAccountNotFound),
		stderrors.Is(err, sharederrors.ErrCardProductNotFound),
		stderrors.Is(err, sharederrors.ErrVirtualAccountNotFound):
		return photonpayerrors.ErrResourceNotFound
	case stderrors.Is(err, sharederrors.ErrDatabaseOperation):
		return photonpayerrors.ErrDatabaseOperation
	default:
		return photonpayerrors.ErrInvalidOperation
	}
}

type CardIDRequest struct {
	OpenAPIAccountRequest
	CardID string `form:"cardId" json:"cardId" binding:"required"`
}

type CardDetailData struct {
	*CardData
	CardBalance               string                `json:"cardBalance"`
	AvailableTransactionLimit string                `json:"availableTransactionLimit"` // Invalid: velocity limits are not persisted.
	TotalTransactionLimit     string                `json:"totalTransactionLimit"`     // Invalid: velocity limits are not persisted.
	UpdatedAt                 timeTypes.ISODateTime `json:"updateAt"`
}

func (s *PhotonPayOpenAPIService) CardDetail(ctx context.Context, req *CardIDRequest) (*CardDetailData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.GetCard(ctx, &biz.ResourceRequest{
		AccountID: &accountID,
		ID:        cardID,
	})
	if err != nil {
		return nil, err
	}

	return &CardDetailData{
		CardData:                  cardData(card),
		CardBalance:               "0",
		AvailableTransactionLimit: "0",
		TotalTransactionLimit:     "0",
		UpdatedAt:                 timeTypes.ISODateTime(card.UpdatedAt.UTC()),
	}, nil
}

func (s *PhotonPayOpenAPIService) CardCVV(ctx context.Context, req *CardIDRequest) (*CardData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.GetCard(ctx, &biz.ResourceRequest{
		AccountID: &accountID,
		ID:        cardID,
	})
	if err != nil {
		return nil, err
	}

	return cardData(card), nil
}

type ListCardsRequest struct {
	OpenAPIAccountRequest
	PageIndex  *int               `form:"pageIndex" binding:"omitempty,min=1"`
	PageSize   *int               `form:"pageSize" binding:"omitempty,min=1"`
	CardBin    *string            `form:"cardBin"`    // Invalid: card BIN filtering is unsupported.
	CardType   *photon.CardType   `form:"cardType"`   // Invalid: card type filtering is unsupported.
	CardStatus *photon.CardStatus `form:"cardStatus"` // Invalid: card status filtering is unsupported.
}

func (s *PhotonPayOpenAPIService) ListCards(ctx context.Context, req *ListCardsRequest) (*[]*CardData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	page, size := types.NormalizePagination(types.Value(req.PageIndex), types.Value(req.PageSize))
	cards, err := s.usecase.ListCards(ctx, &biz.ListRequest{
		AccountID: &accountID,
		Offset:    (page - 1) * size,
		Limit:     size,
	})
	if err != nil {
		return nil, err
	}

	items := types.BulkConvertSlice(cards, cardData)

	return &items, nil
}

type UpdateCardRequest struct {
	OpenAPIAccountRequest
	CardID                     string                 `json:"cardId" binding:"required"`
	RequestID                  string                 `json:"requestId" binding:"required"`
	CardFormFactor             *photon.CardFormFactor `json:"cardFormFactor"`             // Invalid: changing card form factor is unsupported.
	MaxOnDaily                 *int64                 `json:"maxOnDaily"`                 // Invalid: velocity limits are unsupported.
	MaxOnMonthly               *int64                 `json:"maxOnMonthly"`               // Invalid: velocity limits are unsupported.
	MaxOnPercent               *int64                 `json:"maxOnPercent"`               // Invalid: velocity limits are unsupported.
	Nickname                   *string                `json:"nickname"`                   // Invalid: card nickname is unsupported.
	TransactionLimit           *float64               `json:"transactionLimit"`           // Invalid: transaction limits are unsupported.
	TransactionLimitChangeType *string                `json:"transactionLimitChangeType"` // Invalid: transaction limits are unsupported.
	TransactionLimitType       *string                `json:"transactionLimitType"`       // Invalid: transaction limits are unsupported.
}

func (s *PhotonPayOpenAPIService) UpdateCard(ctx context.Context, req *UpdateCardRequest) (*OpenCardData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.UpdateCard(ctx, &biz.UpdateCardRequest{
		AccountID: accountID,
		CardID:    cardID,
		RequestID: req.RequestID,
	})
	if err != nil {
		return nil, err
	}

	return &OpenCardData{
		CardDetail: cardData(card),
		RequestID:  req.RequestID,
		Status:     photon.OperationStatus_Succeed,
	}, nil
}

type RequestResultRequest struct {
	OpenAPIAccountRequest
	RequestID string                   `form:"requestId" json:"requestId" binding:"required"`
	Type      photon.RequestResultType `form:"type" json:"type"`
}

func (s *PhotonPayOpenAPIService) RequestResult(ctx context.Context, req *RequestResultRequest) (*OpenCardData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.GetRequestResult(ctx, &biz.RequestResultResourceRequest{
		AccountID: accountID,
		RequestID: req.RequestID,
	})
	if err != nil {
		if kratosErrors.IsNotFound(err) {
			return nil, photonpayerrors.ErrRequestResultNotFound
		}
		return nil, err
	}

	return &OpenCardData{
		CardDetail: cardData(card),
		RequestID:  card.RequestID,
		Status:     photon.OperationStatus_Succeed,
	}, nil
}

type ChangeCardStatusRequest struct {
	OpenAPIAccountRequest
	CardID    string              `json:"cardId" binding:"required"`
	RequestID string              `json:"requestId" binding:"required"`
	Status    photon.FreezeStatus `json:"status" binding:"required,oneof=freeze unfreeze"`
}

func (s *PhotonPayOpenAPIService) FreezeCard(ctx context.Context, req *ChangeCardStatusRequest) (*CardData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.ChangeCardStatus(ctx, &biz.ChangeCardStatusRequest{
		AccountID: accountID,
		CardID:    cardID,
		RequestID: req.RequestID,
		Status:    photon.ConvertFreezeStatusToGenericCardStatus(req.Status),
		Operation: common.OperationType_FreezeCard,
	})
	if err != nil {
		return nil, err
	}
	_ = s.notificator.NotifyCardStatus(ctx, &sharedbiz.NotifyCardStatusReq{
		AccountID: accountID,
		Channel:   common.Channel_PhotonPay,
		CardID:    card.ID,
		Status:    card.Status,
	})

	return cardData(card), nil
}

type CancelCardRequest struct {
	OpenAPIAccountRequest
	CardID string `json:"cardId" binding:"required"`
}

func (s *PhotonPayOpenAPIService) CancelCard(ctx context.Context, req *CancelCardRequest) (*CardData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromString(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.ChangeCardStatus(ctx, &biz.ChangeCardStatusRequest{
		AccountID: accountID,
		CardID:    cardID,
		RequestID: "",
		Status:    common.CardStatus_Deleted,
		Operation: common.OperationType_CancelCard,
	})
	if err != nil {
		return nil, err
	}
	_ = s.notificator.NotifyCardStatus(ctx, &sharedbiz.NotifyCardStatusReq{
		AccountID: accountID,
		Channel:   common.Channel_PhotonPay,
		CardID:    card.ID,
		Status:    card.Status,
	})

	return cardData(card), nil
}

type ListTradeRequest struct {
	OpenAPIAccountRequest
	PageIndex       *int    `form:"pageIndex" binding:"omitempty,min=1"`
	PageSize        *int    `form:"pageSize" binding:"omitempty,min=1"`
	MemberID        *string `form:"memberId"`      // Invalid: the token selects the account; Matrix members are unsupported.
	MatrixAccount   *string `form:"matrixAccount"` // Invalid: mock does not partition by matrix account.
	CardID          *string `form:"cardId"`
	CardType        *string `form:"cardType"`       // Invalid: transaction filtering by card type is unsupported.
	CardFormFactor  *string `form:"cardFormFactor"` // Invalid: transaction filtering by form factor is unsupported.
	RequestID       *string `form:"requestId"`
	TransactionID   *string `form:"transactionId"`
	TransactionType *string `form:"transactionType"`
	Status          *string `form:"status"`
	CreatedAtStart  *string `form:"createdAtStart"`
	CreatedAtEnd    *string `form:"createdAtEnd"`
}
type TradeData struct {
	TransactionID       string                   `json:"transactionId"`
	CreatedAt           timeTypes.ISODateTime    `json:"createdAt"`
	TxnDate             timeTypes.ISODateTime    `json:"txnDate"`
	TransactionType     photon.TransactionType   `json:"transactionType"`
	CardID              string                   `json:"cardId"`
	RequestID           string                   `json:"requestId"`
	TransactionAmount   float64                  `json:"transactionAmount"`
	TransactionCurrency common.Currency          `json:"transactionCurrency"`
	MerchantName        string                   `json:"merchantNameLocation"`
	Status              photon.TransactionStatus `json:"status"`
}

func (s *PhotonPayOpenAPIService) ListTrades(ctx context.Context, req *ListTradeRequest) (*[]TradeData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromOptionalString(req.CardID)
	if err != nil {
		return nil, err
	}
	transactionID, err := idconv.FromOptionalString(req.TransactionID)
	if err != nil {
		return nil, err
	}
	var typesFilter []common.CardTransactionType
	if req.TransactionType != nil {
		var valid bool
		typesFilter, valid = photon.ConvertStringToGenericTransactionTypes(*req.TransactionType)
		if !valid {
			return nil, photonpayerrors.ErrInvalidOperation
		}
	}
	var statuses []common.CardTransactionStatus
	if req.Status != nil {
		var valid bool
		statuses, valid = photon.ConvertStringToGenericTransactionStatuses(*req.Status)
		if !valid {
			return nil, photonpayerrors.ErrInvalidOperation
		}
	}
	createdFrom, valid := queryconv.OptionalTime(req.CreatedAtStart)
	if !valid {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	createdTo, valid := queryconv.OptionalTime(req.CreatedAtEnd)
	if !valid {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	if createdFrom != nil && createdTo != nil && createdFrom.After(*createdTo) {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	page, size := types.NormalizePagination(types.Value(req.PageIndex), types.Value(req.PageSize))
	transactions, err := s.usecase.ListTransactions(ctx, &biz.ListTransactionsRequest{
		AccountID:   &accountID,
		CardID:      cardID,
		ID:          transactionID,
		RequestID:   req.RequestID,
		Types:       typesFilter,
		Statuses:    statuses,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		Offset:      (page - 1) * size,
		Limit:       size,
	})
	if err != nil {
		return nil, err
	}
	items := make([]TradeData, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, TradeData{
			TransactionID:       idconv.ToString(transaction.ID),
			CreatedAt:           timeTypes.ISODateTime(transaction.CreatedAt.UTC()),
			TxnDate:             timeTypes.ISODateTime(transaction.CreatedAt.UTC()),
			TransactionType:     photon.ConvertGenericTransactionTypeToTransactionType(transaction.Type),
			CardID:              idconv.ToString(transaction.CardID),
			RequestID:           transaction.RequestID,
			TransactionAmount:   transaction.TxAmount.InexactFloat64(),
			TransactionCurrency: transaction.TxCurrency,
			MerchantName:        transaction.MerchantName,
			Status:              photon.ConvertGenericTransactionStatusToTransactionStatus(transaction.Status),
		})
	}
	return &items, nil
}

type UploadRequest struct {
	OpenAPIAccountRequest
	BusinessKey string                `uri:"businessKey" binding:"required"`
	File        *multipart.FileHeader `form:"file" binding:"required"`
}

func (s *PhotonPayOpenAPIService) Upload(_ context.Context, req *UploadRequest) (*string, error) {
	if req.File == nil {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	fileURL := "mock://photonpay/" + req.BusinessKey + "/" + req.File.Filename
	return &fileURL, nil
}

func cardHolderData(holder *model.CardHolder) *CardHolderData {
	return &CardHolderData{
		CardholderID:           idconv.ToString(holder.ID),
		MemberID:               photon.MemberID,
		Status:                 photon.ConvertGenericCardHolderStatusToCardHolderStatus(holder.Status),
		CardholderReviewStatus: photon.ConvertGenericCardHolderReviewStatusToCardHolderReviewStatus(holder.ReviewStatus),
		IdInfoRequirement:      "N",
	}
}

func cardData(card *model.Card) *CardData {
	return &CardData{
		CardID:         idconv.ToString(card.ID),
		CardNo:         card.CardNumber,
		CVV:            card.Cvv,
		ExpirationDate: timefmt.CardExpiration(card.ExpireAt),
		CardCurrency:   card.CardCurrency,
		CardScheme:     card.CardScheme,
		CardStatus:     photon.ConvertGenericCardStatusToCardStatus(card.Status),
		CardFormFactor: photon.ConvertGenericCardFormTypeToCardFormFactor(card.FormType),
		CardType:       photon.ConvertGenericCardTypeToCardType(card.CardType),
		CardholderID:   idconv.ToString(card.CardHolderID),
		CreatedAt:      timeTypes.ISODateTime(card.CreatedAt.UTC()),
		MaskCardNo:     maskCardNumber(card.CardNumber),
	}
}

func maskCardNumber(cardNumber string) string {
	if len(cardNumber) < 10 {
		return cardNumber
	}

	return cardNumber[:6] + photon.CardNumberMask + cardNumber[len(cardNumber)-4:]
}
