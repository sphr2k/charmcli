package output

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestTableFlagsNoHeaders(t *testing.T) {
	flags := &TableFlags{}
	cmd := &cobra.Command{Use: "get"}
	flags.AddFlags(cmd)
	if err := cmd.Flags().Parse([]string{"--no-headers"}); err != nil {
		t.Fatal(err)
	}
	if !flags.NoHeaders {
		t.Fatal("--no-headers was not bound")
	}
}
