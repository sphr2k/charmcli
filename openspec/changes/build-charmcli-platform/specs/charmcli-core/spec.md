## ADDED Requirements

### Requirement: Cobra command model remains native
charmcli MUST execute ordinary Cobra command trees and MUST NOT require consumers to describe commands through a charmcli-specific command DSL.

#### Scenario: Consumer constructs command tree
- **WHEN** a consumer provides a `*cobra.Command` root with native Cobra args, flags, aliases and completion
- **THEN** charmcli executes that tree through Fang without translating it into another command representation

### Requirement: Fang v2 owns normal CLI presentation
charmcli MUST use Fang v2 for standard help, usage, error, version, completion and manpage presentation.

#### Scenario: Help is requested
- **WHEN** the root or subcommand help is requested
- **THEN** Fang renders the help using its Charm-native presentation
- **AND** charmcli does not render a competing custom help format

### Requirement: Injectable process environment
charmcli MUST expose runtime streams and terminal capabilities as injectable state rather than requiring process-global IO.

#### Scenario: Command runs under test
- **WHEN** stdin, stdout and stderr are buffers and terminal capabilities are explicitly provided
- **THEN** command behavior is testable without a real terminal
- **AND** no output is written to process-global stdout/stderr

### Requirement: Standard stream contract
stdout MUST contain only the requested command result. Prompts, progress, diagnostics and rendered errors MUST use stderr.

#### Scenario: Structured output is piped
- **WHEN** a command selects JSON or YAML output while interactive/progress facilities are available
- **THEN** stdout contains one valid machine-readable document only
- **AND** transient UI and errors are not mixed into stdout

### Requirement: Typed exit semantics
charmcli MUST support standard exit classes and custom errors implementing an exit-code contract.

#### Scenario: Standard classes complete
- **WHEN** a command succeeds
- **THEN** exit code is 0
- **WHEN** a command returns a typed operational failure
- **THEN** exit code is 1
- **WHEN** a command returns a typed usage failure
- **THEN** exit code is 2
- **WHEN** execution is interrupted or an interactive form is aborted
- **THEN** exit code is 130

### Requirement: Silent non-zero result
charmcli MUST support a non-zero exit status that does not trigger Fang error rendering.

#### Scenario: Diagnostic command finds hard issues
- **WHEN** a command has already rendered a valid result and returns silent exit 1
- **THEN** the process exits 1
- **AND** Fang does not append an ERROR banner or duplicate diagnostic text

### Requirement: Error presentation may be sanitized
charmcli MUST allow an error transformation hook before human error rendering without changing classification of the original error.

#### Scenario: Secret-bearing error is returned
- **WHEN** a consumer installs a redaction transform
- **THEN** Fang receives the sanitized display error
- **AND** the exit code is determined from the original typed error

### Requirement: Output formats are shared but human views remain domain-owned
charmcli MUST provide common output selection and structured encoders without forcing all human output into one reflection-driven table abstraction.

#### Scenario: Rich domain detail is rendered
- **WHEN** a consumer selects human output for a hierarchical view
- **THEN** its domain renderer may write the human result directly through the shared output selection contract
- **AND** JSON/YAML selection remains handled consistently

### Requirement: Native Charm interactive design
Standard prompts MUST use Huh v2 with its Charm-native theme and context-aware execution.

#### Scenario: Interactive prompt runs
- **WHEN** runtime is interactive
- **THEN** Huh receives the configured input and stderr streams
- **AND** its Charm-native theme and accessibility behavior are retained

### Requirement: Exact destructive confirmation
charmcli MUST provide exact-text confirmation suitable for destructive operations.

#### Scenario: Expected resource name differs
- **WHEN** confirmation requires `worker-03` and the user submits another value
- **THEN** confirmation is rejected
- **AND** the destructive operation is not authorized

#### Scenario: Exact confirmation is non-interactive
- **WHEN** exact confirmation is required without an explicit bypass and no interactive terminal is available
- **THEN** the helper fails safely without consuming stdin or assuming consent

### Requirement: Adaptive progress
Progress/activity helpers MUST use Charm/Bubble Tea primitives for interactive rendering and deterministic linear stderr output otherwise.

#### Scenario: CI executes a long operation
- **WHEN** the runtime is non-interactive
- **THEN** progress emits stable line-oriented stderr messages
- **AND** no cursor-control animation is emitted

### Requirement: Core remains Kubernetes-free
The root charmcli module MUST have no Kubernetes dependency.

#### Scenario: Dependency guard runs
- **WHEN** root source imports and root `go.mod` requirements are inspected
- **THEN** no `k8s.io/*` dependency is present
