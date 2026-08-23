# Design: charmcli platform

## Context

The reference CLIs already demonstrate useful patterns, but independently:

- Cobra + Fang gives `policyopsctl` a strong conventional command tree and attractive normal CLI UX.
- Huh/Bubble Tea/Lip Gloss provide the desired Charm-native interaction vocabulary.
- Kubernetes `genericclioptions.ConfigFlags` already expresses kubectl-native kubeconfig/context/namespace behavior and should remain authoritative.
- `homelabctl` has stronger destructive-interaction safety than a simple yes/no prompt, but its command grammar and rendering are bespoke.
- `resourcectl` has the weakest current CLI surface and therefore provides a useful migration test for the shared conventions.

The design optimizes for composition rather than replacement, but the platform is intentionally opinionated about product semantics. The three consumers should not merely share libraries; they should feel like one CLI family.

## Goals

- Give multiple Go CLIs one consistent execution, interaction, output, error and command-semantic model.
- Preserve Cobra as the command API visible to consumer code.
- Prefer native Charm defaults and APIs, adding only small semantic helpers.
- Keep machine output clean by construction.
- Make interactive and non-interactive behavior explicit and testable.
- Preserve strong destructive-operation guards such as typed-name confirmation.
- Keep Kubernetes support optional at module level.
- Delegate Kubernetes configuration semantics to upstream kubectl/client-go APIs.
- Make standalone and kubectl-plugin invocation names represent the same command tree.
- Make command naming, flags, output selection, human visuals and exit behavior predictable across all reference consumers.

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

charmcli owns execution around that tree, not the tree itself. The semantic conventions are enforced through helpers, tests and consumer acceptance, not by replacing Cobra.

## Decision 3 — App owns execution; Runtime owns environment

`App` is the process-level integration point. It configures Fang and maps command completion to a process exit code.

`Runtime` contains:

```text
stdin
stdout
stderr
terminal capabilities
interactive policy
accessibility policy
```

Streams are injectable. Tests MUST NOT depend on real process-global stdin/stdout/stderr.

## Decision 4 — Fang v2 is the normal CLI renderer

Use `charm.land/fang/v2`.

Fang owns help, usage, error presentation, version integration, completion, manpages, signal-aware execution and terminal color downsampling.

charmcli SHOULD pass through Fang configuration rather than mirror every Fang option.

A charmcli error handler wraps Fang only for optional sanitization/redaction, silent non-zero exits and consistent exit classification.

## Decision 5 — One resource-oriented command grammar

The canonical grammar is verb-first and resource-oriented:

```text
get <plural-resource> [name/filter flags]
describe <singular-resource> <name>
apply <singular-resource> [name]
explain <singular-resource> <name>
suggest <singular-resource> <target>
doctor
```

Not every CLI implements every verb. A verb is only exposed where it has meaningful domain semantics.

Canonical meaning:

- `get`: concise read-only inventory/current-state view; normally table-oriented human output.
- `describe`: detailed read-only view of one logical object/resource.
- `apply`: idempotently converge declared desired state; may create, update, replace or no-op.
- `doctor`: read-only diagnostics over a meaningful scope.
- `explain`: explain why a decision/state exists.
- `suggest`: produce a read-only recommendation or generated result without applying it.

Resource naming is consistent:

- `get` uses plural nouns (`nodes`, `bindings`, `workloads`, `resources`);
- object-oriented verbs use singular nouns (`node`, `binding`, `pod`, `resource`).

Root invocation and incomplete verb invocation MUST show Fang help. No root/verb group performs an implicit default action.

## Decision 6 — homelabctl adopts the same grammar

Canonical target:

```text
homelabctl get cluster
homelabctl get nodes
homelabctl describe cluster
homelabctl describe node NAME
homelabctl apply cluster
homelabctl apply node NAME
homelabctl doctor
```

`apply node` subsumes current create/update/replace/no-op paths. `apply cluster` replaces `cluster bootstrap` as user-facing semantics while retaining phase-range flags where useful.

The old `node create`, `node status`, `cluster bootstrap` forms are non-canonical compatibility surfaces only.

## Decision 7 — policyopsctl remains the semantic reference consumer

Canonical target remains:

```text
policyopsctl get workloads|profiles|bindings|exceptions
policyopsctl describe profile|binding NAME
policyopsctl explain pod NAME
policyopsctl suggest binding TYPE/NAME
policyopsctl doctor
```

The migration changes infrastructure and visual consistency, not PolicyOps domain meaning.

## Decision 8 — resourcectl adopts canonical resource semantics

Canonical target:

```text
resourcectl get resources [-n NAMESPACE | -A]
resourcectl describe resource NAME [-n NAMESPACE]

kubectl resources get resources ...
kubectl resources describe resource NAME ...
```

