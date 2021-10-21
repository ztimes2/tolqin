package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ztimes2/tolqin/app/api/internal/cli/service/importing"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tolqin",
		Short: "A tool for operating Tolqin API",
	}

	cmd.SetOut(os.Stdout)
	cmd.SilenceErrors = true
	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.AddCommand(newImportCmd(newCSVSpotCreationEntrySource, newPostgresSpotStore, importing.ImportSpots))

	return cmd
}
