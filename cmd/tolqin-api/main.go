package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator"
	"github.com/ztimes2/tolqin/internal/config"
	"github.com/ztimes2/tolqin/internal/logging"
	"github.com/ztimes2/tolqin/internal/postgres"
	"github.com/ztimes2/tolqin/internal/router"
	"github.com/ztimes2/tolqin/internal/surfing"
)

func main() {
	conf, err := config.LoadAPI()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := logging.NewLogger(conf.LogLevel, conf.LogFormat)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	db, err := postgres.NewDB(postgres.Config{
		Host:         conf.Database.Host,
		Port:         conf.Database.Port,
		Username:     conf.Database.Username,
		Password:     conf.Database.Password,
		DatabaseName: conf.Database.Name,
		SSLMode:      postgres.NewSSLMode(conf.Database.SSLMode),
	})
	if err != nil {
		logger.WithError(err).Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	spotStore := postgres.NewSpotStore(db)
	validate := validator.New()
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
			logger.WithError(err).Fatalf("server failed to listen: %v", err)
		}
	}()

	stopCh := make(chan os.Signal, 2)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT)
	<-stopCh
	logger.Info("shutting down server")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.WithError(err).Errorf("failed to gracefully shut down server: %v", err)
	}
}
