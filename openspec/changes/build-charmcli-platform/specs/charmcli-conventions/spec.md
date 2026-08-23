## ADDED Requirements

### Requirement: Canonical grammar is verb-first and resource-oriented
Reference consumers MUST use the common canonical verbs `get`, `describe`, `apply`, `doctor`, `explain`, and `suggest` according to their defined meanings.

#### Scenario: Inventory is requested
- **WHEN** a consumer exposes a concise resource inventory
- **THEN** the canonical verb is `get`
- **AND** the resource noun is plural

#### Scenario: One resource is inspected in detail
- **WHEN** a consumer exposes a detailed view of one object
- **THEN** the canonical verb is `describe`
- **AND** the resource noun is singular

#### Scenario: Desired state is converged
- **WHEN** a consumer exposes idempotent convergence of declared state
- **THEN** the canonical verb is `apply`

### Requirement: Incomplete invocation never performs an implicit action
Root commands and verb groups MUST show Fang help when required resource/operation arguments are absent.

#### Scenario: Root is invoked alone
- **WHEN** `homelabctl`, `policyopsctl`, or `resourcectl` is invoked without a command
- **THEN** help is shown
- **AND** no domain read or mutation is executed

#### Scenario: Verb group is incomplete
- **WHEN** `homelabctl get` or `resourcectl describe` is invoked without its required resource
- **THEN** help or a usage error is shown
- **AND** no default action is inferred

### Requirement: Resource naming is predictable
`get` MUST use plural resource names while object-oriented verbs MUST use singular resource names.

#### Scenario: Homelab nodes are listed
- **WHEN** the operator requests node inventory
- **THEN** the canonical command is `homelabctl get nodes`

#### Scenario: Homelab node detail is requested
- **WHEN** the operator requests detail for `worker-03`
- **THEN** the canonical command is `homelabctl describe node worker-03`

### Requirement: Shared flags have shared semantics
Where a common flag is exposed, it MUST use the common spelling and meaning.

#### Scenario: Output is selected
- **WHEN** a command supports alternate result formats
- **THEN** it uses `-o/--output`

#### Scenario: Confirmation is bypassed
- **WHEN** a mutating command supports bypassing a required confirmation
- **THEN** it uses `-y/--yes`
- **AND** the flag bypasses only the confirmation, not validation or safety preconditions

#### Scenario: Verbose evidence is requested
- **WHEN** a command supports additional human detail
- **THEN** it uses `--verbose`
- **AND** `-v` remains available for Fang/version behavior

### Requirement: Common global behavior is not silently ignored
A command MUST omit a shared flag when it cannot implement that flag's contract.

#### Scenario: Dry-run cannot be faithfully implemented
- **WHEN** an `apply` command cannot compute the intended transition without mutating external state
- **THEN** it does not expose `--dry-run`
- **AND** it does not expose a partial or misleading dry-run mode

### Requirement: Apply is idempotent convergence
`apply` MUST converge declared desired state and MAY result in create, update, replace, or no-op depending on current state.

#### Scenario: Node does not exist
- **WHEN** `homelabctl apply node NAME` resolves a valid absent cattle/disposable node
- **THEN** the transition may be `create`

#### Scenario: Node matches desired state
- **WHEN** `homelabctl apply node NAME` finds no required change
- **THEN** the transition is a successful no-op

#### Scenario: Node requires destructive replacement
- **WHEN** `homelabctl apply node NAME` computes a replacement transition
- **THEN** mutation requires the configured destructive confirmation unless `--yes` is supplied

### Requirement: Dry-run is mutation-free transition planning
When `--dry-run` is supported on `apply`, it MUST resolve inputs, validate non-mutating preconditions, compute the intended transition, present the result, and perform zero external mutation.

#### Scenario: Destructive transition is dry-run
- **WHEN** a replacement would be required and `--dry-run` is supplied
- **THEN** the replacement is reported
- **AND** no exact confirmation is required
- **AND** provider/Kubernetes/external mutation methods are not invoked

### Requirement: Destructive confirmation is transition-dependent
Mutating commands MUST request confirmation only after determining that the concrete transition is destructive.

#### Scenario: Apply is non-destructive
- **WHEN** an apply computes create, update, or no-op and the command's domain does not classify that transition as destructive
- **THEN** no destructive confirmation is requested

#### Scenario: Apply is destructive in non-interactive mode
- **WHEN** a destructive transition is required, runtime is non-interactive, and `--yes` is absent
- **THEN** the command fails before mutation

### Requirement: Shared output formats retain one meaning
The shared formats `human`, `wide`, `json`, `yaml`, and `name` MUST have consistent meaning across consumers.

#### Scenario: Wide output is requested
- **WHEN** `-o wide` is supported
- **THEN** it represents the same logical result as human/table output with additional high-signal fields

#### Scenario: JSON or YAML is requested
- **WHEN** `-o json` or `-o yaml` is selected
- **THEN** stdout contains stable structured result data with no ANSI

#### Scenario: Name output is requested
- **WHEN** `-o name` is selected
- **THEN** stdout contains stable machine-friendly resource identity only

