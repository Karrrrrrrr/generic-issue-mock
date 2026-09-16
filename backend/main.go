package main

import (
	"context"
	"os"

	"generic-mock/channel/photonpay/biz"
	photonData "generic-mock/channel/photonpay/data"
	photonHTTP "generic-mock/channel/photonpay/http"
	photonService "generic-mock/channel/photonpay/service"
	"generic-mock/data"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
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

	injector := do.New()
	do.ProvideValue(injector, db)
	do.Provide(injector, photonData.NewRepository)
	do.Provide(injector, photonData.NewTransaction)
	do.Provide(injector, photonData.NewCardHolderRepository)
	do.Provide(injector, photonData.NewCardRepository)
	do.Provide(injector, photonData.NewCardTransactionRepository)
	do.Provide(injector, biz.NewUsecase)
	do.Provide(injector, photonService.NewService)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	photonHTTP.Register(router.Group("/photonpay"), do.MustInvoke[*photonService.Service](injector))
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
