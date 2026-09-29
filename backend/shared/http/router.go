package http

import (
	"net/http"

	"generic-mock/pkg/httpx"
	sharederrors "generic-mock/shared/errors"
	"generic-mock/shared/service"

	"github.com/gin-gonic/gin"
	kratoserrors "github.com/go-kratos/kratos/v2/errors"
)

type RegisterRequest struct {
	Router                *gin.RouterGroup
	Service               *service.Service
	EnableVirtualAccounts bool
	EnableWebhooks        bool
}

func Register(req RegisterRequest) {
	router := req.Router
	management := req.Service
	router.GET("/accounts", bind(management.ListAccounts))
	router.POST("/accounts", bind(management.CreateAccount))
	router.POST("/accounts/rename", bind(management.RenameAccount))
	router.POST("/accounts/adjust", bind(management.AdjustAccount))
	router.GET("/card-products", bind(management.ListCardProducts))
	router.GET("/cardholders", bind(management.ListCardHolders))
	router.GET("/wallets", bind(management.ListWallets))
	router.GET("/cards", bind(management.ListCards))
	router.GET("/cards/detail", bind(management.GetCard))
	router.POST("/cards/status", bind(management.UpdateCardStatus))
	router.POST("/cards/fund", bind(management.FundCard))
	router.GET("/authorizations", bind(management.ListAuthorizations))
	router.GET("/authorizations/detail", bind(management.GetAuthorizationDetail))
	router.GET("/transactions", bind(management.ListCardTransactions))
	router.GET("/transactions/detail", bind(management.GetCardTransaction))
	router.POST("/transactions/stages", bind(management.ApplyTransactionStage))
	router.GET("/transfers", bind(management.ListWalletTransfers))
	router.POST("/simulate/authorizations", bind(management.SimulateAuthorization))
	router.POST("/simulate/clearings", bind(management.SimulateClearing))
	router.POST("/simulate/refunds", bind(management.SimulateRefund))
	router.POST("/simulate/reversals", bind(management.SimulateReversal))
	if req.EnableVirtualAccounts {
		router.GET("/virtual-accounts", bind(management.ListVirtualAccounts))
		router.POST("/virtual-accounts", bind(management.CreateVirtualAccount))
		router.POST("/virtual-accounts/fund", bind(management.FundVirtualAccount))
	}
	if req.EnableWebhooks {
		router.GET("/webhooks", bind(management.ListWebhooks))
		router.POST("/webhooks", bind(management.CreateWebhook))
		router.POST("/webhooks/update", bind(management.UpdateWebhook))
		router.POST("/webhooks/delete", bind(management.DeleteWebhook))
		router.GET("/webhooks/events", bind(management.ListWebhookEvents))
		router.GET("/webhook-records", bind(management.ListWebhookRecords))
		router.GET("/webhook-records/detail", bind(management.GetWebhookRecord))
		router.POST("/webhook-records/replay", bind(management.ReplayWebhookRecord))
	}
}

type errorResponse struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func bind[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.BindUI(httpx.BindUIRequest[Req, Resp]{
		Service:             fn,
		SuccessEncoder:      func(data *Resp) any { return data },
		BindingErrorEncoder: bindingFailure,
		ErrorEncoder:        failure,
	})
}

func bindingFailure(error) (int, any) {
	return failure(sharederrors.ErrInvalidUIRequest)
}

func failure(err error) (int, any) {
	converted := kratoserrors.FromError(err)
	status := int(converted.Code)
	if status < http.StatusBadRequest || status > 599 {
		status = http.StatusInternalServerError
	}
	return status, errorResponse{
		Reason:  converted.Reason,
		Message: converted.Message,
	}
}
