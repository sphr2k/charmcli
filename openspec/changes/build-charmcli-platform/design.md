# Design: charmcli platform

## Context

The reference CLIs already demonstrate three useful patterns, but independently:

- Cobra + Fang gives `policyopsctl` a strong conventional command tree and attractive normal CLI UX.
- Huh/Bubble Tea/Lip Gloss provide the desired Charm-native interaction vocabulary.
- Kubernetes `genericclioptions.ConfigFlags` already expresses kubectl-native kubeconfig/context/namespace behavior and should remain authoritative.

The design therefore optimizes for composition rather than replacement.

## Goals

- Give multiple Go CLIs one consistent execution, interaction, output and error model.
- Preserve Cobra as the command API visible to consumer code.
- Prefer native Charm defaults and APIs, adding only small semantic helpers.
- Keep machine output clean by construction.
- Make interactive and non-interactive behavior explicit and testable.
- Preserve strong destructive-operation guards such as typed-name confirmation.
- Keep Kubernetes support optional at module level.
- Delegate Kubernetes configuration semantics to upstream kubectl/client-go APIs.
- Make standalone and kubectl-plugin invocation names represent the same command tree.

## Decision 1 — Repository contains two modules

Repository layout:

```text
/
  go.mod                         github.com/sphr2k/charmcli
  go.work
  app.go
  runtime.go
  exit.go
  output/
  render/
  ui/
  testkit/

/k8s
  go.mod                         github.com/sphr2k/charmcli/k8s
  scope.go
  invocation.go
  completion.go
  testkit/
```

The root module MUST NOT import `k8s.io/*`.

The `/k8s` module may depend on the root module and Kubernetes libraries. Workspace development uses `go.work`; releases use normal multi-module tags (`vX.Y.Z`, `k8s/vX.Y.Z`).

## Decision 2 — Cobra remains the public command model

Consumer packages build ordinary `*cobra.Command` values. charmcli does not define a parallel command tree, argument grammar or command descriptor system.

Typical consumer shape:

```go
func NewRootCommand(rt *charmcli.Runtime, deps Dependencies) *cobra.Command
```

charmcli owns execution around that tree, not the tree itself.

This keeps all Cobra facilities available directly: `Args`, `ValidArgsFunction`, persistent flags, annotations, completion, aliases and ecosystem integrations.

## Decision 3 — App owns execution; Runtime owns environment

`App` is the process-level integration point. It configures Fang and maps command completion to a process exit code.

`Runtime` contains the environment that commands and UI helpers need:

```text
stdin
stdout
stderr
terminal capabilities
interactive policy
accessibility policy
```

Streams are injectable. Tests MUST NOT depend on real process-global stdin/stdout/stderr.

The runtime distinguishes raw streams from presentation concerns so structured output is not accidentally mixed with progress or prompts.

## Decision 4 — Fang v2 is the normal CLI renderer

Use `charm.land/fang/v2`.

Fang owns:

- help and usage presentation;
- error presentation;
- version integration;
- shell completion command;
- manpages;
- signal-aware execution where configured;
- terminal color downsampling.

charmcli SHOULD pass through Fang configuration rather than mirror every Fang option in its own API.

A charmcli error handler wraps Fang only to support:

1. optional error sanitization/redaction before display;
2. silent non-zero exits;
3. consistent exit classification.

## Decision 5 — Exit status is typed behavior

Base convention:

```text
0    success
1    operational/domain failure or diagnostic hard findings
2    usage/CLI invocation failure
130  user interruption/cancellation
```

Errors may implement an `ExitCoder` contract. charmcli provides helpers for common classes:

```text
Usage(err)
Failure(err)
Exit(code)          // silent status
```

A silent exit MUST NOT cause Fang to print an error banner. This is required for commands such as `doctor`, where the command has already rendered a valid diagnostic result but intentionally returns status 1.

Consumer code remains free to define domain-specific `ExitCoder` errors.

## Decision 6 — stdout/stderr is a hard contract

`stdout` is reserved for the requested command result.

`stderr` is reserved for:

- prompts;
- progress/activity UI;
- diagnostics not constituting the result;
- Fang-rendered errors.

For structured output (`json`, `yaml`, `name` or another explicitly machine-oriented format), stdout MUST contain only the selected format and MUST contain no ANSI escape sequences.

Interactive UI MUST NOT write transient rendering to stdout.

## Decision 7 — Output abstraction is intentionally small

The initial output package standardizes:

- common format values;
- `-o/--output` flag binding;
- validation of allowed formats;
- JSON and YAML encoding;
- selection between human/wide/name render callbacks.

It does NOT attempt to infer rich human tables from struct tags or use reflection to model every CLI view.

Domain-specific human rendering stays in consumer packages because `policyopsctl` has hierarchical operational views that are not equivalent to a table.

## Decision 8 — Native Charm design is the default

Use the native Charm visual vocabulary unless a specific usability requirement demands customization:

- Fang default color scheme for command help/errors;
- Huh `ThemeCharm` for forms;
- Bubbles defaults for reusable Bubble Tea components;
- Lip Gloss v2 for custom human views.

charmcli does not preserve the existing Clack-inspired `homelabctl` appearance as a design requirement.

