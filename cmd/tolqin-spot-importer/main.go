package main

import (
	"log"
	"os"
	"time"

	"github.com/ztimes2/tolqin/internal/config"
	"github.com/ztimes2/tolqin/internal/csv"
	"github.com/ztimes2/tolqin/internal/importing"
	"github.com/ztimes2/tolqin/internal/logging"
	"github.com/ztimes2/tolqin/internal/postgres"
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

	file, err := os.Open(conf.CSVFile)
	if err != nil {
		logger.WithError(err).Fatalf("failed to read csv file: %v", err)
	}
	defer file.Close()

	start := time.Now()

	spots, err := importing.ImportSpots(
		csv.NewSpotEntries(file),
		postgres.NewSpotImporter(db, conf.BatchSize),
	)
	if err != nil {
		logger.WithError(err).Fatalf("failed to read import spots: %v", err)
	}

	duration := time.Since(start)

	logger.Infof("%d spot(s) were imported in %v", len(spots), duration)
}
