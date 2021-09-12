package main

import (
	"log"
	"os"
	"time"

	configimporter "github.com/ztimes2/tolqin/internal/config/importer"
	"github.com/ztimes2/tolqin/internal/importing"
	"github.com/ztimes2/tolqin/internal/importing/csv"
	"github.com/ztimes2/tolqin/internal/importing/psql"
	"github.com/ztimes2/tolqin/internal/pkg/logging"
	"github.com/ztimes2/tolqin/internal/pkg/psqlutil"
)

func main() {
	conf, err := configimporter.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := logging.New(conf.LogLevel, conf.LogFormat)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	file, err := os.Open(conf.CSVFile)
	if err != nil {
		logger.WithError(err).Fatalf("failed to read csv file: %v", err)
	}
	defer file.Close()

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

	start := time.Now()

	count, err := importing.ImportSpots(
		csv.NewSpotEntrySource(file),
		psql.NewSpotImporter(db, conf.BatchSize),
	)
	if err != nil {
		logger.WithError(err).Fatalf("failed to import spots: %v", err)
	}

	duration := time.Since(start)

	logger.Infof("%d spot(s) were imported in %v", count, duration)
}
