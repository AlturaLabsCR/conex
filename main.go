package main

import (
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
)

//go:embed assets/*
var assetsFS embed.FS

func main() {
	config.Init()

	logger, err := config.InitLogger()
	if err != nil {
		print("failed logger initialization: %v\n", err)
		os.Exit(1)
	}

	database, err := config.InitDB()
	if err != nil {
		print("failed database initialization: %v\n", err)
		os.Exit(1)
	}

	locales := map[string]map[string]string{
		"es": i18n.ES,
		"en": i18n.EN,
	}

	handler := handlers.New(
		handlers.HandlerParams{
			Production: config.Production,
			DB:         database,
			Logger:     logger,
		},
		i18n.New(locales).TranslateHTTPRequest,
	)

	routes := router.Routes(handler)

	routes.Handle(
		"GET /assets/",
		middleware.DisableCacheInDevMode(
			config.Production,
			http.FileServer(http.FS(assetsFS)),
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
