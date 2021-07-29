package main

import (
	"log"
	"os"
	"time"

	"github.com/ztimes2/tolqin/internal/csv"
	"github.com/ztimes2/tolqin/internal/importing"
	"github.com/ztimes2/tolqin/internal/logging"
	"github.com/ztimes2/tolqin/internal/postgres"
)

func main() {
	// TODO load from env config
	var logLevel, logFormat string
	logger, err := logging.NewLogger(logLevel, logFormat)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	// TODO load from env config
	var pgConf postgres.Config
	db, err := postgres.NewDB(pgConf)
	if err != nil {
		logger.WithError(err).Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// TODO load from env config
	var filename string
	file, err := os.Open(filename)
	if err != nil {
		logger.WithError(err).Fatalf("failed to read csv file: %v", err)
	}
	defer file.Close()

	start := time.Now()

	// TODO load from env config
	var batchSize int
	spots, err := importing.ImportSpots(
		csv.NewSpotEntries(file),
		postgres.NewSpotImporter(db, batchSize),
	)
	if err != nil {
		logger.WithError(err).Fatalf("failed to read import spots: %v", err)
	}

	duration := time.Since(start)

	logger.Infof("%d spot(s) were imported in %v", len(spots), duration)
}
