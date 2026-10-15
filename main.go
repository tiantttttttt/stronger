package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"stronger/server/service/api"
	"stronger/server/service/config"
	"syscall"
)

// @title Stronger API
// @version 1.0
// @description 居家力量训练记录 API。
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 填写 Bearer <access_token>。
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	api, err := api.New(cfg)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return api.Run(ctx)
}
