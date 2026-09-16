package main

import (
	"context"
	"log"
	"os"

	"generic-mock/channel/photonpay/biz"
	photonData "generic-mock/channel/photonpay/data"
	photonHTTP "generic-mock/channel/photonpay/http"
	photonService "generic-mock/channel/photonpay/service"
	"generic-mock/data"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = data.DefaultPostgresDSN
	}

	db, err := data.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	injector := do.New()
	do.ProvideValue(injector, db)
	do.Provide(injector, func(i *do.Injector) (*photonData.Repository, error) {
		return photonData.NewRepository(do.MustInvoke[*gorm.DB](i)), nil
	})
	do.Provide(injector, func(i *do.Injector) (*biz.Usecase, error) {
		repository := do.MustInvoke[*photonData.Repository](i)
		return biz.NewUsecase(repository, repository, repository), nil
	})
	do.Provide(injector, func(i *do.Injector) (*photonService.Service, error) {
		return photonService.NewService(do.MustInvoke[*biz.Usecase](i)), nil
	})

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	photonHTTP.Register(router.Group("/photonpay"), do.MustInvoke[*photonService.Service](injector))
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8000"
	}

	if err := data.Ping(context.Background(), db); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	if err := router.Run(addr); err != nil {
		log.Fatalf("run http server: %v", err)
	}
}
