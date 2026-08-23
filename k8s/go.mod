module github.com/sphr2k/charmcli/k8s

go 1.26.0

require (
	github.com/sphr2k/charmcli v0.0.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.9
	k8s.io/cli-runtime v0.36.3
	k8s.io/client-go v0.36.3
)

// Pre-1.0 workspace development: replace with a tagged core version before the
// first standalone k8s module release. Dependency-module replace directives are
// ignored by downstream consumers, so the released module must carry a real
// core requirement.
replace github.com/sphr2k/charmcli => ..
