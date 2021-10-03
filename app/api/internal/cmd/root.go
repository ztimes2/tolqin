package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/cmdio"
)

func New() *cobra.Command {
	cio := cmdio.New(os.Stdin, os.Stdout)

	cmd := &cobra.Command{
		Use: "tolqin",
		// TODO add descriptions
	}
	cmd.SetIn(os.Stdin)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SilenceErrors = true
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.AddCommand(newUserCmd(cio))

	return cmd
}