If the plugin-level `resources` name makes repeating `resources` ergonomically undesirable, a consumer spec MAY define a presentation alias, but the standalone canonical grammar remains resource-explicit. `list` may temporarily alias `get`; no-argument invocation shows help.

## Decision 9 — Shared global flag vocabulary

Where applicable, the same flags MUST have the same spelling and semantics:

```text
-o, --output
-y, --yes
--no-color
--non-interactive
--timeout
--verbose
--debug
--dry-run
```

Rules:

- `-v` is reserved for Fang/version behavior and MUST NOT mean verbose.
- `--yes/-y` only bypasses an otherwise-required confirmation; it does not weaken validation or safety preconditions.
- `--non-interactive` disables prompts and animated interaction; required unanswered input fails before mutation.
- `--timeout` is an overall command/workflow deadline where the command has bounded work.
- `--verbose` exposes additional human evidence/detail without changing domain action.
- `--debug` enables diagnostic logging/details and MUST write them to stderr.
- `--no-color` disables semantic color without removing textual meaning.
- `--quiet` is intentionally not a mandatory v1 convention; clean stream contracts and output formats make it unnecessary for the reference CLIs.

Flags that do not apply to a command are omitted rather than accepted and ignored.

## Decision 10 — Dry-run has one meaning

For `apply`, `--dry-run` means:

> Resolve inputs, validate preconditions, compute the intended transition and present the result, but perform no external mutation.

A command MUST NOT expose `--dry-run` unless it can satisfy that contract. Fake dry-run modes that merely skip the last mutation or omit meaningful preflight are forbidden.

Dry-run output follows the same output and stream contracts as normal execution. A destructive transition may be shown without requiring confirmation because no mutation occurs.

## Decision 11 — Exit status is typed behavior

Base convention:

```text
0    success / healthy result
1    valid command completed with negative domain result or hard findings
2    usage or command execution failure
130  user interruption/cancellation
```

Examples of exit 1: `doctor` finds hard issues; a status/read workflow validly determines required health is not satisfied.

Examples of exit 2: invalid arguments, kube API/RBAC failure, provider/API execution failure, malformed configuration, or required interaction impossible in non-interactive mode.

Errors may implement `ExitCoder`. charmcli provides:

```text
Usage(err)
Failure(err)
Exit(code)
```

A silent exit MUST NOT cause Fang to append an error banner.

## Decision 12 — stdout/stderr is a hard contract

`stdout` is reserved for the requested command result.

`stderr` is reserved for prompts, progress/activity UI, debug/diagnostic logs and Fang-rendered errors.

For machine-oriented formats, stdout MUST contain exactly the selected result and no ANSI or transient presentation.

## Decision 13 — Output formats are consistent

Shared formats are:

```text
human
wide
json
yaml
name
```

Commands expose only formats that make semantic sense, but any shared format has identical meaning across CLIs.

- `human`: default concise/readable presentation.
- `wide`: same logical result with additional high-signal fields.
- `json`/`yaml`: stable typed result models, except generator workflows such as `suggest ... -o yaml|json` whose result is intentionally a manifest.
- `name`: stable machine-friendly identity only.

`--no-headers` is available for tabular `get` commands where headers exist.

The output package remains small and callback-based rather than reflection-heavy.

## Decision 14 — One human visual grammar

Charm native presentation is mandatory by default:

- Fang default color scheme for normal CLI chrome;
- Huh `ThemeCharm` for interaction;
- Bubble Tea/Bubbles for live activity;
- Lip Gloss/Charmtone-derived styles for custom human output.

charmcli defines reusable semantic presentation primitives for:

```text
title
section
key/value
muted evidence
code / next command
table header
success
info
warning
error
finding
```

Consumers own domain text and states, but MUST map them onto this shared visual grammar rather than maintaining separate palettes/layout systems.

Color reinforces meaning but never carries meaning alone.

## Decision 15 — Severity is shared; domain state is not flattened

Presentation severity is standardized:

```text
success
info
warning
error
```

Domain states such as `Ready`, `Drift`, `Unmanaged`, `Shadowed`, `Stable`, `Converged` remain explicit. Each consumer maps those states to one of the shared presentation severities.

This avoids both visual inconsistency and loss of domain precision.

## Decision 16 — Huh helpers preserve semantics, not API duplication

The `ui` package provides small high-value workflows:

```text
Secret
Confirm
ConfirmExact
Select
MultiSelect
```

Consumers needing advanced forms use Huh directly.

Every helper uses explicit/inherited streams, stderr for interaction, context, accessible mode and non-interactive refusal where input is required.

