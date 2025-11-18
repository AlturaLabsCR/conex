package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"app/config"
	"app/handlers"
	"app/i18n"
	"app/middleware"
	"app/router"

	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//go:embed assets/*
var assetsFS embed.FS

func main() {
	config.Init()
	ctx := context.Background()

	logger, err := config.InitLogger()
	if err != nil {
		print("failed logger initialization: %v\n", err)
		os.Exit(1)
	}

	database, err := config.InitDB(ctx)
	if err != nil {
		print("failed database initialization: %v\n", err)
		os.Exit(1)
	}
	defer database.Close(ctx)

	locales := map[string]map[string]string{
		"es": i18n.ES,
		"en": i18n.EN,
	}

	smtpAuth := config.InitSMTPAuth()

	s3c, err := s3config.LoadDefaultConfig(context.TODO())
	if err != nil {
		print("failed s3 initialization: %v\n", err)
		os.Exit(1)
	}

	s3client := s3.NewFromConfig(s3c)

	handler := handlers.New(
		handlers.HandlerParams{
			Production:   config.Production,
			Logger:       logger,
			Database:     database,
			Storage:      s3client,
			Locales:      locales,
			SMTPAuth:     smtpAuth,
			ServerSecret: config.ServerSecret,
			CookieName:   config.CookieName,
			CookiePath:   config.Endpoints[config.RootPath],
		},
	)

	routes := router.Routes(handler)

	routes.Handle(
		"GET "+config.Endpoints[config.AssetsPath],
		middleware.DisableCacheInDevMode(
			config.Production,
			handlers.Gzip(http.FileServer(http.FS(assetsFS))),
		),
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server starting", "address", ":"+config.Port)

		if err := http.ListenAndServe(":"+config.Port, routes); err != nil {
			logger.Error("failed to start server", "port", config.Port, "error", err)
			os.Exit(1)
		}
	}()

	<-stop

	logger.Info("shutting down...")
}
