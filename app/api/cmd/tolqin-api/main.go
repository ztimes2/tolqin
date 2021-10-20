package main

import (
	"log"

	_ "github.com/lib/pq"
	config "github.com/ztimes2/tolqin/app/api/internal/api/config"
	"github.com/ztimes2/tolqin/app/api/internal/api/router"
	serviceauth "github.com/ztimes2/tolqin/app/api/internal/api/service/auth"
	"github.com/ztimes2/tolqin/app/api/internal/api/service/management"
	"github.com/ztimes2/tolqin/app/api/internal/api/service/surfer"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/auth"
	authpsql "github.com/ztimes2/tolqin/app/api/internal/pkg/auth/psql"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/geo/nominatim"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/jwt"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf/psql"
	"github.com/ztimes2/tolqin/app/api/pkg/httpserver"
	logx "github.com/ztimes2/tolqin/app/api/pkg/log"
	"github.com/ztimes2/tolqin/app/api/pkg/psqlutil"
)

func main() {
	conf, err := config.Load()
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

	jwtEncodeDecoder := jwt.NewEncodeDecoder(conf.JWTSigningKey, conf.JWTExpiry)

	router := router.New(
		serviceauth.NewService(
			auth.NewPasswordSalter(),
			auth.NewPasswordHasher(),
			jwtEncodeDecoder,
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
		jwtEncodeDecoder,
		logger,
	)

	server := httpserver.New(conf.ServerPort, router, httpserver.WithLogger(logger))
	if err := server.ListenAndServe(); err != nil {
		logger.WithError(err).Fatalf("server failure: %v", err)
	}
}
