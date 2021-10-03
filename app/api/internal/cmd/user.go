package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ztimes2/tolqin/app/api/internal/auth"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/cmdio"
)

type userService interface {
	CreateUser(email, password string, role auth.Role) (auth.User, error)
}

func newUserCmd(cio *cmdio.IO, s userService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users of the application",
		Long:  "Manage users of the application",
	}
	cmd.AddCommand(newCreateUserCmd(cio, s))
	return cmd
}

func newCreateUserCmd(cio *cmdio.IO, s userService) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Long:  "Create a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, err := cio.Prompt("E-mail")
			if err != nil {
				return err
			}

			password, err := cio.Prompt("Password", cmdio.WithMask('*'))
			if err != nil {
				return err
			}

			role, err := cio.Select("Role", []string{
				auth.RoleAdmin.String(),
			})
			if err != nil {
				return err
			}

			user, err := s.CreateUser(email, password, auth.NewRole(role))
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), user.ID)
			return nil
		},
	}
}
