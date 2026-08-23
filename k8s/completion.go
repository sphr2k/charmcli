package k8s

import (
	"context"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// NamespaceListFunc returns namespace names for completion. Consumers may use
// any Kubernetes client implementation; charmcli/k8s does not require one.
type NamespaceListFunc func(context.Context) ([]string, error)

// RegisterNamespaceCompletion installs completion for the standard namespace
// flag using a supplied namespace reader.
func RegisterNamespaceCompletion(cmd *cobra.Command, list NamespaceListFunc) error {
	return cmd.RegisterFlagCompletionFunc("namespace", func(command *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		names, err := list(command.Context())
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		matches := make([]string, 0, len(names))
		for _, name := range names {
			if strings.HasPrefix(name, toComplete) {
				matches = append(matches, name)
			}
		}
		sort.Strings(matches)
		return matches, cobra.ShellCompDirectiveNoFileComp
	})
}
