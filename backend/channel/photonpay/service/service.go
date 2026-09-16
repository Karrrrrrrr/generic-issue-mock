package service

import (
	"context"
	"mime/multipart"
	"time"

	"generic-mock/channel/photonpay/biz"
	photon "generic-mock/channel/photonpay/enums"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/timeparse"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"github.com/shopspring/decimal"
)

type PhotonPayOpenAPIService struct {
	usecase *biz.PhotonPayOpenAPIUsecase
}

func NewPhotonPayOpenAPIService(injector *do.Injector) (*PhotonPayOpenAPIService, error) {
	return &PhotonPayOpenAPIService{
		usecase: do.MustInvoke[*biz.PhotonPayOpenAPIUsecase](injector),
	}, nil
}

type AccessTokenRequest struct {
	GrantType    string `form:"grant_type" json:"grant_type" binding:"required"`
	ClientID     string `form:"client_id" json:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret" json:"client_secret" binding:"required"`
}

type AccessTokenData struct {
	ExpiresIn        int64  `json:"expiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	Token            string `json:"token"`
}

func (s *PhotonPayOpenAPIService) AccessToken(_ context.Context, _ *AccessTokenRequest) (*AccessTokenData, error) {
	return &AccessTokenData{
		ExpiresIn:        3600,
		RefreshExpiresIn: 7200,
		RefreshToken:     "photonpay-mock-refresh-token",
		Token:            "photonpay-mock-token",
	}, nil
}

type AccountSingleRequest struct {
	Currency      *string             `form:"currency" json:"currency"`
	AccountNo     *string             `form:"accountNo" json:"accountNo"` // Invalid: mock has one generic account.
	MemberID      *string             `form:"memberId" json:"memberId"`   // Invalid: mock does not partition by member.
	AccountType   *photon.AccountType `form:"accountType" json:"accountType"`
	MatrixAccount *string             `form:"matrixAccount" json:"matrixAccount"` // Invalid: Matrix is unsupported.
}

type AccountSingleData struct {
	MemberID        string             `json:"memberId"`
	AccountNo       string             `json:"accountNo"`
	AccountType     photon.AccountType `json:"accountType"`
	Currency        common.Currency    `json:"currency"`
	RealTimeBalance float64            `json:"realTimeBalance"`
	ReturnedAt      string             `json:"returnedAt"`
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
		ReturnedAt:      time.Now().UTC().Format(time.RFC3339),
	}, nil
}

type CreateCardHolderRequest struct {
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
	dateOfBirth, err := timeparse.ParseDate(req.DateOfBirth)
	if err != nil {
		return nil, biz.ErrInvalidDateOfBirth
	}

	holder, err := s.usecase.CreateCardHolder(ctx, &biz.CreateCardHolderRequest{
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
	CardholderID string  `json:"cardholderId" binding:"required"`
	Email        *string `json:"email"`
	Mobile       *string `json:"mobile"`
	MobilePrefix *string `json:"mobilePrefix"`
}

func (s *PhotonPayOpenAPIService) EditCardHolder(ctx context.Context, req *EditCardHolderRequest) (*CardHolderData, error) {
	holder, err := s.usecase.UpdateCardHolder(ctx, &biz.UpdateCardHolderRequest{
		CardholderID: req.CardholderID,
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
	PageIndex int `form:"pageIndex"`
	PageSize  int `form:"pageSize"`
}
type CardHolderListItem struct {
	CardholderID           string                        `json:"cardholderId"`
	CreatedAt              string                        `json:"createdAt"`
	FirstName              string                        `json:"firstName"`
	LastName               string                        `json:"lastName"`
	Email                  string                        `json:"email"`
	Mobile                 string                        `json:"mobile"`
	MobilePrefix           string                        `json:"mobilePrefix"`
	Status                 photon.CardHolderStatus       `json:"status"`
	CardholderReviewStatus photon.CardHolderReviewStatus `json:"cardholderReviewStatus"`
}

func (s *PhotonPayOpenAPIService) ListCardHolders(ctx context.Context, req *ListCardHolderRequest) (*[]CardHolderListItem, error) {
	page, size := types.NormalizePagination(req.PageIndex, req.PageSize)
	holders, err := s.usecase.ListCardHolders(ctx, &biz.ListRequest{
		Offset: (page - 1) * size,
		Limit:  size,
	})
	if err != nil {
		return nil, err
	}
	items := make([]CardHolderListItem, 0, len(holders))
	for _, holder := range holders {
		items = append(items, CardHolderListItem{
			CardholderID:           holder.ID,
			CreatedAt:              holder.CreatedAt.UTC().Format(time.RFC3339),
			FirstName:              holder.FirstName,
			LastName:               holder.LastName,
			Email:                  holder.Email,
			Mobile:                 holder.Mobile,
			MobilePrefix:           holder.MobilePrefix,
			Status:                 photon.CardHolderStatusFromGeneric(holder.Status),
			CardholderReviewStatus: photon.CardHolderReviewStatusFromGeneric(holder.ReviewStatus),
		})
	}
	return &items, nil
}

type CardBinRequest struct {
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

func (s *PhotonPayOpenAPIService) CardBins(ctx context.Context, _ *CardBinRequest) (*[]CardBinData, error) {
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
	CreatedAt      string                `json:"createdAt"`
	MaskCardNo     string                `json:"maskCardNo"`
}
type OpenCardData struct {
	CardDetail *CardData              `json:"cardDetail"`
	RequestID  string                 `json:"requestId"`
	Status     photon.OperationStatus `json:"status"`
}

func (s *PhotonPayOpenAPIService) OpenCard(ctx context.Context, req *OpenCardRequest) (*OpenCardData, error) {
	formFactor := req.CardFormFactor
	if formFactor == "" {
		formFactor = photon.CardFormFactor_Virtual
	}
	card, err := s.usecase.OpenCard(ctx, &biz.OpenCardRequest{
		CardBin:          req.CardBin,
		Currency:         req.CardCurrency,
		CardScheme:       req.CardScheme,
		CardType:         req.CardType,
		CardFormFactor:   formFactor,
		CardholderID:     req.CardholderID,
		RequestID:        req.RequestID,
		ExpirationMonths: types.Value(req.CardExpirationDate),
	})
	if err != nil {
		return nil, err
	}

	return &OpenCardData{
		CardDetail: cardData(card),
		RequestID:  card.RequestID,
		Status:     photon.OperationStatus_Succeed,
	}, nil
}

type CardIDRequest struct {
	CardID string `form:"cardId" json:"cardId" binding:"required"`
}

func (s *PhotonPayOpenAPIService) CardDetail(ctx context.Context, req *CardIDRequest) (*CardData, error) {
	card, err := s.usecase.GetCard(ctx, req.CardID)
	if err != nil {
		return nil, err
	}

	return cardData(card), nil
}

func (s *PhotonPayOpenAPIService) CardCVV(ctx context.Context, req *CardIDRequest) (*CardData, error) {
	card, err := s.usecase.GetCard(ctx, req.CardID)
	if err != nil {
		return nil, err
	}

	return cardData(card), nil
}

type ListCardsRequest struct {
	PageIndex  int                `form:"pageIndex"`
	PageSize   int                `form:"pageSize"`
	CardBin    *string            `form:"cardBin"`    // Invalid: card BIN filtering is unsupported.
	CardType   *photon.CardType   `form:"cardType"`   // Invalid: card type filtering is unsupported.
	CardStatus *photon.CardStatus `form:"cardStatus"` // Invalid: card status filtering is unsupported.
}

func (s *PhotonPayOpenAPIService) ListCards(ctx context.Context, req *ListCardsRequest) (*[]*CardData, error) {
	page, size := types.NormalizePagination(req.PageIndex, req.PageSize)
	cards, err := s.usecase.ListCards(ctx, &biz.ListRequest{
		Offset: (page - 1) * size,
		Limit:  size,
	})
	if err != nil {
		return nil, err
	}

	items := types.BulkConvertSlice(cards, cardData)

	return &items, nil
}

type UpdateCardRequest struct {
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
	card, err := s.usecase.UpdateCard(ctx, &biz.UpdateCardRequest{
		CardID:    req.CardID,
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
	RequestID string                   `form:"requestId" json:"requestId" binding:"required"`
	Type      photon.RequestResultType `form:"type" json:"type"`
}

func (s *PhotonPayOpenAPIService) RequestResult(ctx context.Context, req *RequestResultRequest) (*OpenCardData, error) {
	card, err := s.usecase.GetRequestResult(ctx, req.RequestID)
	if err != nil {
		return nil, err
	}

	return &OpenCardData{
		CardDetail: cardData(card),
		RequestID:  card.RequestID,
		Status:     photon.OperationStatus_Succeed,
	}, nil
}

type ChangeCardStatusRequest struct {
	CardID    string              `json:"cardId" binding:"required"`
	RequestID string              `json:"requestId" binding:"required"`
	Status    photon.FreezeStatus `json:"status" binding:"required"`
}

func (s *PhotonPayOpenAPIService) FreezeCard(ctx context.Context, req *ChangeCardStatusRequest) (*CardData, error) {
	card, err := s.usecase.ChangeCardStatus(ctx, &biz.ChangeCardStatusRequest{
		CardID:    req.CardID,
		RequestID: req.RequestID,
		Status:    photon.FreezeStatusToGeneric(req.Status),
		Operation: common.OperationType_FreezeCard,
	})
	if err != nil {
		return nil, err
	}

	return cardData(card), nil
}

type CancelCardRequest struct {
	CardID string `json:"cardId" binding:"required"`
}

func (s *PhotonPayOpenAPIService) CancelCard(ctx context.Context, req *CancelCardRequest) (*CardData, error) {
	card, err := s.usecase.ChangeCardStatus(ctx, &biz.ChangeCardStatusRequest{
		CardID:    req.CardID,
		RequestID: "",
		Status:    common.CardStatus_Deleted,
		Operation: common.OperationType_CancelCard,
	})
	if err != nil {
		return nil, err
	}

	return cardData(card), nil
}

type ListTradeRequest struct {
	PageIndex int `form:"pageIndex"`
	PageSize  int `form:"pageSize"`
}
type TradeData struct {
	TransactionID       string                   `json:"transactionId"`
	CardID              string                   `json:"cardId"`
	RequestID           string                   `json:"requestId"`
	TransactionAmount   float64                  `json:"transactionAmount"`
	TransactionCurrency common.Currency          `json:"transactionCurrency"`
	MerchantName        string                   `json:"merchantName"`
	Status              photon.TransactionStatus `json:"status"`
}

func (s *PhotonPayOpenAPIService) ListTrades(ctx context.Context, req *ListTradeRequest) (*[]TradeData, error) {
	page, size := types.NormalizePagination(req.PageIndex, req.PageSize)
	transactions, err := s.usecase.ListTransactions(ctx, &biz.ListRequest{
		Offset: (page - 1) * size,
		Limit:  size,
	})
	if err != nil {
		return nil, err
	}
	items := make([]TradeData, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, TradeData{
			TransactionID:       transaction.ID,
			CardID:              transaction.CardID,
			RequestID:           transaction.RequestID,
			TransactionAmount:   transaction.TxAmount.InexactFloat64(),
			TransactionCurrency: transaction.TxCurrency,
			MerchantName:        transaction.MerchantName,
			Status:              photon.TransactionStatusFromGeneric(transaction.Status),
		})
	}
	return &items, nil
}

type UploadRequest struct {
	BusinessKey string                `uri:"businessKey" binding:"required"`
	File        *multipart.FileHeader `form:"file" binding:"required"`
}

type SandboxTransactionRequest struct {
	RequestID           string                        `json:"requestId" binding:"required"`
	CardID              string                        `json:"cardID" binding:"required"`
	Cvv                 string                        `json:"cvv" binding:"required"`            // Invalid: the mock does not verify card security codes.
	ExpirationDate      string                        `json:"expirationDate" binding:"required"` // Invalid: the mock does not verify card expiry.
	OriginTransactionID string                        `json:"originTransactionId"`
	TxnCurrency         common.Currency               `json:"txnCurrency" binding:"required"`
	TxnAmount           float64                       `json:"txnAmount" binding:"required"`
	TxnType             photon.SandboxTransactionType `json:"txnType" binding:"required"`
	Mcc                 string                        `json:"mcc" binding:"required"`
	MerchantName        string                        `json:"merchantName" binding:"required"`
	MerchantCountry     string                        `json:"merchantCountry" binding:"required"`
	MerchantCity        string                        `json:"merchantCity" binding:"required"`     // Invalid: merchant city is not persisted by the generic model.
	MerchantPostcode    string                        `json:"merchantPostcode" binding:"required"` // Invalid: merchant postcode is not persisted by the generic model.
}

type SandboxTransactionData struct{}

func (s *PhotonPayOpenAPIService) SandboxTransaction(ctx context.Context, req *SandboxTransactionRequest) (*SandboxTransactionData, error) {
	err := s.usecase.SandboxTransaction(ctx, &biz.SandboxTransactionRequest{
		RequestID:           req.RequestID,
		CardID:              req.CardID,
		OriginTransactionID: req.OriginTransactionID,
		Currency:            req.TxnCurrency,
		Amount:              decimal.NewFromFloat(req.TxnAmount),
		Type:                req.TxnType,
		MerchantName:        req.MerchantName,
		MerchantCountry:     req.MerchantCountry,
		MerchantMCC:         req.Mcc,
	})
	if err != nil {
		return nil, err
	}

	return &SandboxTransactionData{}, nil
}

type UploadData struct {
	FileURL string `json:"fileUrl"`
}

func (s *PhotonPayOpenAPIService) Upload(_ context.Context, req *UploadRequest) (*UploadData, error) {
	return &UploadData{
		FileURL: "mock://photonpay/" + req.BusinessKey + "/" + req.File.Filename,
	}, nil
}

func cardHolderData(holder *model.CardHolder) *CardHolderData {
	return &CardHolderData{
		CardholderID:           holder.ID,
		MemberID:               photon.MemberID,
		Status:                 photon.CardHolderStatusFromGeneric(holder.Status),
		CardholderReviewStatus: photon.CardHolderReviewStatusFromGeneric(holder.ReviewStatus),
		IdInfoRequirement:      "N",
	}
}

func cardData(card *model.Card) *CardData {
	return &CardData{
		CardID:         card.ID,
		CardNo:         card.CardNumber,
		CVV:            card.Cvv,
		ExpirationDate: card.ExpireAt.Format("01/06"),
		CardCurrency:   card.CardCurrency,
		CardScheme:     card.CardScheme,
		CardStatus:     photon.CardStatusFromGeneric(card.Status),
		CardFormFactor: photon.CardFormFactorFromGeneric(card.FormType),
		CardType:       photon.CardTypeFromGeneric(card.CardType),
		CardholderID:   card.CardHolderID,
		CreatedAt:      card.CreatedAt.UTC().Format(time.RFC3339),
		MaskCardNo:     maskCardNumber(card.CardNumber),
	}
}

func maskCardNumber(cardNumber string) string {
	if len(cardNumber) < 10 {
		return cardNumber
	}

	return cardNumber[:6] + photon.CardNumberMask + cardNumber[len(cardNumber)-4:]
}
