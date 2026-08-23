package main

import (
	"context"
	"os"

	"github.com/sphr2k/charmcli"
	"github.com/spf13/cobra"
)

func main() {
	app := charmcli.New(charmcli.WithVersion("dev"))
	code := app.Run(context.Background(), os.Args[1:], charmcli.DefaultStreams(), func(_ *charmcli.Runtime) *cobra.Command {
		return &cobra.Command{
			Use: "example",
			RunE: func(cmd *cobra.Command, _ []string) error {
				_, err := cmd.OutOrStdout().Write([]byte("hello\n"))
				return err
			},
		}
	})
	os.Exit(code)
}
