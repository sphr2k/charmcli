package k8s

import "path/filepath"

// Invocation describes the standalone and kubectl-plugin identities for one
// command tree.
type Invocation struct {
	Standalone string
	Plugin     string
}

// DisplayName returns the user-facing command name for argv0.
func (i Invocation) DisplayName(argv0 string) string {
	base := filepath.Base(argv0)
	if i.Plugin != "" && base == "kubectl-"+i.Plugin {
		return "kubectl " + i.Plugin
	}
	if base == "" || base == "." {
		return i.Standalone
	}
	if i.Standalone != "" && base == i.Standalone {
		return i.Standalone
	}
	return base
}
