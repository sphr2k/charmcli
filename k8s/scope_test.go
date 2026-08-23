package k8s

import (
	"errors"
	"testing"

	"github.com/sphr2k/charmcli"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestScopeBindsUpstreamFlagsAndAllNamespaces(t *testing.T) {
	scope := NewScope(WithAllNamespaces())
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	scope.AddFlags(flags)
	for _, name := range []string{"kubeconfig", "context", "namespace", "all-namespaces"} {
		if flags.Lookup(name) == nil {
			t.Fatalf("missing flag %q", name)
		}
	}
}

func TestScopeExposesUpstreamConfigFlags(t *testing.T) {
	flags := genericclioptions.NewConfigFlags(true)
	scope := NewScope(WithConfigFlags(flags))
	if scope.ConfigFlags() != flags {
		t.Fatal("scope did not retain upstream ConfigFlags")
	}
	if scope.RESTClientGetter() != flags {
		t.Fatal("RESTClientGetter is not upstream ConfigFlags")
	}
}

func TestExplicitNamespaceResolvesThroughUpstreamLoader(t *testing.T) {
	flags := genericclioptions.NewConfigFlags(true)
	namespace := "media"
	flags.Namespace = &namespace
	scope := NewScope(WithConfigFlags(flags))
	got, err := scope.ResolveNamespace()
	if err != nil {
		t.Fatal(err)
	}
	if got != namespace {
		t.Fatalf("namespace=%q, want %q", got, namespace)
	}
}

func TestAllNamespacesResolvesEmptyNamespace(t *testing.T) {
	scope := NewScope(WithAllNamespaces())
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	scope.AddFlags(flags)
	if err := flags.Set("all-namespaces", "true"); err != nil {
		t.Fatal(err)
	}
	resolved, err := scope.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if !resolved.AllNamespaces || resolved.Namespace != "" {
		t.Fatalf("resolved=%#v", resolved)
	}
}

func TestNamespaceAndAllNamespacesConflictIsUsage(t *testing.T) {
	flags := genericclioptions.NewConfigFlags(true)
	namespace := "media"
	flags.Namespace = &namespace
	scope := NewScope(WithConfigFlags(flags), WithAllNamespaces())
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	scope.AddFlags(flagSet)
	if err := flagSet.Set("all-namespaces", "true"); err != nil {
		t.Fatal(err)
	}
	_, err := scope.Resolve()
	if err == nil {
		t.Fatal("expected conflict")
	}
	var coder charmcli.ExitCoder
	if !errors.As(err, &coder) || coder.ExitCode() != 2 {
		t.Fatalf("err=%v does not classify as usage", err)
	}
}