### Requirement: Tabular get commands use common header semantics
A tabular `get` command MAY expose `--no-headers`; when it does, the flag MUST suppress only table headers and MUST NOT change rows or filtering.

#### Scenario: Headers are disabled
- **WHEN** `get ... --no-headers` is invoked
- **THEN** data rows remain unchanged
- **AND** only the header row is omitted

### Requirement: One human visual grammar
Reference consumers MUST use charmcli semantic presentation primitives for shared visual concepts once equivalent primitives exist.

#### Scenario: Human detail is rendered
- **WHEN** a consumer renders sections, key/value fields, code/next commands, findings or status emphasis
- **THEN** it uses the shared Charm-native visual hierarchy
- **AND** it does not maintain an independent competing palette for those concepts

### Requirement: Shared severity does not erase domain state
Human presentation MUST map domain states to `success`, `info`, `warning`, or `error` severity while preserving the original textual domain state.

#### Scenario: Domain state is Drift
- **WHEN** a consumer renders a `Drift` state as warning severity
- **THEN** the visible state remains `Drift`
- **AND** warning styling may reinforce it

### Requirement: Stream semantics are uniform
Requested results MUST use stdout; prompts, progress, debug diagnostics and rendered errors MUST use stderr.

#### Scenario: Machine output is combined with progress-capable work
- **WHEN** `-o json` or `-o yaml` is selected for a long-running command
- **THEN** stdout remains a single valid result document
- **AND** all progress goes to stderr or is suppressed

### Requirement: Progress behavior is adaptive and consistent
Long-running operations MUST use the common activity/step behavior when progress is shown.

#### Scenario: Interactive terminal runs an operation
- **WHEN** progress is enabled on a TTY
- **THEN** Charm/Bubbles rendering may animate on stderr

#### Scenario: CI or pipe runs an operation
- **WHEN** runtime is non-interactive
- **THEN** progress is deterministic line-oriented stderr output
- **AND** no cursor-control animation is emitted

### Requirement: Exit codes are uniform
Reference consumers MUST use the common exit contract.

#### Scenario: Valid diagnostic result has hard findings
- **WHEN** `doctor` completes successfully as a diagnostic operation but finds hard issues
- **THEN** it returns exit 1 without a duplicate Fang error banner

#### Scenario: Command execution itself fails
- **WHEN** API/RBAC/provider/configuration execution prevents the command from producing its requested result
- **THEN** it returns exit 2

#### Scenario: User aborts interaction
- **WHEN** the user cancels or interrupts an interactive flow
- **THEN** it returns exit 130

### Requirement: kubectl aliases are semantically identical
Standalone and kubectl-plugin invocations MUST use the same command tree and semantics.

#### Scenario: PolicyOps is invoked through either entry point
- **WHEN** equivalent arguments are passed to `policyopsctl` and `kubectl policyops`
- **THEN** command behavior, output, flags, and exit code are equivalent apart from displayed invocation identity

#### Scenario: Resource CLI is invoked through either entry point
- **WHEN** equivalent arguments are passed to `resourcectl` and `kubectl resources`
- **THEN** command behavior, output, flags, and exit code are equivalent apart from displayed invocation identity

### Requirement: Deprecated syntax is explicitly non-canonical
Legacy forms MAY exist temporarily but MUST delegate to the canonical implementation and MUST NOT define divergent semantics.

#### Scenario: Legacy alias remains during migration
- **WHEN** a retained legacy command such as resource `list` is invoked
- **THEN** it reaches the same implementation as canonical `get`
- **AND** help examples and new tests use `get`

### Requirement: Canonical homelabctl surface is resource-oriented
The migrated Homelab CLI MUST expose the canonical surface:

```text
homelabctl get cluster
homelabctl get nodes
homelabctl describe cluster
homelabctl describe node NAME
homelabctl apply cluster
homelabctl apply node NAME
homelabctl doctor
```

#### Scenario: Legacy bootstrap behavior is requested canonically
- **WHEN** cluster convergence is requested
- **THEN** the canonical command is `homelabctl apply cluster`
- **AND** existing bootstrap phase controls may remain as flags

#### Scenario: Legacy node status behavior is split canonically
- **WHEN** concise node inventory is requested
- **THEN** `homelabctl get nodes` is used
- **WHEN** detailed node inspection is requested
- **THEN** `homelabctl describe node NAME` is used

### Requirement: Canonical resourcectl surface is resource-oriented
The migrated Resource CLI MUST use `resourcectl` as the standalone product name and `get`/`describe` as canonical verbs.

#### Scenario: Resource inventory is requested
- **WHEN** the operator requests ResourceWorkloadState-oriented inventory
- **THEN** the standalone canonical command begins with `resourcectl get`
- **AND** `kubectl resources` remains a plugin alias for the same implementation

### Requirement: Consumer help and tests document only canonical syntax
Public help examples, README examples, and new golden/visual tests MUST use canonical command forms.

#### Scenario: Deprecated alias exists
- **WHEN** a compatibility alias remains implemented
- **THEN** it is hidden or marked deprecated
- **AND** canonical help/examples do not teach it
