package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	configapi "github.com/ztimes2/tolqin/internal/config/api"
	"github.com/ztimes2/tolqin/internal/geo/nominatim"
	"github.com/ztimes2/tolqin/internal/pkg/logging"
	"github.com/ztimes2/tolqin/internal/pkg/psqlutil"
	"github.com/ztimes2/tolqin/internal/router"
	"github.com/ztimes2/tolqin/internal/service/management"
	managementpsql "github.com/ztimes2/tolqin/internal/service/management/psql"
	"github.com/ztimes2/tolqin/internal/service/surfing"
	surfingpsql "github.com/ztimes2/tolqin/internal/service/surfing/psql"
)

func main() {
	conf, err := configapi.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := logging.New(conf.LogLevel, conf.LogFormat)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	db, err := psqlutil.NewDB(psqlutil.Config{
		Host:         conf.Database.Host,
		Port:         conf.Database.Port,
		Username:     conf.Database.Username,
		Password:     conf.Database.Password,
		DatabaseName: conf.Database.Name,
		SSLMode:      psqlutil.NewSSLMode(conf.Database.SSLMode),
	})
	if err != nil {
		logger.WithError(err).Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	router := router.New(
		surfing.NewService(surfingpsql.NewSpotStore(db)),
		management.NewService(
			managementpsql.NewSpotStore(db),
			nominatim.New(nominatim.Config{
				BaseURL: conf.Nominatim.BaseURL,
				Timeout: conf.Nominatim.Timeout,
			}),
		),
		logger,
	)

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
