package main

import (
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	config "github.com/ztimes2/tolqin/app/api/internal/config/importer"
	"github.com/ztimes2/tolqin/app/api/internal/importing"
	"github.com/ztimes2/tolqin/app/api/internal/importing/csv"
	"github.com/ztimes2/tolqin/app/api/internal/surf/psql"
	logx "github.com/ztimes2/tolqin/app/api/pkg/log"
	"github.com/ztimes2/tolqin/app/api/pkg/psqlutil"
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

	file, err := os.Open(conf.CSVFile)
	if err != nil {
		logger.WithError(err).Fatalf("failed to read csv file: %v", err)
	}
	defer file.Close()

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

	start := time.Now()

	count, err := importing.ImportSpots(
		csv.NewSpotEntrySource(file),
		psql.NewSpotStore(db, psql.WithBatchSize(conf.BatchSize)),
	)
	if err != nil {
		logger.WithError(err).Fatalf("failed to import spots: %v", err)
	}

	duration := time.Since(start)

	logger.Infof("%d spot(s) were imported in %v", count, duration)
}
