package serve

import (
	"context"
	"fmt"
	"time"

	"github.com/moleship-org/moleship/internal/app"
	"github.com/moleship-org/moleship/internal/domain/config"
	"github.com/spf13/cobra"
)

var (
	port            uint16
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	maxHeaderBytes  int
	shutdownTimeout time.Duration
)

var cmd = &cobra.Command{
	Use:   "serve",
	Short: "serve starts the Moleship server to listen and serve API requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := handleFlags()
		if err != nil {
			return err
		}

		a := app.New(opts...)
		a.Logger().Info(fmt.Sprintf("Running on '%s' mode", config.Current().Mode))

		a.Start(context.Background())
		return nil
	},
}

func Command() *cobra.Command {
	cmd.Flags().Uint16VarP(&port, "port", "p", 0, "use to listen and serve")
	return cmd
}

func handleFlags() ([]app.Option, error) {
	opts := make([]app.Option, 0)

	if port != 0 {
		opts = append(opts, app.WithPort(port))
	}
	if readTimeout != 0 {
		opts = append(opts, app.WithReadTimeout(readTimeout))
	}
	if writeTimeout != 0 {
		opts = append(opts, app.WithWriteTimeout(writeTimeout))
	}
	if idleTimeout != 0 {
		opts = append(opts, app.WithIdleTimeout(idleTimeout))
	}
	if maxHeaderBytes != 0 {
		opts = append(opts, app.WithMaxHeaderBytes(maxHeaderBytes))
	}
	if shutdownTimeout != 0 {
		opts = append(opts, app.WithShutdownTimeout(shutdownTimeout))
	}

	return opts, nil
}
