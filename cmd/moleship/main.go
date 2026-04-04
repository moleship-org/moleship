package main

import (
	"log"

	"github.com/moleship-org/moleship/cmd/moleship/serve"
	"github.com/spf13/cobra"
)

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
