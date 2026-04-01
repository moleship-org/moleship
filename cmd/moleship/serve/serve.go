package serve

import (
	"context"

	"github.com/moleship-org/moleship/internal/core/app"
	"github.com/moleship-org/moleship/internal/core/env"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "serve",
	Short: "serve starts a new moleship application server",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := make([]app.Option, 0)
		a := app.New(opts...)

		fEnvFile, err := cmd.Flags().GetString("env-file")
		if err != nil {
			return err
		}

		a.Logger().Info("Loading env file", "file", fEnvFile)
		if err := env.LoadFiles(fEnvFile); err != nil {
			return err
		}
		env.MustLoad()

		fPort, err := cmd.Flags().GetUint16("port")
		if err != nil {
			return err
		}
		if fPort != 0 && fPort != 6000 {
			opts = append(opts, app.WithPort(fPort))
		}

		a.Start(context.Background())
		return nil
	},
}

func Command() *cobra.Command {
	cmd.Flags().Uint16P("port", "p", 6000, "use to listen and serve")
	cmd.Flags().String("env-file", ".env", "read a file of environment variables")
	return cmd
}
