## Why

The homelab toolchain currently has three Go CLIs with overlapping concerns but different command, rendering, interaction, error, output, and Kubernetes plumbing:

- `homelabctl` uses a bespoke parser plus hand-built Clack-inspired Lip Gloss interaction and progress rendering;
- `policyopsctl` uses Cobra + Fang and Kubernetes `genericclioptions`, but owns its own output styling, invocation-name handling, namespace completion, and exit semantics;
- `kubectl-resources` uses a hand-written parser and direct clientcmd loading and is being renamed to the standalone product name `resourcectl`, while `kubectl resources` remains a supported plugin alias.

The duplicated code is not primarily domain logic. It is CLI product infrastructure: execution, IO contracts, machine output, error/exit semantics, prompts, destructive confirmation, progress, terminal capability handling, Kubernetes CLI conventions, and—critically—user-facing command semantics.

The desired result is one coherent, Charm-native CLI product language across all three tools. A user who understands one CLI should be able to predict the command grammar, flags, output behavior, interaction model, visuals, errors, and exit semantics of the others.

## What Changes

Create `github.com/sphr2k/charmcli` as an opinionated but thin Go CLI platform built on the current Charm v2 stack:

- Cobra remains the public command model; charmcli MUST NOT invent a replacement command DSL.
- Fang v2 owns normal help, usage, errors, version, completion and manpage UX.
- Huh v2 owns standard interactive forms and prompts.
- Bubble Tea v2 + Bubbles v2 own stateful/live terminal interaction and progress primitives where required.
- Lip Gloss v2 owns custom human presentation.
- Charm-native defaults are preferred over a custom visual design system.

Create a first-party Kubernetes extension as a separate Go module in the same repository:

```text
github.com/sphr2k/charmcli
github.com/sphr2k/charmcli/k8s
```

The core module MUST have no `k8s.io/*` dependency. The Kubernetes module delegates kubeconfig, context, auth, cluster, namespace and REST-client semantics to Kubernetes upstream `genericclioptions` / client-go instead of maintaining an independent configuration model.

Standardize one cross-CLI behavioral and semantic contract:

- command grammar is verb-first and resource-oriented;
- `get` lists/reads concise resource state;
- `describe` shows detailed human-readable state for one resource;
- `apply` converges declared desired state and is idempotent;
- `doctor` performs read-only diagnostics;
- `explain` and `suggest` remain specialized read-only verbs where domain semantics require them;
- `get` uses plural resource names; object-oriented verbs use singular resource names;
- root and incomplete verb invocation show Fang help and never trigger an implicit action;
- shared flags have the same spelling and semantics everywhere;
- stdout contains the requested result only;
- stderr contains prompts, progress, diagnostics and rendered errors;
- structured output is deterministic and ANSI-free;
- interactive features degrade safely in non-TTY/CI environments;
- exact destructive confirmation is available so `homelabctl` does not regress from its current typed-name guard;
- `-y/--yes` is the common explicit bypass for a required confirmation;
- `--dry-run` has one shared meaning for `apply`: compute and present the intended transition without mutation, and is exposed only when correctly implementable;
- user cancellation is a first-class exit condition;
- commands may return a silent non-zero status for diagnostic commands such as `doctor` without an additional Fang error banner.

Human presentation is standardized as well. charmcli owns a small semantic visual vocabulary for titles, sections, tables, key/value detail, code/next actions, findings, and success/info/warning/error severity. Domain-specific states remain domain-owned but are mapped onto that shared presentation vocabulary. Consumers MUST NOT maintain independent competing palettes or human-layout conventions once equivalent charmcli primitives exist.

## Canonical Consumer Direction

The canonical products remain three binaries, with kubectl plugin aliases where appropriate:

```text
homelabctl
policyopsctl
resourcectl
kubectl policyops   -> policyopsctl command tree
kubectl resources   -> resourcectl command tree
```

Canonical command language after migration:

```text
homelabctl get cluster
homelabctl get nodes
homelabctl describe cluster
homelabctl describe node NAME
homelabctl apply cluster
homelabctl apply node NAME
homelabctl doctor

policyopsctl get workloads|profiles|bindings|exceptions
policyopsctl describe profile|binding NAME
policyopsctl explain pod NAME
policyopsctl suggest binding TYPE/NAME
policyopsctl doctor

resourcectl get resources
resourcectl describe resource NAME

kubectl policyops ...     # semantically identical to policyopsctl
kubectl resources ...     # semantically identical to resourcectl
```

The old `homelabctl cluster bootstrap`, `node create`, `node status`, and resource `list` forms are non-canonical. Compatibility aliases MAY exist temporarily where useful, but documentation, help examples and new tests MUST use the canonical grammar.

`homelabctl apply node` subsumes create/update/replace/no-op behavior. Destructive replacement still requires exact-name confirmation unless `-y/--yes` is supplied. `homelabctl apply cluster` replaces bootstrap as the user-facing convergence verb while preserving phase-selection flags as implementation/workflow controls.

Consumer migrations are integration work in their owning repositories; this change defines and implements the reusable platform and the semantic contract they consume.

## Capabilities

### New Capabilities

- `charmcli-core`: Cobra/Fang application execution, streams/runtime, exit/error semantics, output contracts, terminal behavior and Charm-native presentation primitives.
- `charmcli-conventions`: shared command grammar, flags, resource naming, human visual grammar, apply/doctor semantics, dry-run and compatibility rules.
- `charmcli-interaction`: Huh-based prompts, exact destructive confirmation, non-interactive behavior and Bubble Tea/Bubbles progress primitives.
- `charmcli-k8s`: optional Kubernetes CLI integration using upstream kubectl/client-go semantics and kubectl-plugin invocation identity.
- `charmcli-testkit`: deterministic IO/capability fixtures for unit and visual contract testing.

## Non-goals

- Replacing Cobra with a charmcli-specific command tree DSL.
- Wrapping Bubble Tea into another full-screen TUI framework.
- Reimplementing kubectl config loading, auth, REST discovery, controller-runtime clients, or Kubernetes domain semantics.
- Moving homelab, PolicyOps, or Resource Controller business/domain code into this repository.
- Reproducing the current Clack-inspired `homelabctl` visual language. Its interaction quality and safety are requirements; its visual implementation is not.
- Erasing meaningful domain states merely to force a universal status enum. Shared severity/presentation is standardized; domain meaning remains explicit.
- Providing a stable v1 API immediately. The initial release is intentionally pre-1.0 while the three reference consumers validate the abstractions.

## Impact

- `charmcli` becomes the reusable Go counterpart to the existing Python-oriented `sphr2k/cli` project.
- Core consumers do not acquire Kubernetes dependencies unless they import the `/k8s` module.
- Existing consumer-specific hand-written parsers, TTY detection, style palettes, progress loops and kubectl plumbing can be removed as each CLI migrates.
- Existing command surfaces intentionally change where needed to achieve one predictable resource-oriented language.
- The reference migration must preserve or improve interactive safety and scriptability, especially `homelabctl` exact replacement confirmation and clean structured output.
