package main

import (
	"log"
	"os"
	"time"

	"github.com/ztimes2/tolqin/internal/config"
	"github.com/ztimes2/tolqin/internal/importing"
	"github.com/ztimes2/tolqin/internal/importing/csv"
	"github.com/ztimes2/tolqin/internal/importing/psql"
	"github.com/ztimes2/tolqin/internal/logging"
	"github.com/ztimes2/tolqin/internal/psqlutil"
)

func main() {
	conf, err := config.LoadImporter()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := logging.NewLogger(conf.LogLevel, conf.LogFormat)
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

	file, err := os.Open(conf.CSVFile)
	if err != nil {
		logger.WithError(err).Fatalf("failed to read csv file: %v", err)
	}
	defer file.Close()

	start := time.Now()

	count, err := importing.ImportSpots(
		csv.NewSpotEntries(file),
		psql.NewSpotImporter(db, conf.BatchSize),
	)
	if err != nil {
		logger.WithError(err).Fatalf("failed to read import spots: %v", err)
	}

	duration := time.Since(start)

	logger.Infof("%d spot(s) were imported in %v", count, duration)
}