`ConfirmExact` preserves typed-name confirmation for destructive Homelab node replacement. A migration MUST NOT weaken it to a boolean confirmation.

## Decision 17 — Destructive apply semantics are uniform

`apply` itself is not considered destructive; the computed transition may be.

Rules:

1. resolve desired/current state;
2. validate all non-mutating preconditions;
3. determine transition (`noop`, `create`, `update`, `replace`, etc.);
4. if transition is destructive and execution is mutating, require the command's declared confirmation policy;
5. `-y/--yes` bypasses only that confirmation;
6. perform mutation.

Non-interactive execution that reaches a required confirmation without `--yes` fails before mutation.

## Decision 18 — Bubble Tea abstraction stops at progress/workflows

charmcli provides reusable activity/step components, not a framework around arbitrary full-screen Bubble Tea applications.

Progress adapts:

```text
interactive TTY     animated Charm/Bubbles presentation on stderr
non-interactive     deterministic linear messages on stderr
machine stdout      never polluted
```

## Decision 19 — Accessibility and terminal behavior are first-class

Huh accessible mode and `TERM=dumb` behavior are reused rather than reimplemented. Runtime capabilities are overridable for tests.

`NO_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE` and color-profile downsampling should be delegated to Charm libraries where possible.

## Decision 20 — Kubernetes support is a first-party optional extension

`github.com/sphr2k/charmcli/k8s` is an integration layer, not a Kubernetes SDK.

It MUST use upstream `genericclioptions.ConfigFlags` / `RESTClientGetter` semantics and MUST NOT define a parallel Kubernetes configuration model.

Useful operations include:

```text
AddFlags(*cobra.Command)
ConfigFlags()
RESTClientGetter()
RESTConfig()
ResolveNamespace()
Resolve()
```

## Decision 21 — Scope adds only missing kubectl ergonomics

`ConfigFlags.AddFlags` is passed through natively. charmcli/k8s adds only optional `-A`, `-n`/`-A` validation, current-context namespace resolution, namespace completion, plugin identity and test helpers.

Auth/TLS/context/impersonation/API-server semantics remain upstream behavior.

## Decision 22 — controller-runtime stays outside charmcli/k8s

The extension exposes `*rest.Config` and `genericclioptions.RESTClientGetter` but MUST NOT require controller-runtime. Consumers choose their client implementation.

## Decision 23 — kubectl aliases have semantic parity

The same command tree is presented under:

```text
policyopsctl ...
kubectl policyops ...

resourcectl ...
kubectl resources ...
```

Only the displayed invocation identity may differ. Command availability, flags, output bytes, exit codes and domain behavior MUST otherwise be equivalent.

## Decision 24 — Help and documentation follow one style

Every public command SHOULD provide concise `Short` text and representative examples. Canonical examples MUST use the new grammar, never deprecated aliases.

Root and verb-group help MUST be useful and deterministic enough for golden/visual tests.

## Decision 25 — Compatibility is temporary and explicit

Breaking changes are accepted to reach the unified model.

Legacy forms MAY remain as hidden/deprecated aliases during migration, but:

- they are not documented as canonical;
- new examples/tests use the canonical grammar;
- aliases delegate to the same implementation;
- aliases MUST NOT retain divergent flags/output/exit semantics;
- removal is allowed after the migration window.

## Testing strategy

Core tests cover streams, exit classes, sanitized presentation, output cleanliness, Huh non-interactive refusal, exact confirmation, accessibility and progress fallback.

Convention contract tests cover:

- root/incomplete commands show help and do not execute work;
- canonical plural/singular resource grammar;
- shared flag spelling and semantics;
- structured output is ANSI-free and stdout-clean;
- shared human render primitives produce consistent hierarchy;
- domain state maps to shared severity without losing its textual state;
- destructive apply requires confirmation only when mutation is destructive;
- dry-run performs zero mutation;
- non-interactive required confirmation fails before mutation;
- deprecated aliases reach the same implementation where retained.

Kubernetes tests cover dependency boundaries, ConfigFlags passthrough, scope resolution, completion, plugin identity and absence of controller-runtime.

Consumer acceptance after migration covers:

- PolicyOps command semantics remain intact under the common visual/output/error contracts;
- Resource CLI uses `resourcectl` and canonical `get`/`describe` semantics;
- Homelab uses `get`/`describe`/`apply`/`doctor`, with `apply node` subsuming create/update/replace/no-op;
- Homelab typed-name destructive confirmation is not weakened;
- all machine output remains stdout-clean;
- standalone and kubectl aliases are semantically equivalent;
- consumer-local parser, palette, TTY, progress and Kubernetes plumbing is removed once charmcli provides the equivalent.
