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

type OpenAPIAccountRequest struct {
	Token string `header:"X-PD-TOKEN" binding:"required"`
}

func (s *PhotonPayOpenAPIService) accountID(req *OpenAPIAccountRequest) (model.ID, error) {
	return photonPayAccountID(req.Token)
}

func NewPhotonPayOpenAPIService(injector *do.Injector) (*PhotonPayOpenAPIService, error) {
	return &PhotonPayOpenAPIService{
		usecase: do.MustInvoke[*biz.PhotonPayOpenAPIUsecase](injector),
	}, nil
}

type AccessTokenRequest struct {
	AppID  string `form:"app_id" json:"app_id" binding:"required"`
	Secret string `form:"secret" json:"secret"`
}

type AccessTokenData struct {
	ExpiresIn        int64  `json:"expiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	Token            string `json:"token"`
}

func (s *PhotonPayOpenAPIService) AccessToken(_ context.Context, req *AccessTokenRequest) (*AccessTokenData, error) {
	if req.Secret != "" {
		return nil, biz.ErrInvalidOperation
	}
	accountID, err := photonPayAccountID(req.AppID)
	if err != nil {
		return nil, err
	}
	token := photonPayIDString(accountID)
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
	MemberID        string             `json:"memberId"`
	AccountNo       string             `json:"accountNo"`
	AccountType     photon.AccountType `json:"accountType"`
	Currency        common.Currency    `json:"currency"`
	RealTimeBalance float64            `json:"realTimeBalance"`
	ReturnedAt      time.Time          `json:"returnedAt"`
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
		ReturnedAt:      time.Now().UTC(),
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
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	dateOfBirth, err := timeparse.ParseDate(req.DateOfBirth)
	if err != nil {
		return nil, biz.ErrInvalidDateOfBirth
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
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	cardholderID, err := photonPayID(req.CardholderID)
	if err != nil {
		return nil, err
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
	PageIndex int `form:"pageIndex"`
	PageSize  int `form:"pageSize"`
}
type CardHolderListItem struct {
	CardholderID           string                        `json:"cardholderId"`
	CreatedAt              time.Time                     `json:"createdAt"`
	FirstName              string                        `json:"firstName"`
	LastName               string                        `json:"lastName"`
	Email                  string                        `json:"email"`
	Mobile                 string                        `json:"mobile"`
	MobilePrefix           string                        `json:"mobilePrefix"`
	Status                 photon.CardHolderStatus       `json:"status"`
	CardholderReviewStatus photon.CardHolderReviewStatus `json:"cardholderReviewStatus"`
}

func (s *PhotonPayOpenAPIService) ListCardHolders(ctx context.Context, req *ListCardHolderRequest) (*[]CardHolderListItem, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	page, size := types.NormalizePagination(req.PageIndex, req.PageSize)
	holders, err := s.usecase.ListCardHolders(ctx, &biz.ListRequest{
		AccountID: accountID,
		Offset:    (page - 1) * size,
		Limit:     size,
	})
	if err != nil {
		return nil, err
	}
	items := make([]CardHolderListItem, 0, len(holders))
	for _, holder := range holders {
		items = append(items, CardHolderListItem{
			CardholderID:           photonPayIDString(holder.ID),
			CreatedAt:              holder.CreatedAt.UTC(),
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
	OpenAPIAccountRequest
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
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	products, err := s.usecase.ListCardProducts(ctx, accountID)
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
	CreatedAt      time.Time             `json:"createdAt"`
	MaskCardNo     string                `json:"maskCardNo"`
}
type OpenCardData struct {
	CardDetail *CardData              `json:"cardDetail"`
	RequestID  string                 `json:"requestId"`
	Status     photon.OperationStatus `json:"status"`
}

func (s *PhotonPayOpenAPIService) OpenCard(ctx context.Context, req *OpenCardRequest) (*OpenCardData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	cardholderID, err := photonPayID(req.CardholderID)
	if err != nil {
		return nil, err
	}
	formFactor := req.CardFormFactor
	if formFactor == "" {
		formFactor = photon.CardFormFactor_Virtual
	}
	card, err := s.usecase.OpenCard(ctx, &biz.OpenCardRequest{
		AccountID:        accountID,
		CardBin:          req.CardBin,
		Currency:         req.CardCurrency,
		CardScheme:       req.CardScheme,
		CardType:         req.CardType,
		CardFormFactor:   formFactor,
		CardholderID:     cardholderID,
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
	OpenAPIAccountRequest
	CardID string `form:"cardId" json:"cardId" binding:"required"`
}

func (s *PhotonPayOpenAPIService) CardDetail(ctx context.Context, req *CardIDRequest) (*CardData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := photonPayID(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.GetCard(ctx, &biz.ResourceRequest{AccountID: &accountID, ID: cardID})
	if err != nil {
		return nil, err
	}

	return cardData(card), nil
}

func (s *PhotonPayOpenAPIService) CardCVV(ctx context.Context, req *CardIDRequest) (*CardData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := photonPayID(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.GetCard(ctx, &biz.ResourceRequest{AccountID: &accountID, ID: cardID})
	if err != nil {
		return nil, err
	}

	return cardData(card), nil
}

type ListCardsRequest struct {
	OpenAPIAccountRequest
	PageIndex  int                `form:"pageIndex"`
	PageSize   int                `form:"pageSize"`
	CardBin    *string            `form:"cardBin"`    // Invalid: card BIN filtering is unsupported.
	CardType   *photon.CardType   `form:"cardType"`   // Invalid: card type filtering is unsupported.
	CardStatus *photon.CardStatus `form:"cardStatus"` // Invalid: card status filtering is unsupported.
}

func (s *PhotonPayOpenAPIService) ListCards(ctx context.Context, req *ListCardsRequest) (*[]*CardData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	page, size := types.NormalizePagination(req.PageIndex, req.PageSize)
	cards, err := s.usecase.ListCards(ctx, &biz.ListRequest{
		AccountID: accountID,
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
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := photonPayID(req.CardID)
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
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.GetRequestResult(ctx, &biz.RequestResultResourceRequest{AccountID: accountID, RequestID: req.RequestID})
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
	OpenAPIAccountRequest
	CardID    string              `json:"cardId" binding:"required"`
	RequestID string              `json:"requestId" binding:"required"`
	Status    photon.FreezeStatus `json:"status" binding:"required"`
}

func (s *PhotonPayOpenAPIService) FreezeCard(ctx context.Context, req *ChangeCardStatusRequest) (*CardData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := photonPayID(req.CardID)
	if err != nil {
		return nil, err
	}
	card, err := s.usecase.ChangeCardStatus(ctx, &biz.ChangeCardStatusRequest{
		AccountID: accountID,
		CardID:    cardID,
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
	OpenAPIAccountRequest
	CardID string `json:"cardId" binding:"required"`
}

func (s *PhotonPayOpenAPIService) CancelCard(ctx context.Context, req *CancelCardRequest) (*CardData, error) {
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := photonPayID(req.CardID)
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

	return cardData(card), nil
}

type ListTradeRequest struct {
	OpenAPIAccountRequest
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
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	page, size := types.NormalizePagination(req.PageIndex, req.PageSize)
	transactions, err := s.usecase.ListTransactions(ctx, &biz.ListRequest{
		AccountID: accountID,
		Offset:    (page - 1) * size,
		Limit:     size,
	})
	if err != nil {
		return nil, err
	}
	items := make([]TradeData, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, TradeData{
			TransactionID:       photonPayIDString(transaction.ID),
			CardID:              photonPayIDString(transaction.CardID),
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
	OpenAPIAccountRequest
	BusinessKey string                `uri:"businessKey" binding:"required"`
	File        *multipart.FileHeader `form:"file" binding:"required"`
}

type SandboxTransactionRequest struct {
	OpenAPIAccountRequest
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
	accountID, err := s.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	cardID, err := photonPayID(req.CardID)
	if err != nil {
		return nil, err
	}
	var originTransactionID model.ID
	if req.OriginTransactionID != "" {
		originTransactionID, err = photonPayID(req.OriginTransactionID)
		if err != nil {
			return nil, err
		}
	}
	err = s.usecase.SandboxTransaction(ctx, &biz.SandboxTransactionRequest{
		AccountID:           accountID,
		RequestID:           req.RequestID,
		CardID:              cardID,
		OriginTransactionID: originTransactionID,
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
		CardholderID:           photonPayIDString(holder.ID),
		MemberID:               photon.MemberID,
		Status:                 photon.CardHolderStatusFromGeneric(holder.Status),
		CardholderReviewStatus: photon.CardHolderReviewStatusFromGeneric(holder.ReviewStatus),
		IdInfoRequirement:      "N",
	}
}

func cardData(card *model.Card) *CardData {
	return &CardData{
		CardID:         photonPayIDString(card.ID),
		CardNo:         card.CardNumber,
		CVV:            card.Cvv,
		ExpirationDate: card.ExpireAt.Format("01/06"),
		CardCurrency:   card.CardCurrency,
		CardScheme:     card.CardScheme,
		CardStatus:     photon.CardStatusFromGeneric(card.Status),
		CardFormFactor: photon.CardFormFactorFromGeneric(card.FormType),
		CardType:       photon.CardTypeFromGeneric(card.CardType),
		CardholderID:   photonPayIDString(card.CardHolderID),
		CreatedAt:      card.CreatedAt.UTC(),
		MaskCardNo:     maskCardNumber(card.CardNumber),
	}
}

func maskCardNumber(cardNumber string) string {
	if len(cardNumber) < 10 {
		return cardNumber
	}

	return cardNumber[:6] + photon.CardNumberMask + cardNumber[len(cardNumber)-4:]
}
