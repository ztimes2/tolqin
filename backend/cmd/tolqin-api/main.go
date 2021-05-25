package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator"
	"github.com/ztimes2/tolqin/backend/internal/config"
	"github.com/ztimes2/tolqin/backend/internal/logging"
	"github.com/ztimes2/tolqin/backend/internal/router"
	"github.com/ztimes2/tolqin/backend/internal/surfing"
	"github.com/ztimes2/tolqin/backend/internal/surfing/inmemory"
)

func main() {
	conf, err := config.Load()
	fatalOnError(err, "failed to load config")

	logger, err := logging.NewLogger(conf.LogLevel, conf.LogFormat)
	fatalOnError(err, "failed to initialize logger")

	validate := validator.New()
	spotStore := inmemory.NewSpotStore()
	service := surfing.NewService(validate, spotStore)
	router := router.New(service, logger)

	server := &http.Server{
		Addr:    ":" + conf.ServerPort,
		Handler: router,
		// TODO configure timeouts
	}

	go func() {
		logger.Infof("server listening on port %s", conf.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server failed to listen: %v", err)
		}
	}()

	stopCh := make(chan os.Signal, 2)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT)
	<-stopCh
	logger.Info("shutting down server")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.
			WithError(err).
			Errorf("failed to gracefully shut down server: %v", err)
	}

}

func fatalOnError(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
