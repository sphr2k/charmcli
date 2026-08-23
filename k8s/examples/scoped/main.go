package main

import (
	"fmt"
	"os"

	charmk8s "github.com/sphr2k/charmcli/k8s"
	"github.com/spf13/cobra"
)

func main() {
	scope := charmk8s.NewScope(charmk8s.WithAllNamespaces())
	cmd := &cobra.Command{
		Use: "scoped",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolved, err := scope.Resolve()
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "namespace=%q all=%t\n", resolved.Namespace, resolved.AllNamespaces)
			return err
		},
	}
	scope.AddFlags(cmd.Flags())
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
