package k8s

import (
	"context"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestNamespaceCompletionFiltersAndSorts(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("namespace", "", "namespace")
	err := RegisterNamespaceCompletion(cmd, func(context.Context) ([]string, error) {
		return []string{"monitoring", "media", "default", "metallb-system"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	completion, ok := cmd.GetFlagCompletionFunc("namespace")
	if !ok {
		t.Fatal("namespace completion not registered")
	}
	got, directive := completion(cmd, nil, "me")
	want := []string{"media", "metallb-system"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("completion=%v, want %v", got, want)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive=%v", directive)
	}
}
