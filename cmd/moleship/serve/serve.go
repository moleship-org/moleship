package serve

import (
	"context"
	"fmt"
	"strings"

	"codeberg.org/ungo/env/dotenv"
	"github.com/moleship-org/moleship/internal/core/app"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "serve",
	Short: "serve starts the Moleship server to listen and serve API requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		fEnvFile, err := cmd.Flags().GetString("env-file")
		if err != nil {
			return err
		}
		if fEnvFile != "" {
			if err := dotenv.Load(fEnvFile); err != nil {
				return fmt.Errorf("failed to load environment file: %w", err)
			}
		}

		opts := make([]app.Option, 0)

		fPort, err := cmd.Flags().GetUint16("port")
		if err != nil {
			return err
		}
		if fPort != 0 {
			opts = append(opts, app.WithPort(fPort))
		}

		a := app.New(opts...)
		if fEnvFile != "" {
			a.Logger().Info("Environment file loaded", "file", fEnvFile)
		}
		a.Logger().Info(fmt.Sprintf("Running on '%s' mode", strings.ToUpper(a.Config().Vars.Mode)))

		a.Start(context.Background())
		return nil
	},
}

func Command() *cobra.Command {
	cmd.Flags().Uint16P("port", "p", 0, "use to listen and serve")
	cmd.Flags().String("env-file", "", "read a file of environment variables")
	return cmd
}
