package main

import (
	"log"

	"github.com/moleship-org/moleship/cmd/moleship/serve"
	"github.com/spf13/cobra"
)

// @title           Mi API de Microservicios
// @version         1.0
// @description     Esta es una descripción detallada de lo que hace mi API.
// @termsOfService  http://swagger.io/terms/
// @contact.name   Soporte API
// @contact.url    http://www.tu-sitio.com/support
// @contact.email  soporte@tu-dominio.com
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:5000
// @BasePath  /api/v1
func main() {
	if err := NewCommand().Execute(); err != nil {
		log.Fatalln(err)
	}
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "moleship",
		Short: "Moleship is ...",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
			}
			return nil
		},
	}
	AddSubcommands(cmd)
	return cmd
}

func AddSubcommands(cmd *cobra.Command) {
	cmd.AddCommand(serve.Command())
}
