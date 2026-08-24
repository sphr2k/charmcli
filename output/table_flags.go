package output

import "github.com/spf13/cobra"

// TableFlags owns the common flags for human list/table commands. Consumers
// should add it only where headers are meaningful; structured output ignores it.
type TableFlags struct {
	NoHeaders bool
}

// AddFlags binds the canonical --no-headers spelling.
func (f *TableFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.NoHeaders, "no-headers", false, "Do not print table headers")
}
