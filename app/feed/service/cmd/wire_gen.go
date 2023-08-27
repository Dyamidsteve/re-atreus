// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/toomanysource/atreus/app/feed/service/internal/biz"
	"github.com/toomanysource/atreus/app/feed/service/internal/conf"
	"github.com/toomanysource/atreus/app/feed/service/internal/data"
	"github.com/toomanysource/atreus/app/feed/service/internal/server"
	"github.com/toomanysource/atreus/app/feed/service/internal/service"

	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, client *conf.Client, jwt *conf.JWT, logger log.Logger) (*kratos.App, func(), error) {
	publishConn := server.NewPublishClient(client, logger)
	feedRepo := data.NewFeedRepo(publishConn, logger)
	feedUsecase := biz.NewFeedUsecase(feedRepo, jwt, logger)
	feedService := service.NewFeedService(feedUsecase, logger)
	grpcServer := server.NewGRPCServer(confServer, feedService, logger)
	httpServer := server.NewHTTPServer(confServer, jwt, feedService, logger)
	app := newApp(logger, grpcServer, httpServer)
	return app, func() {
	}, nil
}