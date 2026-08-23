## ADDED Requirements

### Requirement: Kubernetes integration is a separate module
Kubernetes CLI support MUST live in `github.com/sphr2k/charmcli/k8s` as a separate Go module that may depend on charmcli core.

#### Scenario: Core-only consumer resolves dependencies
- **WHEN** a CLI imports only `github.com/sphr2k/charmcli`
- **THEN** Kubernetes client libraries are not required by the core module

### Requirement: Upstream kubectl configuration is authoritative
The Kubernetes module MUST delegate kubeconfig, context, cluster, auth, namespace and REST configuration semantics to Kubernetes upstream `genericclioptions`/client-go APIs.

#### Scenario: Scope is constructed
- **WHEN** a consumer creates a Kubernetes scope
- **THEN** it is backed by a real `genericclioptions.ConfigFlags`
- **AND** charmcli/k8s does not copy those settings into an independent configuration struct

### Requirement: Standard Kubernetes flags are passed through
The Kubernetes module MUST bind upstream ConfigFlags directly to Cobra commands.

#### Scenario: Command binds cluster flags
- **WHEN** `Scope.AddFlags` is called
- **THEN** standard upstream flags such as kubeconfig, context and namespace are available with Kubernetes-native behavior

### Requirement: REST client interfaces remain exposed
The Kubernetes module MUST expose upstream-compatible REST configuration interfaces rather than forcing a charmcli-specific Kubernetes client.

#### Scenario: Consumer uses controller-runtime
- **WHEN** a consumer requests a REST config from Scope
- **THEN** it can construct its own controller-runtime client from that config
- **AND** charmcli/k8s itself does not require controller-runtime

### Requirement: Namespace scope follows kubectl semantics
Namespaced commands MAY opt into `-A/--all-namespaces`; explicit namespace and all-namespaces are mutually exclusive.

#### Scenario: No namespace flag is supplied
- **WHEN** a namespaced command resolves scope without `-n` or `-A`
- **THEN** namespace is resolved through the upstream current-context loader

#### Scenario: All namespaces is selected
- **WHEN** `-A` is true
- **THEN** resolved namespace is empty
- **AND** all-namespaces is reported true

#### Scenario: Namespace conflicts with all namespaces
- **WHEN** both an explicit namespace and `-A` are supplied
- **THEN** resolution returns a usage error before domain API work begins

### Requirement: kubectl plugin identity is presentation-only
The module MUST support standalone and kubectl-plugin display identities for the same command tree.

#### Scenario: PolicyOps standalone is invoked
- **WHEN** executable basename is `policyopsctl`
- **THEN** display invocation is `policyopsctl`

#### Scenario: PolicyOps plugin is invoked
- **WHEN** executable basename is `kubectl-policyops`
- **THEN** display invocation is `kubectl policyops`

#### Scenario: Resource standalone is invoked
- **WHEN** executable basename is `resourcectl`
- **THEN** display invocation is `resourcectl`

#### Scenario: Resource plugin is invoked
- **WHEN** executable basename is `kubectl-resources`
- **THEN** display invocation is `kubectl resources`

### Requirement: Namespace completion is reusable
The Kubernetes module SHOULD provide a namespace completion helper that resolves namespaces through a supplied reader/client while retaining Cobra completion semantics.

#### Scenario: Namespace completion succeeds
- **WHEN** a user completes `--namespace` with a prefix
- **THEN** only matching namespace names are returned
- **AND** file completion is disabled

### Requirement: Kubernetes extension remains infrastructure-only
The module MUST NOT contain Homelab, PolicyOps, ResourceWorkloadState, Kyverno, controller-runtime-domain, or other consumer-specific meaning.

#### Scenario: Consumer domain evolves
- **WHEN** a PolicyOps or Resource Controller API changes
- **THEN** charmcli/k8s requires no change unless the generic Kubernetes CLI integration contract itself changes
