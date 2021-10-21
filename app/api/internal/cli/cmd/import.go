package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/ztimes2/tolqin/app/api/internal/cli/config"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf/csv"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf/psql"
	"github.com/ztimes2/tolqin/app/api/pkg/psqlutil"
)

func newCSVSpotCreationEntrySource(filename string) (*csv.SpotCreationEntrySource, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read csv file: %w", err)
	}

	return csv.NewSpotCreationEntrySource(bytes.NewReader(b)), nil
}

func newPostgresSpotStore() (*psql.SpotStore, error) {
	cfg, err := config.LoadDatabase()
	if err != nil {
		return nil, fmt.Errorf("could not load database config: %w", err)
	}

	db, err := psqlutil.NewDB(psqlutil.DriverNamePQ, psqlutil.Config{
		Host:         cfg.Host,
		Port:         cfg.Port,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DatabaseName: cfg.Name,
		SSLMode:      psqlutil.NewSSLMode(cfg.SSLMode),
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres db: %w", err)
	}

	return psql.NewSpotStore(db), nil
}

func newImportCmd(
	csvSourceFn func(filename string) (*csv.SpotCreationEntrySource, error),
	postgresStoreFn func() (*psql.SpotStore, error),
	importFn func(surf.SpotCreationEntrySource, surf.MultiSpotWriter) (int, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import spots from a CSV file to the database",
		Long: `Import spots from a CSV file to the database.

Environment variables:
  - DB_HOST
  - DB_PORT
  - DB_USERNAME
  - DB_PASSWORD
  - DB_NAME
  - DB_SSLMODE
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			filename, err := cmd.Flags().GetString("csv")
			if err != nil {
				return err
			}

			src, err := csvSourceFn(filename)
			if err != nil {
				return err
			}

			dest, err := postgresStoreFn()
			if err != nil {
				return err
			}

			n, err := importFn(src, dest)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%d spot(s) were imported!\n", n)

			return nil
		},
	}

	cmd.Flags().String("csv", "", "Name of a CSV file to import spots from.")

	return cmd
}