Where custom semantic styling is required, charmcli should derive it from Charm-native theme primitives instead of introducing an unrelated palette.

## Decision 9 — Huh helpers preserve semantics, not API duplication

The `ui` package provides small high-value workflows rather than wrapping every Huh field method:

```text
Secret
Confirm
ConfirmExact
Select
MultiSelect
```

Consumers needing advanced forms should use Huh directly.

Every helper:

- accepts/inherits explicit input/output streams;
- writes interaction to stderr by default through Runtime;
- runs with context;
- respects accessibility mode;
- rejects required interaction when policy is non-interactive unless an explicit bypass/default exists.

### Exact confirmation

`ConfirmExact` requires the user to type an expected string. This preserves the current `homelabctl node replace` safety property. A migration to Charm-native UI MUST NOT weaken this to a boolean confirmation.

## Decision 10 — Bubble Tea abstraction stops at progress/workflows

charmcli may provide reusable progress/activity/step components for long-running command workflows.

It MUST NOT provide a generic wrapper around arbitrary full-screen Bubble Tea applications. Consumers needing a custom TUI use Bubble Tea directly.

Progress adapts by environment:

```text
interactive TTY     animated Charm/Bubbles presentation
non-interactive     deterministic linear messages on stderr
machine stdout      never polluted
```

## Decision 11 — Accessibility and terminal behavior are first-class

Huh's accessible mode and `TERM=dumb` behavior are reused rather than reimplemented.

Runtime capability detection is overridable for tests and unusual execution environments.

`NO_COLOR` and terminal color-profile behavior should be delegated to Charm libraries wherever possible instead of each consumer manually checking environment variables and stripping ANSI.

## Decision 12 — Kubernetes support is a first-party optional extension

`github.com/sphr2k/charmcli/k8s` exists because two reference CLIs are kubectl-like and `homelabctl` also consumes kubeconfig/context behavior.

It is an integration layer, not a Kubernetes SDK.

The module MUST use Kubernetes upstream `genericclioptions.ConfigFlags` / `RESTClientGetter` semantics as the source of truth.

It MUST NOT define a parallel configuration model like:

```go
type Config struct {
    Kubeconfig string
    Context string
    Namespace string
}
```

Instead it may hold and expose the real upstream object:

```go
type Scope struct {
    ConfigFlags *genericclioptions.ConfigFlags
    AllNamespaces bool
}
```

Useful public operations include:

```text
AddFlags(*cobra.Command)
ConfigFlags()
RESTClientGetter()
RESTConfig()
ResolveNamespace()
Resolve()
```

## Decision 13 — Scope adds only missing kubectl ergonomics

`ConfigFlags.AddFlags` is passed through natively.

charmcli/k8s adds only behavior that reference consumers otherwise repeat:

- optional `-A/--all-namespaces`;
- mutual-exclusion validation for namespace and all-namespaces;
- current-context namespace resolution;
- namespace completion helper;
- kubectl-plugin invocation display identity;
- test helpers.

Kubeconfig parsing, context overrides, auth, TLS, impersonation, API server selection and REST config creation remain upstream Kubernetes behavior.

## Decision 14 — controller-runtime stays outside charmcli/k8s

The extension may expose `*rest.Config` and `genericclioptions.RESTClientGetter`.

It MUST NOT depend on `sigs.k8s.io/controller-runtime` merely because current consumers happen to use its client. Consumers construct whichever Kubernetes client implementation they need.

## Decision 15 — kubectl plugin identity is a reusable concept

The same command tree can be presented under a standalone and kubectl-plugin identity:

```text
policyopsctl ...
kubectl policyops ...

resourcectl ...
kubectl resources ...
```

An invocation helper derives a display name from `argv[0]` without changing domain behavior. Packaging may use a second binary name, symlink, hardlink or equivalent; that choice is outside the command tree.

## Decision 16 — resourcectl migration target

The framework must support the intended Resource CLI migration:

```text
resourcectl get [-n NAMESPACE | -A]
resourcectl describe NAME [-n NAMESPACE]

kubectl resources get ...
kubectl resources describe ...
```

`list` may remain an alias for `get` during migration. Root invocation without a command should show Fang help instead of implicitly performing a list operation.

## Testing strategy

Core tests cover:

- injected streams;
- default/raw error exit mapping;
- usage/failure/silent/cancelled exits;
- error transform applied to display but not classification;
- structured output JSON/YAML cleanliness;
- Huh helper non-interactive refusal;
- exact confirmation validation;
- accessible mode wiring;
- progress non-interactive fallback.

Kubernetes tests cover:

- root core module imports no Kubernetes packages;
- upstream ConfigFlags are exposed and bind standard flags;
- namespace/current-context resolution is upstream-driven;
- `-n` and `-A` conflict deterministically;
- all-namespaces resolves to empty namespace;
- standalone and kubectl-plugin invocation display names;
- no controller-runtime dependency.

Consumer acceptance after migration covers:

- `policyopsctl` standalone and kubectl invocation remain behaviorally equivalent;
- `resourcectl` is canonical while `kubectl resources` remains supported;
- `homelabctl` typed-name destructive confirmation is not weakened;
- `homelabctl node status -o json` remains stdout-clean;
- current consumer-specific manual TTY/style/parser infrastructure can be removed rather than duplicated beside charmcli.
