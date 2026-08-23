## Why

The homelab toolchain currently has three Go CLIs with overlapping concerns but different command, rendering, interaction, error, output, and Kubernetes plumbing:

- `homelabctl` uses a bespoke parser plus hand-built Clack-inspired Lip Gloss interaction and progress rendering;
- `policyopsctl` uses Cobra + Fang and Kubernetes `genericclioptions`, but owns its own output styling, invocation-name handling, namespace completion, and exit semantics;
- `kubectl-resources` uses a hand-written parser and direct clientcmd loading and is being renamed to the standalone product name `resourcectl`, while `kubectl resources` remains a supported plugin alias.

The duplicated code is not primarily domain logic. It is CLI product infrastructure: execution, IO contracts, machine output, error/exit semantics, prompts, destructive confirmation, progress, terminal capability handling, and Kubernetes CLI conventions.

The desired result is one coherent, Charm-native CLI experience without forcing unrelated domain code into one monolith.

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

Standardize cross-CLI behavioral contracts:

- stdout contains the requested result only;
- stderr contains prompts, progress, diagnostics and rendered errors;
- structured output is deterministic and ANSI-free;
- interactive features degrade safely in non-TTY/CI environments;
- exact destructive confirmation is available so `homelabctl` does not regress from its current typed-name guard;
- user cancellation is a first-class exit condition;
- commands may return a silent non-zero status for diagnostic commands such as `doctor` without an additional Fang error banner.

## Initial Consumer Direction

The framework is designed to support these migrations without embedding their domain logic:

```text
homelabctl
policyopsctl
resourcectl
kubectl policyops   -> policyopsctl command tree
kubectl resources   -> resourcectl command tree
```

`resourcectl` becomes the canonical standalone name. Its canonical resource-list verb becomes `get`; the old `list` spelling may be retained as a compatibility alias during migration. Invoking the root with no command should show normal Fang help rather than implicitly listing resources.

Consumer migrations are integration work in their owning repositories; this change defines and implements the reusable platform they consume.

## Capabilities

### New Capabilities

- `charmcli-core`: Cobra/Fang application execution, streams/runtime, exit/error semantics, output contracts, terminal behavior and Charm-native presentation primitives.
- `charmcli-interaction`: Huh-based prompts, exact destructive confirmation, non-interactive behavior and Bubble Tea/Bubbles progress primitives.
- `charmcli-k8s`: optional Kubernetes CLI integration using upstream kubectl/client-go semantics and kubectl-plugin invocation identity.
- `charmcli-testkit`: deterministic IO/capability fixtures for unit and visual contract testing.

## Non-goals

- Replacing Cobra with a charmcli-specific command tree DSL.
- Wrapping Bubble Tea into another full-screen TUI framework.
- Reimplementing kubectl config loading, auth, REST discovery, controller-runtime clients, or Kubernetes domain semantics.
- Moving homelab, PolicyOps, or Resource Controller business/domain code into this repository.
- Reproducing the current Clack-inspired `homelabctl` visual language. Its interaction quality and safety are requirements; its visual implementation is not.
- Providing a stable v1 API immediately. The initial release is intentionally pre-1.0 while the three reference consumers validate the abstractions.

## Impact

- `charmcli` becomes the reusable Go counterpart to the existing Python-oriented `sphr2k/cli` project.
- Core consumers do not acquire Kubernetes dependencies unless they import the `/k8s` module.
- Existing consumer-specific hand-written parsers, TTY detection, style palettes, progress loops and kubectl plumbing can be removed as each CLI migrates.
- The reference migration must preserve or improve interactive safety and scriptability, especially `homelabctl` exact replacement confirmation and clean JSON output.
