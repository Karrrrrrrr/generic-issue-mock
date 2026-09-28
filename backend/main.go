package main

import (
	"context"
	"os"

	"generic-mock/channel/paynda"
	payndaHTTP "generic-mock/channel/paynda/http"
	payndaService "generic-mock/channel/paynda/service"
	"generic-mock/channel/photonpay"
	photonHTTP "generic-mock/channel/photonpay/http"
	photonService "generic-mock/channel/photonpay/service"
	"generic-mock/channel/pingpong"
	pingHTTP "generic-mock/channel/pingpong/http"
	pingService "generic-mock/channel/pingpong/service"
	"generic-mock/channel/slash"
	slashHTTP "generic-mock/channel/slash/http"
	slashService "generic-mock/channel/slash/service"
	"generic-mock/data"
	"generic-mock/shared"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()
	zap.ReplaceGlobals(logger)

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = data.DefaultPostgresDSN
	}

	db, err := data.NewPostgresDB(dsn)
	if err != nil {
		zap.S().Fatalw("connect database", "error", err)
	}
	if err := data.SeedInitialData(context.Background(), db); err != nil {
		zap.S().Fatalw("seed initial data", "error", err)
	}

	injector := do.New()
	do.ProvideValue(injector, db)
	shared.RegisterProviders(injector)
	shared.RegisterUIProviders(injector)
	photonpay.RegisterProviders(injector)
	paynda.RegisterProviders(injector)
	slash.RegisterProviders(injector)
	pingpong.RegisterProviders(injector)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	pingHTTP.Register(pingHTTP.RegisterRequest{
		Router:  router.Group("/pingpong"),
		OpenAPI: do.MustInvoke[*pingService.PingPongOpenAPIService](injector),
		UI:      do.MustInvoke[*pingService.PingPongUIService](injector),
	})
	photonHTTP.Register(photonHTTP.RegisterRequest{
		Router:  router.Group("/photonpay"),
		OpenAPI: do.MustInvoke[*photonService.PhotonPayOpenAPIService](injector),
		UI:      do.MustInvoke[*photonService.PhotonPayUIService](injector),
	})
	payndaHTTP.Register(payndaHTTP.RegisterRequest{
		Router:  router.Group("/paynda"),
		OpenAPI: do.MustInvoke[*payndaService.PayndaOpenAPIService](injector),
		UI:      do.MustInvoke[*payndaService.PayndaUIService](injector),
	})
	slashHTTP.Register(slashHTTP.RegisterRequest{
		Router:  router.Group("/slash"),
		OpenAPI: do.MustInvoke[*slashService.SlashOpenAPIService](injector),
		UI:      do.MustInvoke[*slashService.SlashUIService](injector),
	})
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8000"
	}

	if err := data.Ping(context.Background(), db); err != nil {
		zap.S().Fatalw("ping database", "error", err)
	}

	if err := router.Run(addr); err != nil {
		zap.S().Fatalw("run http server", "error", err)
	}
}
