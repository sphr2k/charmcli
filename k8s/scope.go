package k8s

import (
	"fmt"

	"github.com/sphr2k/charmcli"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

// Scope is a thin kubectl-native scoping layer backed by the real Kubernetes
// ConfigFlags object. It deliberately does not copy kubeconfig/context/auth
// settings into a charmcli-owned configuration model.
type Scope struct {
	configFlags   *genericclioptions.ConfigFlags
	allNamespaces bool
	allowAll      bool
}

type ScopeOption func(*Scope)

// WithAllNamespaces enables the conventional -A/--all-namespaces flag.
func WithAllNamespaces() ScopeOption {
	return func(scope *Scope) {
		scope.allowAll = true
	}
}

// WithConfigFlags supplies an existing upstream ConfigFlags object. This is
// useful when a consumer already owns one or for focused tests.
func WithConfigFlags(flags *genericclioptions.ConfigFlags) ScopeOption {
	return func(scope *Scope) {
		if flags != nil {
			scope.configFlags = flags
		}
	}
}

// NewScope creates a scope using kubectl's standard ConfigFlags behavior.
func NewScope(options ...ScopeOption) *Scope {
	scope := &Scope{configFlags: genericclioptions.NewConfigFlags(true)}
	for _, option := range options {
		option(scope)
	}
	return scope
}

// AddFlags passes all upstream Kubernetes ConfigFlags directly onto flags and
// adds -A only when explicitly enabled for this command scope.
func (s *Scope) AddFlags(flags *pflag.FlagSet) {
	s.configFlags.AddFlags(flags)
	if s.allowAll {
		flags.BoolVarP(&s.allNamespaces, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces")
	}
}

// ConfigFlags returns the actual upstream ConfigFlags object.
func (s *Scope) ConfigFlags() *genericclioptions.ConfigFlags {
	return s.configFlags
}

// RESTClientGetter exposes the upstream interface expected by Kubernetes CLI
// helpers and other Kubernetes ecosystem packages.
func (s *Scope) RESTClientGetter() genericclioptions.RESTClientGetter {
	return s.configFlags
}

// RESTConfig resolves the active REST configuration through upstream
// Kubernetes loading behavior.
func (s *Scope) RESTConfig() (*rest.Config, error) {
	config, err := s.configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("resolve kubernetes REST config: %w", err)
	}
	return config, nil
}

// ResolvedScope is the small amount of scope state not already represented by
// RESTClientGetter.
type ResolvedScope struct {
	RESTClientGetter genericclioptions.RESTClientGetter
	Namespace        string
	AllNamespaces    bool
}

// Resolve resolves namespace through Kubernetes upstream loading semantics and
// validates the only charmcli-added scope rule: -n and -A are exclusive.
func (s *Scope) Resolve() (ResolvedScope, error) {
	namespace, err := s.ResolveNamespace()
	if err != nil {
		return ResolvedScope{}, err
	}
	return ResolvedScope{
		RESTClientGetter: s.RESTClientGetter(),
		Namespace:        namespace,
		AllNamespaces:    s.allNamespaces,
	}, nil
}

// ResolveNamespace returns an empty namespace for all-namespaces, otherwise it
// delegates to ConfigFlags' raw kubeconfig loader/current context.
func (s *Scope) ResolveNamespace() (string, error) {
	explicit := ""
	if s.configFlags.Namespace != nil {
		explicit = *s.configFlags.Namespace
	}
	if s.allNamespaces && explicit != "" {
		return "", charmcli.Usage(fmt.Errorf("--namespace and --all-namespaces cannot be used together"))
	}
	if s.allNamespaces {
		return "", nil
	}

	namespace, _, err := s.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return "", fmt.Errorf("resolve kubernetes namespace: %w", err)
	}
	if namespace == "" {
		namespace = "default"
	}
	return namespace, nil
}

// AllNamespaces reports the value of the optional -A flag.
func (s *Scope) AllNamespaces() bool {
	return s.allNamespaces
}
