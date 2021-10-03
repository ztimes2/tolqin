package main

import (
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/ztimes2/tolqin/app/api/internal/auth"
	authpsql "github.com/ztimes2/tolqin/app/api/internal/auth/psql"
	config "github.com/ztimes2/tolqin/app/api/internal/config/api"
	"github.com/ztimes2/tolqin/app/api/internal/geo/nominatim"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/httpserver"
	logx "github.com/ztimes2/tolqin/app/api/internal/pkg/log"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/psqlutil"
	"github.com/ztimes2/tolqin/app/api/internal/router"
	serviceauth "github.com/ztimes2/tolqin/app/api/internal/service/auth"
	"github.com/ztimes2/tolqin/app/api/internal/service/management"
	"github.com/ztimes2/tolqin/app/api/internal/service/surfer"
	"github.com/ztimes2/tolqin/app/api/internal/surf/psql"
)

func main() {
	conf, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := logx.New(conf.LogLevel, conf.LogFormat)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	db, err := psqlutil.NewDB(psqlutil.DriverNamePQ, psqlutil.Config{
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

	spotStore := psql.NewSpotStore(db)

	router := router.New(
		serviceauth.NewService(
			auth.NewPasswordSalter(),
			auth.NewPasswordHasher(),
			auth.NewTokener("tolqin", "123456", 10*time.Minute),
			authpsql.NewUserStore(db),
		),
		surfer.NewService(spotStore),
		management.NewService(
			spotStore,
			nominatim.New(nominatim.Config{
				BaseURL: conf.Nominatim.BaseURL,
				Timeout: conf.Nominatim.Timeout,
			}),
		),
		logger,
	)

	server := httpserver.New(conf.ServerPort, router, httpserver.WithLogger(logger))
	if err := server.ListenAndServe(); err != nil {
		logger.WithError(err).Fatalf("server failure: %v", err)
	}
}
